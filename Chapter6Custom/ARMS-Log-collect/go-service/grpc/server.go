package grpcx

import (
	"context"
	"net"

	"armslogcollect/go-service/logger"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/proto" // ensure proto codec is registered
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

type GoBridgeDynamicServer interface {
	ProcessPayment(ctx context.Context, req proto.Message) (proto.Message, error)
	QueryPayment(ctx context.Context, req proto.Message) (proto.Message, error)
	IssueRefund(ctx context.Context, req proto.Message) (proto.Message, error)
}

type goBridgeService struct {
    log *logger.Logger
}

func (s *goBridgeService) ProcessPayment(ctx context.Context, req proto.Message) (proto.Message, error) {
	// business: pretend everything is OK
	bd, _ := getBridgeDesc()
	reply := dynamicpb.NewMessage(bd.actionReply)
	setStr(reply, "trace_id", mustGetStr(req.(*dynamicpb.Message), "trace_id"))
	setStr(reply, "result", "PAY_OK")
	return reply, nil
}

func (s *goBridgeService) QueryPayment(ctx context.Context, req proto.Message) (proto.Message, error) {
	bd, _ := getBridgeDesc()
	reply := dynamicpb.NewMessage(bd.actionReply)
	setStr(reply, "trace_id", mustGetStr(req.(*dynamicpb.Message), "trace_id"))
	setStr(reply, "result", "QUERY_OK")
	return reply, nil
}

func (s *goBridgeService) IssueRefund(ctx context.Context, req proto.Message) (proto.Message, error) {
	bd, _ := getBridgeDesc()
	reply := dynamicpb.NewMessage(bd.actionReply)
	setStr(reply, "trace_id", mustGetStr(req.(*dynamicpb.Message), "trace_id"))
	setStr(reply, "result", "REFUND_OK")
	return reply, nil
}

// StartGRPC starts the Go gRPC server (bridge.v1.GoBridge) and returns the server and listener.
func StartGRPC(l *logger.Logger, port string) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, nil, err
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryServerInterceptor(l),
		),
	)

	RegisterGoBridgeDynamicServer(s, &goBridgeService{log: l})
	return s, lis, nil
}

// ---- manual registration (no generated code) ----

func RegisterGoBridgeDynamicServer(s grpc.ServiceRegistrar, srv GoBridgeDynamicServer) {
	s.RegisterService(&GoBridge_ServiceDesc, srv)
}

var GoBridge_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "bridge.v1.GoBridge",
	HandlerType: (*GoBridgeDynamicServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "ProcessPayment", Handler: _GoBridge_ProcessPayment_Handler},
		{MethodName: "QueryPayment", Handler: _GoBridge_QueryPayment_Handler},
		{MethodName: "IssueRefund", Handler: _GoBridge_IssueRefund_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/bridge/v1/bridge.proto",
}

func _GoBridge_ProcessPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	bd, err := getBridgeDesc()
	if err != nil {
		return nil, err
	}
	in := dynamicpb.NewMessage(bd.actionReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoBridgeDynamicServer).ProcessPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/bridge.v1.GoBridge/ProcessPayment"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoBridgeDynamicServer).ProcessPayment(ctx, req.(proto.Message))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoBridge_QueryPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	bd, err := getBridgeDesc()
	if err != nil {
		return nil, err
	}
	in := dynamicpb.NewMessage(bd.actionReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoBridgeDynamicServer).QueryPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/bridge.v1.GoBridge/QueryPayment"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoBridgeDynamicServer).QueryPayment(ctx, req.(proto.Message))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoBridge_IssueRefund_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	bd, err := getBridgeDesc()
	if err != nil {
		return nil, err
	}
	in := dynamicpb.NewMessage(bd.actionReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoBridgeDynamicServer).IssueRefund(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/bridge.v1.GoBridge/IssueRefund"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoBridgeDynamicServer).IssueRefund(ctx, req.(proto.Message))
	}
	return interceptor(ctx, in, info, handler)
}
