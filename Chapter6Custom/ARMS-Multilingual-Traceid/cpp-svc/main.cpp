#include <httplib.h>
#include <nlohmann/json.hpp>

#include <chrono>
#include <cstdlib>
#include <fstream>
#include <iostream>
#include <map>
#include <sstream>
#include <string>
#include <utility>
#include <vector>

#include <opentelemetry/context/context.h>
#include <opentelemetry/context/propagation/global_propagator.h>
#include <opentelemetry/context/propagation/text_map_propagator.h>

#include <opentelemetry/trace/provider.h>
#include <opentelemetry/trace/scope.h>
#include <opentelemetry/trace/span.h>
#include <opentelemetry/trace/span_kind.h>
#include <opentelemetry/trace/span_startoptions.h>

#include <opentelemetry/sdk/resource/resource.h>
#include <opentelemetry/sdk/trace/batch_span_processor.h>
#include <opentelemetry/sdk/trace/batch_span_processor_options.h>
#include <opentelemetry/sdk/trace/tracer_provider.h>

#include <opentelemetry/exporters/otlp/otlp_http_exporter.h>
#include <opentelemetry/exporters/otlp/otlp_http_exporter_options.h>

// W3C tracecontext propagator
#include <opentelemetry/trace/propagation/http_trace_context.h>

namespace trace = opentelemetry::trace;
namespace nostd = opentelemetry::nostd;
namespace ctx = opentelemetry::context;
namespace propagation = opentelemetry::context::propagation;
namespace sdktrace = opentelemetry::sdk::trace;
namespace resource = opentelemetry::sdk::resource;

// -------------------- util: read .env --------------------
static bool LoadEnvFile(const std::string &path, std::map<std::string, std::string> &env)
{
  if (path.empty())
    return false;

  std::ifstream in(path);
  if (!in.is_open())
    return false;

  std::string line;
  while (std::getline(in, line))
  {
    if (line.empty() || line[0] == '#')
      continue;

    auto pos = line.find('=');
    if (pos == std::string::npos)
      continue;

    std::string k = line.substr(0, pos);
    std::string v = line.substr(pos + 1);

    while (!k.empty() && (k.back() == ' ' || k.back() == '\t' || k.back() == '\r'))
      k.pop_back();
    while (!v.empty() && (v.front() == ' ' || v.front() == '\t'))
      v.erase(v.begin());
    while (!v.empty() && (v.back() == '\r' || v.back() == '\n'))
      v.pop_back();

    env[k] = v;
  }
  return true;
}

static std::map<std::string, std::string> LoadEnv()
{
  std::map<std::string, std::string> env;

  const char *dotenv = std::getenv("DOTENV_PATH");
  const std::vector<std::string> candidates = {
      dotenv ? std::string(dotenv) : std::string(),
      "/app/.env",
      "./.env",
      "../.env",
      "../../.env",
  };

  std::cout << "[cpp] env candidates:\n";
  for (const auto &p : candidates)
  {
    if (p.empty())
      continue;
    std::cout << "  - " << p << "\n";
  }

  for (const auto &p : candidates)
  {
    if (LoadEnvFile(p, env))
    {
      std::cout << "[cpp] loaded env from " << p << std::endl;
      return env;
    }
  }

  std::cerr << "[cpp] WARN: cannot open .env (tried DOTENV_PATH,/app/.env,./.env,../.env,../../.env)\n";
  return env;
}

static std::string GetEnv(const std::map<std::string, std::string> &env, const std::string &k, const std::string &def = "")
{
  auto it = env.find(k);
  if (it == env.end() || it->second.empty())
    return def;
  return it->second;
}

