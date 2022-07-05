package bgp

import (
	"context"

	"github.com/terassyi/grp/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type apiServer struct {
	port    int
	options []grpc.ServerOption
}

func newApiServer(port int) (*apiServer, error) {
	return &apiServer{
		port:    port,
		options: []grpc.ServerOption{},
	}, nil
}

func (s *server) Health(ctx context.Context, in *pb.HealthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) GetNeighbor(ctx context.Context, in *pb.GetNeighborRequest) (*pb.GetNeighborResponse, error) {
	return nil, nil
}

func (s *server) ListNeighbor(ctx context.Context, in *pb.ListNeighborRequest) (*pb.ListNeighborResponse, error) {
	return nil, nil
}
