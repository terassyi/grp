package bgp

import (
	"context"
	"fmt"
	"net"

	"github.com/terassyi/grp/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	port    int
	options []grpc.ServerOption
	pb.UnimplementedBgpApiServer
}

func newServer(port int) (*server, error) {
	return &server{
		port:    port,
		options: []grpc.ServerOption{},
	}, nil
}

func (s *server) serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer(s.options...)
	pb.RegisterBgpApiServer(grpcServer, s)
	return grpcServer.Serve(listener)
}

func (s *server) Health(ctx context.Context, in *pb.HealthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
