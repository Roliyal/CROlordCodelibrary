package grpcx

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	"armslogcollect/go-service/logger"
	"armslogcollect/go-service/middleware"
	"armslogcollect/go-service/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func traceIDFromIncomingMD(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		x := ""
		if v := md.Get("x-trace-id"); len(v) > 0 {
			x = v[0]
		}
		tp := ""
		if v := md.Get("traceparent"); len(v) > 0 {
			tp = v[0]
		}
		return trace.ExtractOrCreate(x, tp)
	}
	return trace.ExtractOrCreate("", "")
}

func UnaryServerInterceptor(l *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		traceID := traceIDFromIncomingMD(ctx)
		ctx = middleware.WithTraceID(ctx, traceID)

		l.Info(
			"source", "GoGrpcServer",
			"category", "grpc.inbound.start",
			"traceId", traceID,
			"protocol", "grpc",
			"direction", "inbound",
			"method", info.FullMethod,
			"path", info.FullMethod,
			"message", "grpc request start",
		)

		defer func() {
			if rec := recover(); rec != nil {
				stack := strings.ReplaceAll(string(debug.Stack()), "\n", "\\n")
				l.Error(
					"source", "GoGrpcServer",
					"category", "grpc.panic",
					"traceId", traceID,
					"protocol", "grpc",
					"direction", "inbound",
					"method", info.FullMethod,
					"path", info.FullMethod,
					"status", "ERR",
					"errorType", "panic",
					"errorMessage", rec,
					"errorStack", stack,
					"message", "panic recovered",
				)
				// IMPORTANT: never re-panic; return a gRPC error instead.
				err = status.Error(codes.Internal, "internal error")
				resp = nil
			}

			l.Info(
				"source", "GoGrpcServer",
				"category", "grpc.inbound.done",
				"traceId", traceID,
				"protocol", "grpc",
				"direction", "inbound",
				"method", info.FullMethod,
				"path", info.FullMethod,
				"status", statusFromErr(err),
				"costMs", time.Since(start).Milliseconds(),
				"message", "grpc request done",
			)
		}()

		return handler(ctx, req)
	}
}

func UnaryClientInterceptor(l *logger.Logger, remoteService string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		traceID := middleware.TraceIDFrom(ctx)
		if traceID == "unknown" {
			traceID = trace.ExtractOrCreate("", "")
			ctx = middleware.WithTraceID(ctx, traceID)
		}

		md := metadata.Pairs(
			"x-trace-id", traceID,
			"traceparent", "00-"+traceID+"-0000000000000000-01",
		)
		ctx = metadata.NewOutgoingContext(ctx, md)

		l.Info(
			"source", "GoGrpcClient",
			"category", "grpc.outbound.start",
			"traceId", traceID,
			"protocol", "grpc",
			"direction", "outbound",
			"method", method,
			"path", method,
			"remoteService", remoteService,
			"message", "grpc outbound start",
		)

		err := invoker(ctx, method, req, reply, cc, opts...)

		l.Info(
			"source", "GoGrpcClient",
			"category", "grpc.outbound.done",
			"traceId", traceID,
			"protocol", "grpc",
			"direction", "outbound",
			"method", method,
			"path", method,
			"remoteService", remoteService,
			"status", statusFromErr(err),
			"costMs", time.Since(start).Milliseconds(),
			"message", "grpc outbound done",
		)

		if err != nil {
			l.Error(
				"source", "GoGrpcClient",
				"category", "grpc.outbound.error",
				"traceId", traceID,
				"protocol", "grpc",
				"direction", "outbound",
				"method", method,
				"path", method,
				"remoteService", remoteService,
				"errorType", "grpc",
				"errorMessage", err.Error(),
				"message", "grpc outbound error",
			)
		}
		return err
	}
}

func statusFromErr(err error) any {
	if err == nil {
		return "OK"
	}
	return "ERR"
}