// 解析 "k=v,k2=v2"
static std::vector<std::pair<std::string, std::string>> ParseHeadersPairs(const std::string &raw)
{
  std::vector<std::pair<std::string, std::string>> out;
  if (raw.empty())
    return out;

  std::stringstream ss(raw);
  std::string item;
  while (std::getline(ss, item, ','))
  {
    while (!item.empty() && (item.front() == ' ' || item.front() == '\t'))
      item.erase(item.begin());
    while (!item.empty() && (item.back() == ' ' || item.back() == '\t'))
      item.pop_back();
    if (item.empty())
      continue;

    auto eq = item.find('=');
    if (eq == std::string::npos)
      continue;

    std::string k = item.substr(0, eq);
    std::string v = item.substr(eq + 1);

    while (!k.empty() && (k.back() == ' ' || k.back() == '\t'))
      k.pop_back();
    while (!v.empty() && (v.front() == ' ' || v.front() == '\t'))
      v.erase(v.begin());

    if (!k.empty())
      out.emplace_back(std::move(k), std::move(v));
  }
  return out;
}

// -------------------- util: ids --------------------
static std::string HexLower(const uint8_t *data, size_t n)
{
  static const char *hex = "0123456789abcdef";
  std::string s;
  s.resize(n * 2);
  for (size_t i = 0; i < n; i++)
  {
    s[i * 2] = hex[(data[i] >> 4) & 0xF];
    s[i * 2 + 1] = hex[data[i] & 0xF];
  }
  return s;
}

static void LogWithSpan(const std::string &prefix, const trace::Span &span, const std::string &extra)
{
  auto sc = span.GetContext();
  auto tid = sc.trace_id().Id();
  auto sid = sc.span_id().Id();
  std::string trace_id = HexLower(tid.data(), tid.size());
  std::string span_id = HexLower(sid.data(), sid.size());
  std::cout << prefix << " trace_id=" << trace_id << " span_id=" << span_id;
  if (!extra.empty())
    std::cout << " " << extra;
  std::cout << std::endl;
}

// -------------------- carrier for extract (httplib) --------------------
class HttplibRequestCarrier final : public propagation::TextMapCarrier
{
public:
  explicit HttplibRequestCarrier(const httplib::Request &req) : req_(req) {}

  nostd::string_view Get(nostd::string_view key) const noexcept override
  {
    std::string k(key.data(), key.size());

    // 直接查
    auto it = req_.headers.find(k);
    if (it != req_.headers.end())
      return nostd::string_view(it->second.data(), it->second.size());

    // 大小写不敏感兜底
    for (const auto &kv : req_.headers)
    {
      if (kv.first.size() != k.size())
        continue;

      bool same = true;
      for (size_t i = 0; i < k.size(); i++)
      {
        char a = kv.first[i];
        char b = k[i];
        if ('A' <= a && a <= 'Z')
          a = char(a - 'A' + 'a');
        if ('A' <= b && b <= 'Z')
          b = char(b - 'A' + 'a');
        if (a != b)
        {
          same = false;
          break;
        }
      }
      if (same)
        return nostd::string_view(kv.second.data(), kv.second.size());
    }

    return nostd::string_view{};
  }

  void Set(nostd::string_view, nostd::string_view) noexcept override {}

private:
  const httplib::Request &req_;
};

