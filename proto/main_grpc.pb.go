// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: proto/main.proto

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

// ControlServiceClient is the client API for ControlService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ControlServiceClient interface {
	Publish(ctx context.Context, in *PublishRequest, opts ...grpc.CallOption) (*PublishResponse, error)
	Subscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (*SubscribeResponse, error)
	Unsubscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (*SubscribeResponse, error)
}

type controlServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewControlServiceClient(cc grpc.ClientConnInterface) ControlServiceClient {
	return &controlServiceClient{cc}
}

func (c *controlServiceClient) Publish(ctx context.Context, in *PublishRequest, opts ...grpc.CallOption) (*PublishResponse, error) {
	out := new(PublishResponse)
	err := c.cc.Invoke(ctx, "/proto.ControlService/Publish", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) Subscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (*SubscribeResponse, error) {
	out := new(SubscribeResponse)
	err := c.cc.Invoke(ctx, "/proto.ControlService/Subscribe", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) Unsubscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (*SubscribeResponse, error) {
	out := new(SubscribeResponse)
	err := c.cc.Invoke(ctx, "/proto.ControlService/Unsubscribe", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ControlServiceServer is the server API for ControlService service.
// All implementations must embed UnimplementedControlServiceServer
// for forward compatibility
type ControlServiceServer interface {
	Publish(context.Context, *PublishRequest) (*PublishResponse, error)
	Subscribe(context.Context, *SubscribeRequest) (*SubscribeResponse, error)
	Unsubscribe(context.Context, *SubscribeRequest) (*SubscribeResponse, error)
	mustEmbedUnimplementedControlServiceServer()
}

// UnimplementedControlServiceServer must be embedded to have forward compatible implementations.
type UnimplementedControlServiceServer struct {
}

func (UnimplementedControlServiceServer) Publish(context.Context, *PublishRequest) (*PublishResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Publish not implemented")
}
func (UnimplementedControlServiceServer) Subscribe(context.Context, *SubscribeRequest) (*SubscribeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedControlServiceServer) Unsubscribe(context.Context, *SubscribeRequest) (*SubscribeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Unsubscribe not implemented")
}
func (UnimplementedControlServiceServer) mustEmbedUnimplementedControlServiceServer() {}

// UnsafeControlServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ControlServiceServer will
// result in compilation errors.
type UnsafeControlServiceServer interface {
	mustEmbedUnimplementedControlServiceServer()
}

func RegisterControlServiceServer(s grpc.ServiceRegistrar, srv ControlServiceServer) {
	s.RegisterService(&ControlService_ServiceDesc, srv)
}

func _ControlService_Publish_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).Publish(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ControlService/Publish",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).Publish(ctx, req.(*PublishRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_Subscribe_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscribeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).Subscribe(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ControlService/Subscribe",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).Subscribe(ctx, req.(*SubscribeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_Unsubscribe_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscribeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).Unsubscribe(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ControlService/Unsubscribe",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).Unsubscribe(ctx, req.(*SubscribeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ControlService_ServiceDesc is the grpc.ServiceDesc for ControlService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ControlService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ControlService",
	HandlerType: (*ControlServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Publish",
			Handler:    _ControlService_Publish_Handler,
		},
		{
			MethodName: "Subscribe",
			Handler:    _ControlService_Subscribe_Handler,
		},
		{
			MethodName: "Unsubscribe",
			Handler:    _ControlService_Unsubscribe_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/main.proto",
}

// MsgRouterClient is the client API for MsgRouter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgRouterClient interface {
	Listen(ctx context.Context, in *ListenRequest, opts ...grpc.CallOption) (MsgRouter_ListenClient, error)
}

type msgRouterClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgRouterClient(cc grpc.ClientConnInterface) MsgRouterClient {
	return &msgRouterClient{cc}
}

func (c *msgRouterClient) Listen(ctx context.Context, in *ListenRequest, opts ...grpc.CallOption) (MsgRouter_ListenClient, error) {
	stream, err := c.cc.NewStream(ctx, &MsgRouter_ServiceDesc.Streams[0], "/proto.MsgRouter/Listen", opts...)
	if err != nil {
		return nil, err
	}
	x := &msgRouterListenClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type MsgRouter_ListenClient interface {
	Recv() (*Message, error)
	grpc.ClientStream
}

type msgRouterListenClient struct {
	grpc.ClientStream
}

func (x *msgRouterListenClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// MsgRouterServer is the server API for MsgRouter service.
// All implementations must embed UnimplementedMsgRouterServer
// for forward compatibility
type MsgRouterServer interface {
	Listen(*ListenRequest, MsgRouter_ListenServer) error
	mustEmbedUnimplementedMsgRouterServer()
}

// UnimplementedMsgRouterServer must be embedded to have forward compatible implementations.
type UnimplementedMsgRouterServer struct {
}

func (UnimplementedMsgRouterServer) Listen(*ListenRequest, MsgRouter_ListenServer) error {
	return status.Errorf(codes.Unimplemented, "method Listen not implemented")
}
func (UnimplementedMsgRouterServer) mustEmbedUnimplementedMsgRouterServer() {}

// UnsafeMsgRouterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgRouterServer will
// result in compilation errors.
type UnsafeMsgRouterServer interface {
	mustEmbedUnimplementedMsgRouterServer()
}

func RegisterMsgRouterServer(s grpc.ServiceRegistrar, srv MsgRouterServer) {
	s.RegisterService(&MsgRouter_ServiceDesc, srv)
}

func _MsgRouter_Listen_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListenRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(MsgRouterServer).Listen(m, &msgRouterListenServer{stream})
}

type MsgRouter_ListenServer interface {
	Send(*Message) error
	grpc.ServerStream
}

type msgRouterListenServer struct {
	grpc.ServerStream
}

func (x *msgRouterListenServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

// MsgRouter_ServiceDesc is the grpc.ServiceDesc for MsgRouter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MsgRouter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.MsgRouter",
	HandlerType: (*MsgRouterServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _MsgRouter_Listen_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/main.proto",
}

// ValidationRouterClient is the client API for ValidationRouter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ValidationRouterClient interface {
	Handle(ctx context.Context, opts ...grpc.CallOption) (ValidationRouter_HandleClient, error)
}

type validationRouterClient struct {
	cc grpc.ClientConnInterface
}

func NewValidationRouterClient(cc grpc.ClientConnInterface) ValidationRouterClient {
	return &validationRouterClient{cc}
}

func (c *validationRouterClient) Handle(ctx context.Context, opts ...grpc.CallOption) (ValidationRouter_HandleClient, error) {
	stream, err := c.cc.NewStream(ctx, &ValidationRouter_ServiceDesc.Streams[0], "/proto.ValidationRouter/Handle", opts...)
	if err != nil {
		return nil, err
	}
	x := &validationRouterHandleClient{stream}
	return x, nil
}

type ValidationRouter_HandleClient interface {
	Send(*ValidatedMessage) error
	Recv() (*Message, error)
	grpc.ClientStream
}

type validationRouterHandleClient struct {
	grpc.ClientStream
}

func (x *validationRouterHandleClient) Send(m *ValidatedMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *validationRouterHandleClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ValidationRouterServer is the server API for ValidationRouter service.
// All implementations must embed UnimplementedValidationRouterServer
// for forward compatibility
type ValidationRouterServer interface {
	Handle(ValidationRouter_HandleServer) error
	mustEmbedUnimplementedValidationRouterServer()
}

// UnimplementedValidationRouterServer must be embedded to have forward compatible implementations.
type UnimplementedValidationRouterServer struct {
}

func (UnimplementedValidationRouterServer) Handle(ValidationRouter_HandleServer) error {
	return status.Errorf(codes.Unimplemented, "method Handle not implemented")
}
func (UnimplementedValidationRouterServer) mustEmbedUnimplementedValidationRouterServer() {}

// UnsafeValidationRouterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ValidationRouterServer will
// result in compilation errors.
type UnsafeValidationRouterServer interface {
	mustEmbedUnimplementedValidationRouterServer()
}

func RegisterValidationRouterServer(s grpc.ServiceRegistrar, srv ValidationRouterServer) {
	s.RegisterService(&ValidationRouter_ServiceDesc, srv)
}

func _ValidationRouter_Handle_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ValidationRouterServer).Handle(&validationRouterHandleServer{stream})
}

type ValidationRouter_HandleServer interface {
	Send(*Message) error
	Recv() (*ValidatedMessage, error)
	grpc.ServerStream
}

type validationRouterHandleServer struct {
	grpc.ServerStream
}

func (x *validationRouterHandleServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func (x *validationRouterHandleServer) Recv() (*ValidatedMessage, error) {
	m := new(ValidatedMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ValidationRouter_ServiceDesc is the grpc.ServiceDesc for ValidationRouter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ValidationRouter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ValidationRouter",
	HandlerType: (*ValidationRouterServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Handle",
			Handler:       _ValidationRouter_Handle_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/main.proto",
}
