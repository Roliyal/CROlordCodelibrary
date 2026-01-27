package grpcx

import (
	"context"
	"time"

	"armslogcollect/go-service/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/proto" // ensure proto codec is registered
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/dynamicpb"
)

type JavaGRPCClient struct {
	conn    *grpc.ClientConn
	timeout time.Duration
}

func NewJavaGRPCClient(addr string, l *logger.Logger) (*JavaGRPCClient, error) {
	// Ensure proto descriptor is loadable early so errors are clear.
	if _, err := getBridgeDesc(); err != nil {
		return nil, err
	}

	// Dial (non-blocking): in k8s the downstream (java-service) may not be ready
	// when this pod starts. Avoid exiting the process just because the initial
	// connection isn't up yet; gRPC will reconnect in the background.
	conn, err := grpc.DialContext(
		context.Background(),
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(UnaryClientInterceptor(l, "java-service")),
	)
	if err != nil {
		return nil, err
	}

	return &JavaGRPCClient{conn: conn, timeout: 1500 * time.Millisecond}, nil
}

func (c *JavaGRPCClient) Close() {
	_ = c.conn.Close()
}

// ---- typed helpers used by HTTP handlers ----

func (c *JavaGRPCClient) ValidateUser(ctx context.Context, traceID string) (string, error) {
	return c.call(ctx, traceID, "ValidateUser", "{}", "/bridge.v1.JavaBridge/ValidateUser")
}

func (c *JavaGRPCClient) ReserveInventory(ctx context.Context, traceID string) (string, error) {
	return c.call(ctx, traceID, "ReserveInventory", "{}", "/bridge.v1.JavaBridge/ReserveInventory")
}

func (c *JavaGRPCClient) AuditOrder(ctx context.Context, traceID string) (string, error) {
	return c.call(ctx, traceID, "AuditOrder", "{}", "/bridge.v1.JavaBridge/AuditOrder")
}

// call invokes a unary RPC against Java service using runtime descriptors.
func (c *JavaGRPCClient) call(parent context.Context, traceID, action, payload, fullMethod string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, c.timeout)
	defer cancel()

	// outgoing metadata for tracing
	if traceID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", traceID)
	}

	req, err := newActionRequest(traceID, action, payload)
	if err != nil {
		return "", err
	}

	resp, err := newActionReply()
	if err != nil {
		return "", err
	}

	if err := c.conn.Invoke(ctx, fullMethod, req, resp); err != nil {
		return "", err
	}

	return getStringField(resp, "result"), nil
}

// compile-time check that we only ever pass proto messages
var _ = (*dynamicpb.Message)(nil)
