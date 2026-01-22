package httpx

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"armslogcollect/go-service/grpc"
	"armslogcollect/go-service/logger"
	"armslogcollect/go-service/middleware"
)

type Handlers struct {
	Log      *logger.Logger
	JavaHTTP *JavaHTTPClient
	JavaGRPC *grpcx.JavaGRPCClient
}

func nFromQuery(r *http.Request, def int) int {
	n := def
	if v := r.URL.Query().Get("n"); v != "" {
		if x, err := strconv.Atoi(v); err == nil && x > 0 && x <= 2000 {
			n = x
		}
	}
	return n
}

func (h *Handlers) Pay(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.TraceIDFrom(r.Context())
	start := time.Now()
	n := nFromQuery(r, 10)

	h.Log.Info("source", "PayHandler", "category", "payment.pay", "traceId", traceID,
		"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
		"message", "pay request received",
	)

	for i := 0; i < n; i++ {
		h.Log.Info("source", "PayHandler", "category", "payment.pay.step", "traceId", traceID,
			"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
			"message", "step processing",
		)
	}

	// HTTP outbound to Java
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := h.JavaHTTP.Do(ctx, http.MethodGet, "/api/user/get?n=5", traceID)
	if err != nil {
		h.Log.Error("source", "PayHandler", "category", "remote.http.error", "traceId", traceID,
			"protocol", "http", "direction", "outbound", "method", "GET", "path", "/api/user/get",
			"remoteService", "java-service",
			"errorType", "http", "errorMessage", err.Error(),
			"message", "java http call failed",
		)
		http.Error(w, "bad gateway", 502)
		return
	}
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	// gRPC outbound to Java
	if _, err := h.JavaGRPC.ValidateUser(r.Context(), traceID); err != nil {
		h.Log.Error("source", "PayHandler", "category", "remote.grpc.error", "traceId", traceID,
			"protocol", "grpc", "direction", "outbound", "method", "ValidateUser", "path", "ValidateUser",
			"remoteService", "java-service",
			"errorType", "grpc", "errorMessage", err.Error(),
			"message", "java grpc call failed",
		)
	}

	h.Log.Info("source", "PayHandler", "category", "payment.pay.done", "traceId", traceID,
		"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
		"costMs", time.Since(start).Milliseconds(),
		"message", "pay done",
	)
	_, _ = w.Write([]byte("OK"))
}

func (h *Handlers) Refund(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.TraceIDFrom(r.Context())
	start := time.Now()
	n := nFromQuery(r, 10)

	h.Log.Warn("source", "RefundHandler", "category", "payment.refund", "traceId", traceID,
		"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
		"message", "refund request received",
	)

	for i := 0; i < n; i++ {
		h.Log.Info("source", "RefundHandler", "category", "payment.refund.step", "traceId", traceID,
			"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
			"message", "refund step",
		)
	}

	// outbound HTTP + gRPC to Java
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	_, _ = h.JavaHTTP.Do(ctx, http.MethodPost, "/api/inventory/reserve?n=5", traceID)
	_, _ = h.JavaGRPC.ReserveInventory(r.Context(), traceID)

	h.Log.Info("source", "RefundHandler", "category", "payment.refund.done", "traceId", traceID,
		"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
		"costMs", time.Since(start).Milliseconds(),
		"message", "refund done",
	)
	_, _ = w.Write([]byte("OK"))
}

func (h *Handlers) Query(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.TraceIDFrom(r.Context())
	start := time.Now()
	n := nFromQuery(r, 10)

	h.Log.Info("source", "QueryHandler", "category", "payment.query", "traceId", traceID,
		"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
		"message", "query request received",
	)

	for i := 0; i < n; i++ {
		h.Log.Info("source", "QueryHandler", "category", "payment.query.step", "traceId", traceID,
			"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
			"message", "query step",
		)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	_, _ = h.JavaHTTP.Do(ctx, http.MethodPost, "/api/order/create?n=5", traceID)
	_, _ = h.JavaGRPC.AuditOrder(r.Context(), traceID)

	h.Log.Info("source", "QueryHandler", "category", "payment.query.done", "traceId", traceID,
		"protocol", "http", "direction", "inbound", "method", r.Method, "path", r.URL.Path,
		"costMs", time.Since(start).Milliseconds(),
		"message", "query done",
	)
	_, _ = w.Write([]byte("OK"))
}
