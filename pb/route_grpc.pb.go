// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.2
// source: route.proto

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

// RouteApiClient is the client API for RouteApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RouteApiClient interface {
	Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	SetRoute(ctx context.Context, in *SetRouteRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	DeleteRoute(ctx context.Context, in *DeleteRouteRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type routeApiClient struct {
	cc grpc.ClientConnInterface
}

func NewRouteApiClient(cc grpc.ClientConnInterface) RouteApiClient {
	return &routeApiClient{cc}
}

func (c *routeApiClient) Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/grp.RouteApi/Health", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *routeApiClient) SetRoute(ctx context.Context, in *SetRouteRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/grp.RouteApi/SetRoute", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *routeApiClient) DeleteRoute(ctx context.Context, in *DeleteRouteRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/grp.RouteApi/DeleteRoute", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RouteApiServer is the server API for RouteApi service.
// All implementations must embed UnimplementedRouteApiServer
// for forward compatibility
type RouteApiServer interface {
	Health(context.Context, *HealthRequest) (*empty.Empty, error)
	SetRoute(context.Context, *SetRouteRequest) (*empty.Empty, error)
	DeleteRoute(context.Context, *DeleteRouteRequest) (*empty.Empty, error)
	mustEmbedUnimplementedRouteApiServer()
}

// UnimplementedRouteApiServer must be embedded to have forward compatible implementations.
type UnimplementedRouteApiServer struct {
}

func (UnimplementedRouteApiServer) Health(context.Context, *HealthRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Health not implemented")
}
func (UnimplementedRouteApiServer) SetRoute(context.Context, *SetRouteRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetRoute not implemented")
}
func (UnimplementedRouteApiServer) DeleteRoute(context.Context, *DeleteRouteRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRoute not implemented")
}
func (UnimplementedRouteApiServer) mustEmbedUnimplementedRouteApiServer() {}

// UnsafeRouteApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RouteApiServer will
// result in compilation errors.
type UnsafeRouteApiServer interface {
	mustEmbedUnimplementedRouteApiServer()
}

func RegisterRouteApiServer(s grpc.ServiceRegistrar, srv RouteApiServer) {
	s.RegisterService(&RouteApi_ServiceDesc, srv)
}

func _RouteApi_Health_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RouteApiServer).Health(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.RouteApi/Health",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RouteApiServer).Health(ctx, req.(*HealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RouteApi_SetRoute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetRouteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RouteApiServer).SetRoute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.RouteApi/SetRoute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RouteApiServer).SetRoute(ctx, req.(*SetRouteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RouteApi_DeleteRoute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRouteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RouteApiServer).DeleteRoute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.RouteApi/DeleteRoute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RouteApiServer).DeleteRoute(ctx, req.(*DeleteRouteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RouteApi_ServiceDesc is the grpc.ServiceDesc for RouteApi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RouteApi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grp.RouteApi",
	HandlerType: (*RouteApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Health",
			Handler:    _RouteApi_Health_Handler,
		},
		{
			MethodName: "SetRoute",
			Handler:    _RouteApi_SetRoute_Handler,
		},
		{
			MethodName: "DeleteRoute",
			Handler:    _RouteApi_DeleteRoute_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "route.proto",
}
