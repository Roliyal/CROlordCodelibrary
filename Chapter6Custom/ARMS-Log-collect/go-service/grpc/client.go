package grpcx

import (
	"context"
	"strings"
	"time"

	"armslogcollect/go-service/logger"
	"armslogcollect/go-service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type JavaGRPCClient struct {
	conn *grpc.ClientConn
	cl   pb.JavaBridgeClient
}

func NewJavaGRPCClient(addr string, l *logger.Logger) (*JavaGRPCClient, error) {
	// Plaintext for demo; for production use mTLS.
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor(l, "java-service")),
	}

	conn, err := grpc.Dial(addr, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &JavaGRPCClient{
		conn: conn,
		cl:   pb.NewJavaBridgeClient(conn),
	}, nil
}

func (c *JavaGRPCClient) Close() error { return c.conn.Close() }

func (c *JavaGRPCClient) ValidateUser(ctx context.Context, traceID string) (*pb.ActionReply, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.cl.ValidateUser(ctx, &pb.ActionRequest{TraceId: traceID, Action: "VALIDATE", Payload: "go_http"})
}

func (c *JavaGRPCClient) ReserveInventory(ctx context.Context, traceID string) (*pb.ActionReply, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.cl.ReserveInventory(ctx, &pb.ActionRequest{TraceId: traceID, Action: "RESERVE", Payload: "go_http"})
}

func (c *JavaGRPCClient) AuditOrder(ctx context.Context, traceID string) (*pb.ActionReply, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.cl.AuditOrder(ctx, &pb.ActionRequest{TraceId: traceID, Action: "AUDIT", Payload: "go_http"})
}

// helper to format addr with default port
func NormalizeAddr(addr, defaultPort string) string {
	if strings.Contains(addr, ":") {
		return addr
	}
	return addr + ":" + defaultPort
}
