package route

import (
	"context"
	"fmt"
	"net"

	"github.com/terassyi/grp/pb"
	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const logPath = "/var/log/grp/route"
const RouteServerPort = 6789

type RouteServer struct {
	pb.UnimplementedRouteApiServer
	routes map[string]netlink.Route
}

func New() (*RouteServer, error) {
	r := &RouteServer{
		routes: map[string]netlink.Route{},
	}
	return r, nil
}

func (r *RouteServer) Serve() error {
	server := grpc.NewServer()
	pb.RegisterRouteApiServer(server, r)
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", RouteServerPort))
	if err != nil {
		return err
	}
	return server.Serve(listener)
}

func prepare() (map[string]netlink.Route, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}
	routes := make(map[string]netlink.Route)
	netlink.RouteList()
	return routes, nil
}

func (r *RouteServer) Health(ctx context.Context, in *pb.HealthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
