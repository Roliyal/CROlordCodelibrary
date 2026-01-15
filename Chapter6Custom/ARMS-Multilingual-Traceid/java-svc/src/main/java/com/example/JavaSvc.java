package com.example;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.trace.*;
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.*;

import io.opentelemetry.exporter.otlp.http.trace.OtlpHttpSpanExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;

import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.net.URI;
import java.net.http.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.time.Instant;
import java.util.*;

public class JavaSvc {
    static final ObjectMapper M = new ObjectMapper();

    static boolean exists(String p) {
        try {
            return p != null && !p.isBlank()
                    && Files.exists(Path.of(p))
                    && !Files.isDirectory(Path.of(p));
        } catch (Exception e) {
            return false;
        }
    }

    static Map<String, String> loadEnv() {
        String dotenv = System.getenv().getOrDefault("DOTENV_PATH", "/app/.env");
        List<String> candidates = List.of(dotenv, ".env", "../.env");

        System.out.println("[java] env candidates:");
        for (String p : candidates) {
            System.out.println("  - " + p + " (exists=" + exists(p) + ")");
        }

        Map<String, String> env = new HashMap<>(System.getenv());

        for (String p : candidates) {
            if (!exists(p)) continue;
            try {
                List<String> lines = Files.readAllLines(Path.of(p), StandardCharsets.UTF_8);
                for (String line : lines) {
                    String s = line.trim();
                    if (s.isEmpty() || s.startsWith("#") || !s.contains("=")) continue;
                    String[] kv = s.split("=", 2);
                    env.put(kv[0].trim(), kv[1].trim());
                }
                System.out.println("[java] loaded env file: " + p);
                break;
            } catch (Exception e) {
                System.out.println("[java] failed to read " + p + ": " + e);
            }
        }

        System.out.println("[java] effective OTEL_EXPORTER_OTLP_ENDPOINT=" + env.getOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", ""));
        System.out.println("[java] effective JAVA_PORT=" + env.getOrDefault("JAVA_PORT", "8082"));
        System.out.println("[java] effective CPP_URL=" + env.getOrDefault("CPP_URL", ""));
        return env;
    }

    static OpenTelemetry initOtel(Map<String, String> env) {
        String endpoint = env.getOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "");
        if (endpoint.isEmpty()) throw new RuntimeException("missing OTEL_EXPORTER_OTLP_ENDPOINT");

        // ✅ OTLP over HTTP exporter（匹配你的 http://.../api/otlp/traces）
        OtlpHttpSpanExporter exporter = OtlpHttpSpanExporter.builder()
                .setEndpoint(endpoint)
                .build();

        Resource resource = Resource.getDefault().merge(
                Resource.create(Attributes.builder().put("service.name", "java-svc").build())
        );

        SdkTracerProvider tp = SdkTracerProvider.builder()
                .setResource(resource)
                .addSpanProcessor(BatchSpanProcessor.builder(exporter).build())
                .build();

        OpenTelemetrySdk otel = OpenTelemetrySdk.builder()
                .setTracerProvider(tp)
                .setPropagators(ContextPropagators.create(W3CTraceContextPropagator.getInstance()))
                .build();

        GlobalOpenTelemetry.set(otel);
        return otel;
    }

    static String traceIdFromContext(Context ctx) {
        Span span = Span.fromContext(ctx);
        SpanContext sc = span.getSpanContext();
        return sc.isValid() ? sc.getTraceId() : "";
    }

    public static void main(String[] args) throws Exception {
        Map<String, String> env = loadEnv();
        OpenTelemetry otel = initOtel(env);

        int port = Integer.parseInt(env.getOrDefault("JAVA_PORT", "8082"));
        String cppUrl = env.getOrDefault("CPP_URL", "");

        TextMapPropagator propagator = otel.getPropagators().getTextMapPropagator();
        Tracer tracer = otel.getTracer("java-svc");

        HttpClient client = HttpClient.newHttpClient();

        // ✅ K8s 必须 0.0.0.0
        HttpServer server = HttpServer.create(new InetSocketAddress("0.0.0.0", port), 0);

        server.createContext("/java/work", exchange -> {
            System.out.println("[java] headers keys=" + exchange.getRequestHeaders().keySet());
            System.out.println("[java] traceparent_in(raw)=" + firstHeaderCI(exchange, "traceparent"));

            Context extracted = propagator.extract(Context.current(), exchange, new HttpExchangeGetter());

            Span serverSpan = tracer.spanBuilder("java /java/work")
                    .setSpanKind(SpanKind.SERVER)
                    .setParent(extracted)
                    .startSpan();

            try (Scope scope = serverSpan.makeCurrent()) {
                String tid = traceIdFromContext(Context.current());
                System.out.println("[java] /java/work trace_id=" + tid);

                Map<String, Object> cppResp = Map.of("warn", "CPP_URL is empty");
                if (cppUrl != null && !cppUrl.isBlank()) {
                    HttpRequest.Builder reqB = HttpRequest.newBuilder()
                            .uri(URI.create(cppUrl + "/cpp/work"))
                            .GET();

                    // ✅ 注入 traceparent -> C++
                    propagator.inject(Context.current(), reqB, (builder, key, value) -> builder.header(key, value));

                    try {
                        HttpResponse<String> resp = client.send(reqB.build(), HttpResponse.BodyHandlers.ofString());
                        cppResp = M.readValue(resp.body(), Map.class);
                    } catch (Exception e) {
                        cppResp = Map.of("error", e.toString());
                    }
                }

                Map<String, Object> out = new LinkedHashMap<>();
                out.put("message", "hello from java");
                out.put("time", Instant.now().toString());
                out.put("trace_id", tid);
                out.put("cpp", cppResp);

                byte[] body = M.writeValueAsBytes(out);
                exchange.getResponseHeaders().add("Content-Type", "application/json");
                exchange.sendResponseHeaders(200, body.length);
                try (OutputStream os = exchange.getResponseBody()) {
                    os.write(body);
                }
            } finally {
                serverSpan.end();
            }
        });

        server.createContext("/healthz", exchange -> {
            byte[] body = "ok".getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().add("Content-Type", "text/plain");
            exchange.sendResponseHeaders(200, body.length);
            try (OutputStream os = exchange.getResponseBody()) {
                os.write(body);
            }
        });

        System.out.println("[java] listening on :" + port);
        server.start();
    }

    static String firstHeaderCI(HttpExchange ex, String name) {
        String v = ex.getRequestHeaders().getFirst(name);
        if (v != null) return v;
        for (String k : ex.getRequestHeaders().keySet()) {
            if (k != null && k.equalsIgnoreCase(name)) {
                return ex.getRequestHeaders().getFirst(k);
            }
        }
        return null;
    }

    static class HttpExchangeGetter implements TextMapGetter<HttpExchange> {
        @Override
        public Iterable<String> keys(HttpExchange carrier) {
            return carrier.getRequestHeaders().keySet();
        }

        @Override
        public String get(HttpExchange carrier, String key) {
            if (key == null) return null;

            List<String> v = carrier.getRequestHeaders().get(key);
            if (v != null && !v.isEmpty()) return v.get(0);

            for (String k : carrier.getRequestHeaders().keySet()) {
                if (k != null && k.equalsIgnoreCase(key)) {
                    List<String> vv = carrier.getRequestHeaders().get(k);
                    if (vv != null && !vv.isEmpty()) return vv.get(0);
                }
            }
            return null;
        }
    }
}
