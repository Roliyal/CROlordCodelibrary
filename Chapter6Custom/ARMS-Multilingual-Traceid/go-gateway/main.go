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

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// ✅ 关键：容器里固定从 /app/.env 读取（Dockerfile 会 COPY 进去）
func loadEnv() {
	// 1) 容器固定路径（优先）
	containerEnv := envOr("DOTENV_PATH", "/app/.env")

	// 2) 本地开发路径（fallback）
	_, thisFile, _, ok := runtime.Caller(0)
	var gwDir, rootDir string
	if ok {
		gwDir = filepath.Dir(thisFile)
		rootDir = filepath.Dir(gwDir)
	}

	localEnv := ""
	rootEnv := ""
	if gwDir != "" {
		localEnv = filepath.Join(gwDir, ".env")
	}
	if rootDir != "" {
		rootEnv = filepath.Join(rootDir, ".env")
	}

	candidates := []string{
		containerEnv, // /app/.env
		localEnv,     // go-gateway/.env（如果你放了）
		rootEnv,      // repo 根目录 .env（本地）
	}

	log.Printf("[go] env candidates:")
	for _, p := range candidates {
		if p == "" {
			continue
		}
		log.Printf("  - %s (exists=%v)", p, fileExists(p))
	}

	// Overload：文件里有值就覆盖（因为你就是要“打进镜像就生效”）
	for _, p := range candidates {
		if p != "" && fileExists(p) {
			if err := godotenv.Overload(p); err != nil {
				log.Printf("[go] failed to load %s: %v", p, err)
			} else {
				log.Printf("[go] loaded env file: %s", p)
			}
		}
	}

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
	loadEnv()

	pyBase := envFirst("PY_BASE_URL", "PY_URL")
	if pyBase == "" {
		log.Fatalf("missing env PY_BASE_URL or PY_URL (check /app/.env baked into image)")
	}

	ctx := context.Background()
	shutdown, err := initOTel(ctx, "go-gateway")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	goPort := envOr("GO_PORT", "8080")

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   5 * time.Second,
	}

	mux := http.NewServeMux()

	mux.Handle("/api/hello", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logWithSpan("[go] /api/hello", ctx, fmt.Sprintf("traceparent_in=%s", r.Header.Get("traceparent")))

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/py/work", pyBase), nil)

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
