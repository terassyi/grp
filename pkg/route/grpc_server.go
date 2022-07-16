package route

import (
	"context"
	"fmt"
	"net"

	"github.com/terassyi/grp/pb"
	"github.com/terassyi/grp/pkg/log"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const logPath = "/var/log/grp/route"
const RouteServerPort = 6789

type RouteServer struct {
	pb.UnimplementedRouteApiServer
	routes map[string]*Route
	logger log.Logger
}

func New() (*RouteServer, error) {
	logger, err := log.New(log.Info, "stdout") // for dev
	if err != nil {
		return nil, err
	}
	all, err := getAllRoutes()
	if err != nil {
		return nil, err
	}
	for _, r := range all {
		logger.Info("%s", r.String())
	}
	r := &RouteServer{
		routes: all,
		logger: logger,
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

func getAllRoutes() (map[string]*Route, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}
	routeMap := make(map[string]*Route)
	for _, link := range links {
		routes, err := netlink.RouteList(link, nl.FAMILY_V4)
		if err != nil {
			return nil, err
		}
		for _, route := range routes {
			// default, if bgp route is found, external set true.
			routeMap[route.Dst.String()] = RouteFromNetLink(route, true)
		}
	}
	return routeMap, nil
}

func (r *RouteServer) Health(ctx context.Context, in *pb.HealthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (r *RouteServer) SetRoute(ctx context.Context, in *pb.SetRouteRequest) (*emptypb.Empty, error) {

	return &emptypb.Empty{}, nil
}
