package grpcx

import (
	"context"
	"net"

	"armslogcollect/go-service/logger"
	"armslogcollect/go-service/middleware"
	"armslogcollect/go-service/pb"
	"google.golang.org/grpc"
)

type GoBridgeServer struct {
	pb.UnimplementedGoBridgeServer
	log *logger.Logger
}

func NewGoBridgeServer(l *logger.Logger) *GoBridgeServer {
	return &GoBridgeServer{log: l}
}

func (s *GoBridgeServer) ProcessPayment(ctx context.Context, r *pb.ActionRequest) (*pb.ActionReply, error) {
	traceID := middleware.TraceIDFrom(ctx)
	s.log.Info(
		"source", "GoBridgeServer",
		"category", "grpc.go.process_payment",
		"traceId", traceID,
		"protocol", "grpc",
		"direction", "inbound",
		"method", "ProcessPayment",
		"path", "ProcessPayment",
		"message", "process payment",
	)
	// business logs
	for i := 0; i < 5; i++ {
		s.log.Info("source", "GoBridgeServer", "category", "payment.step", "traceId", traceID,
			"protocol", "grpc", "direction", "inbound", "method", "ProcessPayment", "path", "ProcessPayment",
			"message", "payment step",
		)
	}
	return &pb.ActionReply{TraceId: r.TraceId, Code: 0, Result: "PAY_OK"}, nil
}

func (s *GoBridgeServer) IssueRefund(ctx context.Context, r *pb.ActionRequest) (*pb.ActionReply, error) {
	traceID := middleware.TraceIDFrom(ctx)
	s.log.Warn(
		"source", "GoBridgeServer",
		"category", "grpc.go.issue_refund",
		"traceId", traceID,
		"protocol", "grpc",
		"direction", "inbound",
		"method", "IssueRefund",
		"path", "IssueRefund",
		"message", "issue refund",
	)
	return &pb.ActionReply{TraceId: r.TraceId, Code: 0, Result: "REFUND_OK"}, nil
}

func (s *GoBridgeServer) QueryPayment(ctx context.Context, r *pb.ActionRequest) (*pb.ActionReply, error) {
	traceID := middleware.TraceIDFrom(ctx)
	s.log.Info(
		"source", "GoBridgeServer",
		"category", "grpc.go.query_payment",
		"traceId", traceID,
		"protocol", "grpc",
		"direction", "inbound",
		"method", "QueryPayment",
		"path", "QueryPayment",
		"message", "query payment",
	)
	return &pb.ActionReply{TraceId: r.TraceId, Code: 0, Result: "PAYMENT_FOUND"}, nil
}

func StartGRPC(l *logger.Logger, port string) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, nil, err
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(UnaryServerInterceptor(l)))
	pb.RegisterGoBridgeServer(s, NewGoBridgeServer(l))
	return s, lis, nil
}
