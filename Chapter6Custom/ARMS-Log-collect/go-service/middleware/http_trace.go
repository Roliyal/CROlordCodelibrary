package middleware

import (
	"net/http"

	"armslogcollect/go-service/trace"
)

func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := trace.ExtractOrCreate(r.Header.Get("X-Trace-Id"), r.Header.Get("traceparent"))
		w.Header().Set("X-Trace-Id", traceID)
		ctx := WithTraceID(r.Context(), traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
