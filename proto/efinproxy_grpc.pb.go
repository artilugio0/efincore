// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.28.2
// source: efinproxy.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	EfinProxy_GetStats_FullMethodName        = "/efincore.EfinProxy/GetStats"
	EfinProxy_GetRequestsIn_FullMethodName   = "/efincore.EfinProxy/GetRequestsIn"
	EfinProxy_RequestsMod_FullMethodName     = "/efincore.EfinProxy/RequestsMod"
	EfinProxy_GetRequestsOut_FullMethodName  = "/efincore.EfinProxy/GetRequestsOut"
	EfinProxy_GetResponsesIn_FullMethodName  = "/efincore.EfinProxy/GetResponsesIn"
	EfinProxy_ResponsesMod_FullMethodName    = "/efincore.EfinProxy/ResponsesMod"
	EfinProxy_GetResponsesOut_FullMethodName = "/efincore.EfinProxy/GetResponsesOut"
)

// EfinProxyClient is the client API for EfinProxy service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EfinProxyClient interface {
	GetStats(ctx context.Context, in *GetStatsInput, opts ...grpc.CallOption) (*GetStatsOutput, error)
	GetRequestsIn(ctx context.Context, in *GetRequestsInInput, opts ...grpc.CallOption) (EfinProxy_GetRequestsInClient, error)
	RequestsMod(ctx context.Context, opts ...grpc.CallOption) (EfinProxy_RequestsModClient, error)
	GetRequestsOut(ctx context.Context, in *GetRequestsOutInput, opts ...grpc.CallOption) (EfinProxy_GetRequestsOutClient, error)
	GetResponsesIn(ctx context.Context, in *GetResponsesInInput, opts ...grpc.CallOption) (EfinProxy_GetResponsesInClient, error)
	ResponsesMod(ctx context.Context, opts ...grpc.CallOption) (EfinProxy_ResponsesModClient, error)
	GetResponsesOut(ctx context.Context, in *GetResponsesOutInput, opts ...grpc.CallOption) (EfinProxy_GetResponsesOutClient, error)
}

type efinProxyClient struct {
	cc grpc.ClientConnInterface
}

func NewEfinProxyClient(cc grpc.ClientConnInterface) EfinProxyClient {
	return &efinProxyClient{cc}
}

