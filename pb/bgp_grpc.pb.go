// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.2
// source: bgp.proto

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

// BgpApiClient is the client API for BgpApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BgpApiClient interface {
	Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	GetNeighbor(ctx context.Context, in *GetNeighborRequest, opts ...grpc.CallOption) (*GetNeighborResponse, error)
	ListNeighbor(ctx context.Context, in *ListNeighborRequest, opts ...grpc.CallOption) (*ListNeighborResponse, error)
	Network(ctx context.Context, in *NetworkRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type bgpApiClient struct {
	cc grpc.ClientConnInterface
}

func NewBgpApiClient(cc grpc.ClientConnInterface) BgpApiClient {
	return &bgpApiClient{cc}
}

func (c *bgpApiClient) Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/grp.BgpApi/Health", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bgpApiClient) GetNeighbor(ctx context.Context, in *GetNeighborRequest, opts ...grpc.CallOption) (*GetNeighborResponse, error) {
	out := new(GetNeighborResponse)
	err := c.cc.Invoke(ctx, "/grp.BgpApi/GetNeighbor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bgpApiClient) ListNeighbor(ctx context.Context, in *ListNeighborRequest, opts ...grpc.CallOption) (*ListNeighborResponse, error) {
	out := new(ListNeighborResponse)
	err := c.cc.Invoke(ctx, "/grp.BgpApi/ListNeighbor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bgpApiClient) Network(ctx context.Context, in *NetworkRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/grp.BgpApi/Network", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BgpApiServer is the server API for BgpApi service.
// All implementations must embed UnimplementedBgpApiServer
// for forward compatibility
type BgpApiServer interface {
	Health(context.Context, *HealthRequest) (*empty.Empty, error)
	GetNeighbor(context.Context, *GetNeighborRequest) (*GetNeighborResponse, error)
	ListNeighbor(context.Context, *ListNeighborRequest) (*ListNeighborResponse, error)
	Network(context.Context, *NetworkRequest) (*empty.Empty, error)
	mustEmbedUnimplementedBgpApiServer()
}

// UnimplementedBgpApiServer must be embedded to have forward compatible implementations.
type UnimplementedBgpApiServer struct {
}

func (UnimplementedBgpApiServer) Health(context.Context, *HealthRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Health not implemented")
}
func (UnimplementedBgpApiServer) GetNeighbor(context.Context, *GetNeighborRequest) (*GetNeighborResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNeighbor not implemented")
}
func (UnimplementedBgpApiServer) ListNeighbor(context.Context, *ListNeighborRequest) (*ListNeighborResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListNeighbor not implemented")
}
func (UnimplementedBgpApiServer) Network(context.Context, *NetworkRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Network not implemented")
}
func (UnimplementedBgpApiServer) mustEmbedUnimplementedBgpApiServer() {}

// UnsafeBgpApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BgpApiServer will
// result in compilation errors.
type UnsafeBgpApiServer interface {
	mustEmbedUnimplementedBgpApiServer()
}

func RegisterBgpApiServer(s grpc.ServiceRegistrar, srv BgpApiServer) {
	s.RegisterService(&BgpApi_ServiceDesc, srv)
}

func _BgpApi_Health_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BgpApiServer).Health(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.BgpApi/Health",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BgpApiServer).Health(ctx, req.(*HealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BgpApi_GetNeighbor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNeighborRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BgpApiServer).GetNeighbor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.BgpApi/GetNeighbor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BgpApiServer).GetNeighbor(ctx, req.(*GetNeighborRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BgpApi_ListNeighbor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListNeighborRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BgpApiServer).ListNeighbor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.BgpApi/ListNeighbor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BgpApiServer).ListNeighbor(ctx, req.(*ListNeighborRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BgpApi_Network_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BgpApiServer).Network(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grp.BgpApi/Network",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BgpApiServer).Network(ctx, req.(*NetworkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BgpApi_ServiceDesc is the grpc.ServiceDesc for BgpApi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BgpApi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grp.BgpApi",
	HandlerType: (*BgpApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Health",
			Handler:    _BgpApi_Health_Handler,
		},
		{
			MethodName: "GetNeighbor",
			Handler:    _BgpApi_GetNeighbor_Handler,
		},
		{
			MethodName: "ListNeighbor",
			Handler:    _BgpApi_ListNeighbor_Handler,
		},
		{
			MethodName: "Network",
			Handler:    _BgpApi_Network_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "bgp.proto",
}
