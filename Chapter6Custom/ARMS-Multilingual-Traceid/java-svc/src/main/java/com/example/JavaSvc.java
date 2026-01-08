package com.example;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.trace.*;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.TextMapGetter;
import io.opentelemetry.context.propagation.TextMapPropagator;
import io.opentelemetry.exporter.otlp.http.trace.OtlpHttpSpanExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;

import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.time.Instant;
import java.util.*;

public class JavaSvc {
    static final ObjectMapper M = new ObjectMapper();

    static Map<String, String> loadRootEnv() {
        try {
            Path root = Paths.get("").toAbsolutePath().getParent(); // java-svc 的上一级
            List<String> lines = Files.readAllLines(root.resolve(".env"), StandardCharsets.UTF_8);
            Map<String, String> env = new HashMap<>();
            for (String line : lines) {
                String s = line.trim();
                if (s.isEmpty() || s.startsWith("#") || !s.contains("=")) continue;
                String[] kv = s.split("=", 2);
                env.put(kv[0].trim(), kv[1].trim());
            }
            return env;
        } catch (Exception e) {
            return Map.of();
        }
    }

    static OpenTelemetry initOtel(Map<String, String> env) {
        String endpoint = env.getOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "");
        // headers 允许为空（HTTP adapt_URL 模式不需要）
        String headers = env.getOrDefault("OTEL_EXPORTER_OTLP_HEADERS", "");

        if (endpoint.isEmpty()) {
            throw new RuntimeException("missing OTEL_EXPORTER_OTLP_ENDPOINT in .env");
        }

        OtlpHttpSpanExporter.Builder exporterBuilder = OtlpHttpSpanExporter.builder()
                .setEndpoint(endpoint);

        // 如果你未来切到需要 header 的模式，这里也支持
        Map<String, String> headerMap = parseHeaders(headers);
        if (headerMap.containsKey("Authentication") && !headerMap.get("Authentication").isEmpty()) {
            exporterBuilder.addHeader("Authentication", headerMap.get("Authentication"));
        }

        OtlpHttpSpanExporter exporter = exporterBuilder.build();

        Resource resource = Resource.getDefault().merge(
                Resource.create(Attributes.builder().put("service.name", "java-svc").build())
        );

        SdkTracerProvider tp = SdkTracerProvider.builder()
                .setResource(resource)
                .addSpanProcessor(BatchSpanProcessor.builder(exporter).build())
                .build();

        OpenTelemetrySdk otel = OpenTelemetrySdk.builder()
                .setTracerProvider(tp)
                .build();

        GlobalOpenTelemetry.set(otel);
        return otel;
    }

    static Map<String, String> parseHeaders(String headers) {
        Map<String, String> out = new HashMap<>();
        if (headers == null) return out;
        for (String part : headers.split(",")) {
            String p = part.trim();
            if (p.isEmpty() || !p.contains("=")) continue;
            String[] kv = p.split("=", 2);
            out.put(kv[0].trim(), kv[1].trim());
        }
        return out;
    }

    static String traceIdFromContext(Context ctx) {
        Span span = Span.fromContext(ctx);
        SpanContext sc = span.getSpanContext();
        return sc.isValid() ? sc.getTraceId() : "";
    }

    public static void main(String[] args) throws Exception {
        Map<String, String> env = loadRootEnv();
        OpenTelemetry otel = initOtel(env);

        int port = Integer.parseInt(env.getOrDefault("JAVA_PORT", "8082"));
        String cppUrl = env.getOrDefault("CPP_URL", "http://127.0.0.1:8083");

        TextMapPropagator propagator = otel.getPropagators().getTextMapPropagator();
        Tracer tracer = otel.getTracer("java-svc");

        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", port), 0);

        server.createContext("/java/work", exchange -> {
            Context extracted = propagator.extract(Context.current(), exchange, new HttpExchangeGetter());
            Span serverSpan = tracer.spanBuilder("java /java/work")
                    .setSpanKind(SpanKind.SERVER)
                    .setParent(extracted)
                    .startSpan();

            try (Scope scope = serverSpan.makeCurrent()) {
                String tid = traceIdFromContext(Context.current());
                System.out.println("[java] /java/work trace_id=" + tid);

                HttpClient client = HttpClient.newHttpClient();
                HttpRequest.Builder reqB = HttpRequest.newBuilder()
                        .uri(URI.create(cppUrl + "/cpp/work"))
                        .GET();

                // 注入 traceparent 等 header 到下游
                propagator.inject(Context.current(), reqB, (builder, key, value) -> builder.header(key, value));

                Map<String, Object> cppResp;
                try {
                    HttpResponse<String> resp = client.send(reqB.build(), HttpResponse.BodyHandlers.ofString());
                    cppResp = M.readValue(resp.body(), Map.class);
                } catch (Exception e) {
                    cppResp = Map.of("error", e.toString());
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
            exchange.sendResponseHeaders(200, body.length);
            try (OutputStream os = exchange.getResponseBody()) {
                os.write(body);
            }
        });

        System.out.println("[java] listening on :" + port);
        server.start();
    }

    static class HttpExchangeGetter implements TextMapGetter<HttpExchange> {
        @Override
        public Iterable<String> keys(HttpExchange carrier) {
            return carrier.getRequestHeaders().keySet();
        }

        @Override
        public String get(HttpExchange carrier, String key) {
            List<String> v = carrier.getRequestHeaders().get(key);
            if (v == null || v.isEmpty()) return null;
            return v.get(0);
        }
    }
}
