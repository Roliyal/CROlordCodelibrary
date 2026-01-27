package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"

	"armslogcollect/go-service/logger"
)

// Recover catches panics, logs a single-line JSON error record, and returns 500.
// NOTE: stack is flattened to one line by replacing newlines with the literal \n sequence.
func Recover(l *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				traceID := TraceIDFrom(r.Context())
				stack := strings.ReplaceAll(string(debug.Stack()), "\n", "\\n")

				l.Error(
					"source", "RecoverMiddleware",
					"category", "http.panic",
					"traceId", traceID,
					"protocol", "http",
					"direction", "inbound",
					"method", r.Method,
					"path", r.URL.Path,
					"status", 500,
					"errorType", "panic",
					"errorMessage", rec,
					"errorStack", stack,
					"message", "panic recovered",
				)

				http.Error(w, "internal error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
