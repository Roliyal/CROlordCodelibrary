package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"armslogcollect/go-service/grpc"
	"armslogcollect/go-service/http"
	"armslogcollect/go-service/logger"
	"armslogcollect/go-service/middleware"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func main() {
	// Config
	println("go-service started")

	env := getenv("APP_ENV", "dev")
	version := getenv("APP_VERSION", "1.0.0")

	logPath := getenv("GO_LOG_PATH", "logs/app.log")
	maxSize := getenvInt("GO_LOG_MAX_SIZE_MB", 50)
	maxBackups := getenvInt("GO_LOG_MAX_BACKUPS", 5)
	maxAge := getenvInt("GO_LOG_MAX_AGE_DAYS", 7)

	httpPort := getenv("GO_HTTP_PORT", "8081")
	grpcPort := getenv("GO_GRPC_PORT", "9091")

	javaHTTP := getenv("JAVA_HTTP_BASE_URL", "http://localhost:8080")
	javaGrpcAddr := getenv("JAVA_GRPC_ADDR", "localhost:9090")

	l := logger.New(logger.Config{
		Path:       logPath,
		MaxSizeMB:  maxSize,
		MaxBackups: maxBackups,
		MaxAgeDays: maxAge,
		Env:        env,
		Version:    version,
	})

	// gRPC client to Java
	javaGrpc, err := grpcx.NewJavaGRPCClient(javaGrpcAddr, l)
	if err != nil {
		l.Error("source", "main", "category", "startup.error", "errorType", "grpc", "errorMessage", err.Error(), "message", "init java grpc client failed")
		return
	}
	defer javaGrpc.Close()

	// HTTP server
	handlers := &httpx.Handlers{
		Log:      l,
		JavaHTTP: httpx.NewJavaHTTPClient(javaHTTP),
		JavaGRPC: javaGrpc,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/payment/pay", handlers.Pay)       // 1/3
	mux.HandleFunc("/api/payment/refund", handlers.Refund) // 2/3
	mux.HandleFunc("/api/payment/query", handlers.Query)   // 3/3

	var handler http.Handler = mux
	handler = middleware.Trace(handler)
	handler = middleware.Recover(l, handler)
	handler = middleware.AccessLog(l, handler)

	srv := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// gRPC server
	grpcServer, lis, err := grpcx.StartGRPC(l, grpcPort)
	if err != nil {
		l.Error("source", "main", "category", "startup.error", "errorType", "grpc", "errorMessage", err.Error(), "message", "grpc listen failed")
		return
	}

	// Start servers
	go func() {
		l.Info("source", "main", "category", "startup", "protocol", "grpc", "direction", "inbound", "method", "listen", "path", ":"+grpcPort, "message", "grpc server start")
		if err := grpcServer.Serve(lis); err != nil {
			l.Error("source", "main", "category", "runtime.error", "protocol", "grpc", "errorType", "grpc", "errorMessage", err.Error(), "message", "grpc serve failed")
		}
	}()

	go func() {
		l.Info("source", "main", "category", "startup", "protocol", "http", "direction", "inbound", "method", "listen", "path", ":"+httpPort, "message", "http server start")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("source", "main", "category", "runtime.error", "protocol", "http", "errorType", "http", "errorMessage", err.Error(), "message", "http serve failed")
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	l.Warn("source", "main", "category", "shutdown", "message", "shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if err := srv.Shutdown(ctx); err != nil {
		l.Error("source", "main", "category", "shutdown.error", "errorType", "http", "errorMessage", err.Error(), "message", "http shutdown failed")
	}

	l.Info("source", "main", "category", "shutdown.done", "message", "shutdown complete")
}
