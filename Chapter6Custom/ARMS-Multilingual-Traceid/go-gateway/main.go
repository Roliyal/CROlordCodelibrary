package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envFirst(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing env %s", key)
	}
	return v
}

func loadEnvFromProjectRoot() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("runtime.Caller failed")
	}
	gwDir := filepath.Dir(thisFile)
	rootDir := filepath.Dir(gwDir)
	rootEnv := filepath.Join(rootDir, ".env")
	localEnv := filepath.Join(gwDir, ".env")

	log.Printf("[go] env candidates:\n  local=%s\n  root=%s", localEnv, rootEnv)

	// 优先根目录 .env
	_ = godotenv.Overload(rootEnv, localEnv)
	log.Printf("[go] loaded: OTEL=%q", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	log.Printf("[go] loaded: PY_URL=%q PY_BASE_URL=%q", os.Getenv("PY_URL"), os.Getenv("PY_BASE_URL"))
}

func initOTel(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	otlpURL := mustEnv("OTEL_EXPORTER_OTLP_ENDPOINT")

	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(otlpURL),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithHeaders(map[string]string{}),
		otlptracehttp.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}

	res, err := sdkresource.New(ctx, sdkresource.WithAttributes(semconv.ServiceName(serviceName)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

func logWithSpan(prefix string, ctx context.Context, extra string) {
	sc := trace.SpanContextFromContext(ctx)
	tid := sc.TraceID().String()
	sid := sc.SpanID().String()
	log.Printf("%s trace_id=%s span_id=%s %s", prefix, tid, sid, extra)
}

func main() {
	loadEnvFromProjectRoot()

	pyBase := envFirst("PY_BASE_URL", "PY_URL")
	if pyBase == "" {
		log.Fatalf("missing env PY_BASE_URL or PY_URL")
	}

	ctx := context.Background()
	shutdown, err := initOTel(ctx, "go-gateway")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	goPort := envOr("GO_PORT", "8080")

	// ✅ client 的 Transport 用 otelhttp，保证 client span + 自动注入不会丢
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   5 * time.Second,
	}

	mux := http.NewServeMux()

	// ✅ 入口用 otelhttp.NewHandler 创建 SERVER span
	mux.Handle("/api/hello", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 这个 ctx 已经包含 SERVER span
		ctx := r.Context()

		logWithSpan("[go] /api/hello", ctx, fmt.Sprintf("traceparent_in=%s", r.Header.Get("traceparent")))

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/py/work", pyBase), nil)

		// ✅ 双保险：显式注入 tracecontext（traceparent）
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "call python failed: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()

		var pyData any
		_ = json.NewDecoder(resp.Body).Decode(&pyData)

		sc := trace.SpanContextFromContext(ctx)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message":         "hello from go-gateway",
			"time":            time.Now().Format(time.RFC3339),
			"trace_id":        sc.TraceID().String(),
			"span_id":         sc.SpanID().String(),
			"traceparent_in":  r.Header.Get("traceparent"),
			"python_response": pyData,
		})
	}), "api/hello"))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":" + goPort
	log.Println("[go] listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