// -------------------- init tracer + propagator --------------------
static void InitTracer(const std::map<std::string, std::string> &env)
{
  const std::string endpoint = GetEnv(env, "OTEL_EXPORTER_OTLP_ENDPOINT");
  if (endpoint.empty())
    throw std::runtime_error("missing OTEL_EXPORTER_OTLP_ENDPOINT in .env");

  opentelemetry::exporter::otlp::OtlpHttpExporterOptions opts;
  opts.url = endpoint;

  // OTEL_EXPORTER_OTLP_HEADERS="k=v,k2=v2"
  opts.http_headers.clear();
  const std::string rawHeaders = GetEnv(env, "OTEL_EXPORTER_OTLP_HEADERS");
  for (const auto &kv : ParseHeadersPairs(rawHeaders))
  {
    opts.http_headers.insert(kv);
  }

  auto exporter = std::unique_ptr<sdktrace::SpanExporter>(
      new opentelemetry::exporter::otlp::OtlpHttpExporter(opts));

  sdktrace::BatchSpanProcessorOptions bsp;
  bsp.schedule_delay_millis = std::chrono::milliseconds(200);
  bsp.export_timeout = std::chrono::milliseconds(10000);

  auto processor = std::unique_ptr<sdktrace::SpanProcessor>(
      new sdktrace::BatchSpanProcessor(std::move(exporter), bsp));

  auto res = resource::Resource::Create({
      {"service.name", GetEnv(env, "OTEL_SERVICE_NAME", "cpp-svc")},
  });

  auto provider = nostd::shared_ptr<trace::TracerProvider>(
      new sdktrace::TracerProvider(std::move(processor), res));

  trace::Provider::SetTracerProvider(provider);

  // ✅ W3C tracecontext
  propagation::GlobalTextMapPropagator::SetGlobalPropagator(
      nostd::shared_ptr<propagation::TextMapPropagator>(
          new opentelemetry::trace::propagation::HttpTraceContext()));

  std::cout << "[cpp] OTEL_EXPORTER_OTLP_ENDPOINT=" << endpoint << "\n";
  std::cout << "[cpp] OTEL_EXPORTER_OTLP_HEADERS=" << (rawHeaders.empty() ? "(empty)" : "(set)") << "\n";
  std::cout << "[cpp] OTEL_SERVICE_NAME=" << GetEnv(env, "OTEL_SERVICE_NAME", "cpp-svc") << "\n";
}

static std::string CurTraceId(const trace::Span &span)
{
  auto tid = span.GetContext().trace_id().Id();
  return HexLower(tid.data(), tid.size());
}

static std::string CurSpanId(const trace::Span &span)
{
  auto sid = span.GetContext().span_id().Id();
  return HexLower(sid.data(), sid.size());
}

int main()
{
  auto env = LoadEnv();
  InitTracer(env);

  int port = 8083;
  try
  {
    port = std::stoi(GetEnv(env, "CPP_PORT", "8083"));
  }
  catch (...)
  {
    port = 8083;
  }

  auto tracer = trace::Provider::GetTracerProvider()->GetTracer("cpp-svc");

  httplib::Server svr;

  svr.Get("/healthz", [&](const httplib::Request &, httplib::Response &res) {
    auto span = tracer->StartSpan("GET /healthz");
    trace::Scope scope(span);
    LogWithSpan("[cpp] /healthz", *span, "");
    res.set_content("ok", "text/plain");
    span->End();
  });

  svr.Get("/cpp/work", [&](const httplib::Request &req, httplib::Response &res) {
    // 1) extract 上游 context（traceparent）
    HttplibRequestCarrier carrier(req);
    auto propagator = propagation::GlobalTextMapPropagator::GetGlobalPropagator();

    ctx::Context parent = propagator->Extract(carrier, ctx::Context{});

    trace::StartSpanOptions opt;
    opt.kind = trace::SpanKind::kServer;
    opt.parent = parent;

    auto span = tracer->StartSpan("GET /cpp/work", {}, opt);
    trace::Scope scope(span);

    // 2) 打日志
    std::string tp_in;
    {
      auto tp_sv = carrier.Get("traceparent");
      tp_in = std::string(tp_sv.data(), tp_sv.size());
    }
    LogWithSpan("[cpp] /cpp/work", *span, "traceparent_in=" + tp_in);

    // 3) 返回 JSON
    nlohmann::json j;
    j["message"] = "hello from cpp";
    j["time"] = std::chrono::duration_cast<std::chrono::seconds>(
                    std::chrono::system_clock::now().time_since_epoch())
                    .count();
    j["traceparent_in"] = tp_in;
    j["trace_id"] = CurTraceId(*span);
    j["span_id"] = CurSpanId(*span);

    res.set_header("Content-Type", "application/json");
    res.set_content(j.dump(2), "application/json");

    span->End();
  });

  std::cout << "[cpp] listening on :" << port << std::endl;

  // ✅ K8s/容器必须 0.0.0.0
  svr.listen("0.0.0.0", port);
  return 0;
}
