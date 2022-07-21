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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	logPath                 = "/var/log/grp/route"
	DefaultRouteManagerPort = 6789
	DefaultRouteManagerHost = "localhost"
)

type RouteManger struct {
	pb.UnimplementedRouteApiServer
	routes   map[string]*Route
	logger   log.Logger
	endpoint string
}

func New(host string, port int) (*RouteManger, error) {
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
	endpoint := fmt.Sprintf("%s:%d", host, port)
	r := &RouteManger{
		routes:   all,
		logger:   logger,
		endpoint: endpoint,
	}
	return r, nil
}

func (r *RouteManger) Serve() error {
	server := grpc.NewServer()
	pb.RegisterRouteApiServer(server, r)
	reflection.Register(server)
	listener, err := net.Listen("tcp", r.endpoint)
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

func (r *RouteManger) Health(ctx context.Context, in *pb.HealthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (r *RouteManger) SetRoute(ctx context.Context, in *pb.SetRouteRequest) (*emptypb.Empty, error) {
	r.logger.Info("set route request to %s", in.Route.Destination)
	route, err := RouteFromReq(in.Route)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse request to route")
	}
	existingRoute, ok := r.routes[route.Dst.String()]
	if !ok {
		r.logger.Info("add: %s", route)
		if err := route.add(); err != nil {
			r.logger.Err("%s", err)
			return nil, status.Error(codes.Aborted, err.Error())
		}
		r.routes[route.Dst.String()] = route
		return &emptypb.Empty{}, nil
	}
	if route.Ad < existingRoute.Ad {
		r.logger.Info("replace: %s", route)
		if err := route.replace(); err != nil {
			r.logger.Err("%s", err)
			return nil, status.Error(codes.Aborted, err.Error())
		}
	}
	return &emptypb.Empty{}, nil
}

func RouteManagerHealthCheck() bool {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", DefaultRouteManagerHost, DefaultRouteManagerPort), grpc.WithInsecure())
	if err != nil {
		return false
	}
	defer conn.Close()
	client := pb.NewRouteApiClient(conn)
	if _, err := client.Health(context.Background(), &pb.HealthRequest{}); err != nil {
		return false
	}
	return true
}

func NewRouteManagerClient(endpoint string) (pb.RouteApiClient, error) {
	client, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewRouteApiClient(client), nil
}
