package middleware

import (
	"net/http"
	"time"

	"armslogcollect/go-service/logger"
)

type respWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *respWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *respWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func AccessLog(l *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		traceID := TraceIDFrom(r.Context())
		rw := &respWriter{ResponseWriter: w}

		l.Info(
			"source", "HttpAccessLog",
			"category", "http.inbound.start",
			"traceId", traceID,
			"protocol", "http",
			"direction", "inbound",
			"method", r.Method,
			"path", r.URL.Path,
			"peer", r.RemoteAddr,
			"message", "http request start",
		)

		next.ServeHTTP(rw, r)

		l.Info(
			"source", "HttpAccessLog",
			"category", "http.inbound.done",
			"traceId", traceID,
			"protocol", "http",
			"direction", "inbound",
			"method", r.Method,
			"path", r.URL.Path,
			"peer", r.RemoteAddr,
			"status", rw.status,
			"bytes", rw.bytes,
			"costMs", time.Since(start).Milliseconds(),
			"message", "http request done",
		)
	})
}