func (c *efinProxyClient) GetStats(ctx context.Context, in *GetStatsInput, opts ...grpc.CallOption) (*GetStatsOutput, error) {
	out := new(GetStatsOutput)
	err := c.cc.Invoke(ctx, EfinProxy_GetStats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *efinProxyClient) GetRequestsIn(ctx context.Context, in *GetRequestsInInput, opts ...grpc.CallOption) (EfinProxy_GetRequestsInClient, error) {
	stream, err := c.cc.NewStream(ctx, &EfinProxy_ServiceDesc.Streams[0], EfinProxy_GetRequestsIn_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &efinProxyGetRequestsInClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EfinProxy_GetRequestsInClient interface {
	Recv() (*Request, error)
	grpc.ClientStream
}

type efinProxyGetRequestsInClient struct {
	grpc.ClientStream
}

func (x *efinProxyGetRequestsInClient) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *efinProxyClient) RequestsMod(ctx context.Context, opts ...grpc.CallOption) (EfinProxy_RequestsModClient, error) {
	stream, err := c.cc.NewStream(ctx, &EfinProxy_ServiceDesc.Streams[1], EfinProxy_RequestsMod_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &efinProxyRequestsModClient{stream}
	return x, nil
}

type EfinProxy_RequestsModClient interface {
	Send(*Request) error
	Recv() (*Request, error)
	grpc.ClientStream
}

type efinProxyRequestsModClient struct {
	grpc.ClientStream
}

func (x *efinProxyRequestsModClient) Send(m *Request) error {
	return x.ClientStream.SendMsg(m)
}

func (x *efinProxyRequestsModClient) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *efinProxyClient) GetRequestsOut(ctx context.Context, in *GetRequestsOutInput, opts ...grpc.CallOption) (EfinProxy_GetRequestsOutClient, error) {
	stream, err := c.cc.NewStream(ctx, &EfinProxy_ServiceDesc.Streams[2], EfinProxy_GetRequestsOut_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &efinProxyGetRequestsOutClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EfinProxy_GetRequestsOutClient interface {
	Recv() (*Request, error)
	grpc.ClientStream
}

type efinProxyGetRequestsOutClient struct {
	grpc.ClientStream
}

func (x *efinProxyGetRequestsOutClient) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *efinProxyClient) GetResponsesIn(ctx context.Context, in *GetResponsesInInput, opts ...grpc.CallOption) (EfinProxy_GetResponsesInClient, error) {
	stream, err := c.cc.NewStream(ctx, &EfinProxy_ServiceDesc.Streams[3], EfinProxy_GetResponsesIn_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &efinProxyGetResponsesInClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EfinProxy_GetResponsesInClient interface {
	Recv() (*Response, error)
	grpc.ClientStream
}

type efinProxyGetResponsesInClient struct {
	grpc.ClientStream
}

func (x *efinProxyGetResponsesInClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *efinProxyClient) ResponsesMod(ctx context.Context, opts ...grpc.CallOption) (EfinProxy_ResponsesModClient, error) {
	stream, err := c.cc.NewStream(ctx, &EfinProxy_ServiceDesc.Streams[4], EfinProxy_ResponsesMod_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &efinProxyResponsesModClient{stream}
	return x, nil
}

type EfinProxy_ResponsesModClient interface {
	Send(*Response) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type efinProxyResponsesModClient struct {
	grpc.ClientStream
}

func (x *efinProxyResponsesModClient) Send(m *Response) error {
	return x.ClientStream.SendMsg(m)
}

func (x *efinProxyResponsesModClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *efinProxyClient) GetResponsesOut(ctx context.Context, in *GetResponsesOutInput, opts ...grpc.CallOption) (EfinProxy_GetResponsesOutClient, error) {
	stream, err := c.cc.NewStream(ctx, &EfinProxy_ServiceDesc.Streams[5], EfinProxy_GetResponsesOut_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &efinProxyGetResponsesOutClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EfinProxy_GetResponsesOutClient interface {
	Recv() (*Response, error)
	grpc.ClientStream
}

type efinProxyGetResponsesOutClient struct {
	grpc.ClientStream
}

func (x *efinProxyGetResponsesOutClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EfinProxyServer is the server API for EfinProxy service.
// All implementations must embed UnimplementedEfinProxyServer
// for forward compatibility
type EfinProxyServer interface {
	GetStats(context.Context, *GetStatsInput) (*GetStatsOutput, error)
	GetRequestsIn(*GetRequestsInInput, EfinProxy_GetRequestsInServer) error
	RequestsMod(EfinProxy_RequestsModServer) error
	GetRequestsOut(*GetRequestsOutInput, EfinProxy_GetRequestsOutServer) error
	GetResponsesIn(*GetResponsesInInput, EfinProxy_GetResponsesInServer) error
	ResponsesMod(EfinProxy_ResponsesModServer) error
	GetResponsesOut(*GetResponsesOutInput, EfinProxy_GetResponsesOutServer) error
	mustEmbedUnimplementedEfinProxyServer()
}

// UnimplementedEfinProxyServer must be embedded to have forward compatible implementations.
type UnimplementedEfinProxyServer struct {
}

func (UnimplementedEfinProxyServer) GetStats(context.Context, *GetStatsInput) (*GetStatsOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}
func (UnimplementedEfinProxyServer) GetRequestsIn(*GetRequestsInInput, EfinProxy_GetRequestsInServer) error {
	return status.Errorf(codes.Unimplemented, "method GetRequestsIn not implemented")
}
func (UnimplementedEfinProxyServer) RequestsMod(EfinProxy_RequestsModServer) error {
	return status.Errorf(codes.Unimplemented, "method RequestsMod not implemented")
}
func (UnimplementedEfinProxyServer) GetRequestsOut(*GetRequestsOutInput, EfinProxy_GetRequestsOutServer) error {
	return status.Errorf(codes.Unimplemented, "method GetRequestsOut not implemented")
}
func (UnimplementedEfinProxyServer) GetResponsesIn(*GetResponsesInInput, EfinProxy_GetResponsesInServer) error {
	return status.Errorf(codes.Unimplemented, "method GetResponsesIn not implemented")
}
func (UnimplementedEfinProxyServer) ResponsesMod(EfinProxy_ResponsesModServer) error {
	return status.Errorf(codes.Unimplemented, "method ResponsesMod not implemented")
}
func (UnimplementedEfinProxyServer) GetResponsesOut(*GetResponsesOutInput, EfinProxy_GetResponsesOutServer) error {
	return status.Errorf(codes.Unimplemented, "method GetResponsesOut not implemented")
}
func (UnimplementedEfinProxyServer) mustEmbedUnimplementedEfinProxyServer() {}

// UnsafeEfinProxyServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EfinProxyServer will
// result in compilation errors.
type UnsafeEfinProxyServer interface {
	mustEmbedUnimplementedEfinProxyServer()
}

func RegisterEfinProxyServer(s grpc.ServiceRegistrar, srv EfinProxyServer) {
	s.RegisterService(&EfinProxy_ServiceDesc, srv)
}

func _EfinProxy_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatsInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EfinProxyServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EfinProxy_GetStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EfinProxyServer).GetStats(ctx, req.(*GetStatsInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _EfinProxy_GetRequestsIn_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetRequestsInInput)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EfinProxyServer).GetRequestsIn(m, &efinProxyGetRequestsInServer{stream})
}

type EfinProxy_GetRequestsInServer interface {
	Send(*Request) error
	grpc.ServerStream
}

type efinProxyGetRequestsInServer struct {
	grpc.ServerStream
}

func (x *efinProxyGetRequestsInServer) Send(m *Request) error {
	return x.ServerStream.SendMsg(m)
}

func _EfinProxy_RequestsMod_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EfinProxyServer).RequestsMod(&efinProxyRequestsModServer{stream})
}

type EfinProxy_RequestsModServer interface {
	Send(*Request) error
	Recv() (*Request, error)
	grpc.ServerStream
}

type efinProxyRequestsModServer struct {
	grpc.ServerStream
}

func (x *efinProxyRequestsModServer) Send(m *Request) error {
	return x.ServerStream.SendMsg(m)
}

func (x *efinProxyRequestsModServer) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _EfinProxy_GetRequestsOut_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetRequestsOutInput)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EfinProxyServer).GetRequestsOut(m, &efinProxyGetRequestsOutServer{stream})
}

type EfinProxy_GetRequestsOutServer interface {
	Send(*Request) error
	grpc.ServerStream
}

type efinProxyGetRequestsOutServer struct {
	grpc.ServerStream
}

func (x *efinProxyGetRequestsOutServer) Send(m *Request) error {
	return x.ServerStream.SendMsg(m)
}

func _EfinProxy_GetResponsesIn_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetResponsesInInput)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EfinProxyServer).GetResponsesIn(m, &efinProxyGetResponsesInServer{stream})
}

