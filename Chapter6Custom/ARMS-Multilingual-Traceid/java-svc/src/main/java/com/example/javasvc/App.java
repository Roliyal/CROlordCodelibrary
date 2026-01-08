package com.example.javasvc;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import spark.Spark;

import java.io.IOException;
import java.time.OffsetDateTime;
import java.util.Collections;
import java.util.concurrent.TimeUnit;

import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.*;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.*;

import io.opentelemetry.sdk.autoconfigure.AutoConfiguredOpenTelemetrySdk;

public class App {

  static String getEnv(String k, String def) {
    String v = System.getenv(k);
    return (v == null || v.trim().isEmpty()) ? def : v.trim();
  }

  static boolean isEmpty(String s) {
    return s == null || s.trim().isEmpty();
  }

  static String jsonEscape(String s) {
    if (s == null) return "";
    return s.replace("\\", "\\\\")
        .replace("\"", "\\\"")
        .replace("\n", "\\n")
        .replace("\r", "\\r")
        .replace("\t", "\\t");
  }

  static void logWithSpan(String prefix, String extra) {
    Span scur = Span.current();
    SpanContext sctx = scur.getSpanContext();
    String tid = sctx.isValid() ? sctx.getTraceId() : "";
    String sid = sctx.isValid() ? sctx.getSpanId() : "";
    System.out.println(prefix + " trace_id=" + tid + " span_id=" + sid + (extra == null || extra.isEmpty() ? "" : " " + extra));
  }

  // Spark 请求头取值（extract 用）
  static final TextMapGetter<spark.Request> SPARK_GETTER = new TextMapGetter<spark.Request>() {
    @Override
    public Iterable<String> keys(spark.Request carrier) {
      return Collections.emptyList();
    }

    @Override
    public String get(spark.Request carrier, String key) {
      return carrier.headers(key);
    }
  };

  // OkHttp 注入 traceparent（inject 用）
  static final TextMapSetter<Request.Builder> OKHTTP_SETTER = new TextMapSetter<Request.Builder>() {
    @Override
    public void set(Request.Builder carrier, String key, String value) {
      if (carrier != null && key != null && value != null) {
        carrier.header(key, value);
      }
    }
  };

  static OpenTelemetry initOtelOrDie() {
    // 你使用 ARMS adapt URL（已经是 /api/otlp/traces），因此这里建议用 TRACES_ENDPOINT
    String tracesEndpoint = System.getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT");
    if (isEmpty(tracesEndpoint)) {
      // 如果你不用 TRACES_ENDPOINT，至少 ENDPOINT 要有
      String endpoint = System.getenv("OTEL_EXPORTER_OTLP_ENDPOINT");
      if (isEmpty(endpoint)) {
        throw new RuntimeException("missing env OTEL_EXPORTER_OTLP_TRACES_ENDPOINT or OTEL_EXPORTER_OTLP_ENDPOINT");
      }
    }

    // initialize() 内部会 buildAndRegisterGlobal（只会注册一次）
    AutoConfiguredOpenTelemetrySdk.initialize();
    OpenTelemetry otel = GlobalOpenTelemetry.get();

    System.out.println("[java] OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=" + getEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "(empty)"));
    System.out.println("[java] OTEL_EXPORTER_OTLP_ENDPOINT=" + getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "(empty)"));
    System.out.println("[java] OTEL_EXPORTER_OTLP_HEADERS=" + (getEnv("OTEL_EXPORTER_OTLP_HEADERS", "").isEmpty() ? "(empty)" : "(set)"));
    System.out.println("[java] OTEL_EXPORTER_OTLP_PROTOCOL=" + getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "(default)"));
    System.out.println("[java] OTEL_SERVICE_NAME=" + getEnv("OTEL_SERVICE_NAME", "(default)"));
    System.out.println("[java] OTEL_METRICS_EXPORTER=" + getEnv("OTEL_METRICS_EXPORTER", "(default)"));
    System.out.println("[java] OTEL_LOGS_EXPORTER=" + getEnv("OTEL_LOGS_EXPORTER", "(default)"));
    System.out.println("[java] OTEL_EXPORTER_OTLP_TIMEOUT=" + getEnv("OTEL_EXPORTER_OTLP_TIMEOUT", "(default)"));

    return otel;
  }

  public static void main(String[] args) {
    String javaPort = getEnv("JAVA_PORT", "8082");
    String cppUrl = getEnv("CPP_URL", "");

    OpenTelemetry otel = initOtelOrDie();
    Tracer tracer = otel.getTracer("java-svc");
    TextMapPropagator propagator = otel.getPropagators().getTextMapPropagator();

    OkHttpClient http = new OkHttpClient.Builder()
        .callTimeout(5, TimeUnit.SECONDS)
        .build();

    Spark.port(Integer.parseInt(javaPort));
    Spark.get("/healthz", (req, res) -> {
      logWithSpan("[java] /healthz", "");
      return "ok";
    });

    Spark.get("/java/work", (req, res) -> {
      res.type("application/json");

      String traceparentIn = req.headers("traceparent");

      // ✅ 从上游 header extract parent
      Context parent = propagator.extract(Context.current(), req, SPARK_GETTER);

      Span serverSpan = tracer.spanBuilder("GET /java/work")
          .setParent(parent)
          .setSpanKind(SpanKind.SERVER)
          .startSpan();

      try (Scope scope = serverSpan.makeCurrent()) {
        String traceId = serverSpan.getSpanContext().getTraceId();
        String spanId = serverSpan.getSpanContext().getSpanId();

        String cppBodyJson = "null";
        String traceparentToCpp = "";

        if (!isEmpty(cppUrl)) {
          Span clientSpan = tracer.spanBuilder("HTTP GET /cpp/work")
              .setSpanKind(SpanKind.CLIENT)
              .startSpan();

          try (Scope cs = clientSpan.makeCurrent()) {
            Request.Builder b = new Request.Builder().url(cppUrl + "/cpp/work").get();

            // ✅ 在 client span scope 下，把当前 context 注入到下游
            propagator.inject(Context.current(), b, OKHTTP_SETTER);

            Request built = b.build();
            String tp = built.header("traceparent");
            traceparentToCpp = (tp == null) ? "" : tp;

            try (Response out = http.newCall(built).execute()) {
              String body = (out.body() != null) ? out.body().string() : "";
              cppBodyJson = (body == null || body.isEmpty()) ? "\"\"" : body;
              clientSpan.setStatus(StatusCode.OK);
            } catch (IOException e) {
              clientSpan.recordException(e);
              clientSpan.setStatus(StatusCode.ERROR);
              cppBodyJson = "{\"error\":\"" + jsonEscape(e.getMessage()) + "\"}";
            }
          } finally {
            clientSpan.end();
          }
        }

        logWithSpan("[java] /java/work", "traceparent_in=" + (traceparentIn == null ? "" : traceparentIn) +
            " traceparent_to_cpp=" + traceparentToCpp);

        return "{"
            + "\"message\":\"hello from java\","
            + "\"time\":\"" + OffsetDateTime.now() + "\","
            + "\"trace_id\":\"" + traceId + "\","
            + "\"span_id\":\"" + spanId + "\","
            + "\"traceparent_in\":\"" + jsonEscape(traceparentIn == null ? "" : traceparentIn) + "\","
            + "\"traceparent_to_cpp\":\"" + jsonEscape(traceparentToCpp) + "\","
            + "\"cpp\":" + cppBodyJson
            + "}";
      } finally {
        serverSpan.end();
      }
    });

    System.out.println("[java] listening on :" + javaPort);
  }
}
