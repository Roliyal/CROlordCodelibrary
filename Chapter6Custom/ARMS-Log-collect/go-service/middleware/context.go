package middleware

import "context"

type ctxKey string

const (
	TraceIDKey ctxKey = "traceId"
)

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func TraceIDFrom(ctx context.Context) string {
	if v := ctx.Value(TraceIDKey); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "unknown"
}