type EfinProxy_GetResponsesInServer interface {
	Send(*Response) error
	grpc.ServerStream
}

type efinProxyGetResponsesInServer struct {
	grpc.ServerStream
}

func (x *efinProxyGetResponsesInServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func _EfinProxy_ResponsesMod_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EfinProxyServer).ResponsesMod(&efinProxyResponsesModServer{stream})
}

type EfinProxy_ResponsesModServer interface {
	Send(*Response) error
	Recv() (*Response, error)
	grpc.ServerStream
}

type efinProxyResponsesModServer struct {
	grpc.ServerStream
}

func (x *efinProxyResponsesModServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *efinProxyResponsesModServer) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _EfinProxy_GetResponsesOut_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetResponsesOutInput)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EfinProxyServer).GetResponsesOut(m, &efinProxyGetResponsesOutServer{stream})
}

type EfinProxy_GetResponsesOutServer interface {
	Send(*Response) error
	grpc.ServerStream
}

type efinProxyGetResponsesOutServer struct {
	grpc.ServerStream
}

func (x *efinProxyGetResponsesOutServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

// EfinProxy_ServiceDesc is the grpc.ServiceDesc for EfinProxy service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EfinProxy_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "efincore.EfinProxy",
	HandlerType: (*EfinProxyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetStats",
			Handler:    _EfinProxy_GetStats_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetRequestsIn",
			Handler:       _EfinProxy_GetRequestsIn_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "RequestsMod",
			Handler:       _EfinProxy_RequestsMod_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "GetRequestsOut",
			Handler:       _EfinProxy_GetRequestsOut_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetResponsesIn",
			Handler:       _EfinProxy_GetResponsesIn_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "ResponsesMod",
			Handler:       _EfinProxy_ResponsesMod_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "GetResponsesOut",
			Handler:       _EfinProxy_GetResponsesOut_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "efinproxy.proto",
}
