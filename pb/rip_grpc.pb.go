// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.2
// source: rip.proto

package pb

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RipApiClient is the client API for RipApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RipApiClient interface {
	Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	GetLogPath(ctx context.Context, in *GetLogPathRequest, opts ...grpc.CallOption) (*GetLogPathResponse, error)
	GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error)
	Show(ctx context.Context, in *RipShowRequest, opts ...grpc.CallOption) (*RipShowResponse, error)
	Network(ctx context.Context, in *NetworkRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type ripApiClient struct {
	cc grpc.ClientConnInterface
}

func NewRipApiClient(cc grpc.ClientConnInterface) RipApiClient {
	return &ripApiClient{cc}
}

func (c *ripApiClient) Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/grp.RipApi/Health", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ripApiClient) GetLogPath(ctx context.Context, in *GetLogPathRequest, opts ...grpc.CallOption) (*GetLogPathResponse, error) {
	out := new(GetLogPathResponse)
	err := c.cc.Invoke(ctx, "/grp.RipApi/GetLogPath", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ripApiClient) GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error) {
	out := new(GetVersionResponse)
	err := c.cc.Invoke(ctx, "/grp.RipApi/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ripApiClient) Show(ctx context.Context, in *RipShowRequest, opts ...grpc.CallOption) (*RipShowResponse, error) {
	out := new(RipShowResponse)
	err := c.cc.Invoke(ctx, "/grp.RipApi/Show", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ripApiClient) Network(ctx context.Context, in *NetworkRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/grp.RipApi/Network", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RipApiServer is the server API for RipApi service.
// All implementations must embed UnimplementedRipApiServer
// for forward compatibility
type RipApiServer interface {
	Health(context.Context, *HealthRequest) (*empty.Empty, error)
	GetLogPath(context.Context, *GetLogPathRequest) (*GetLogPathResponse, error)
	GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error)
	Show(context.Context, *RipShowRequest) (*RipShowResponse, error)
	Network(context.Context, *NetworkRequest) (*empty.Empty, error)
	mustEmbedUnimplementedRipApiServer()
}

// UnimplementedRipApiServer must be embedded to have forward compatible implementations.
type UnimplementedRipApiServer struct {
}

func (UnimplementedRipApiServer) Health(context.Context, *HealthRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Health not implemented")
}
func (UnimplementedRipApiServer) GetLogPath(context.Context, *GetLogPathRequest) (*GetLogPathResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLogPath not implemented")
}
func (UnimplementedRipApiServer) GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}
func (UnimplementedRipApiServer) Show(context.Context, *RipShowRequest) (*RipShowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Show not implemented")
}
func (UnimplementedRipApiServer) Network(context.Context, *NetworkRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Network not implemented")
}
func (UnimplementedRipApiServer) mustEmbedUnimplementedRipApiServer() {}

// UnsafeRipApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RipApiServer will
// result in compilation errors.
type UnsafeRipApiServer interface {
	mustEmbedUnimplementedRipApiServer()
}

func RegisterRipApiServer(s grpc.ServiceRegistrar, srv RipApiServer) {
	s.RegisterService(&RipApi_ServiceDesc, srv)
}

func _RipApi_Health_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RipApiServer).Health(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.RipApi/Health",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RipApiServer).Health(ctx, req.(*HealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RipApi_GetLogPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLogPathRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RipApiServer).GetLogPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.RipApi/GetLogPath",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RipApiServer).GetLogPath(ctx, req.(*GetLogPathRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RipApi_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RipApiServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.RipApi/GetVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RipApiServer).GetVersion(ctx, req.(*GetVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RipApi_Show_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RipShowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RipApiServer).Show(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.RipApi/Show",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RipApiServer).Show(ctx, req.(*RipShowRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RipApi_Network_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RipApiServer).Network(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.RipApi/Network",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RipApiServer).Network(ctx, req.(*NetworkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RipApi_ServiceDesc is the grpc.ServiceDesc for RipApi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RipApi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grp.RipApi",
	HandlerType: (*RipApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Health",
			Handler:    _RipApi_Health_Handler,
		},
		{
			MethodName: "GetLogPath",
			Handler:    _RipApi_GetLogPath_Handler,
		},
		{
			MethodName: "GetVersion",
			Handler:    _RipApi_GetVersion_Handler,
		},
		{
			MethodName: "Show",
			Handler:    _RipApi_Show_Handler,
		},
		{
			MethodName: "Network",
			Handler:    _RipApi_Network_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rip.proto",
}