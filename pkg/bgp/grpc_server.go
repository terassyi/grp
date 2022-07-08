package bgp

import (
	"context"
	"errors"

	"github.com/terassyi/grp/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrNeighborNotFound error = errors.New("Neighbor is not found")
)

func (s *server) Health(ctx context.Context, in *pb.HealthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) GetNeighbor(ctx context.Context, in *pb.GetNeighborRequest) (*pb.GetNeighborResponse, error) {
	var peer *peer
	if in.As != 0 {
		for _, p := range s.bgp.peers {
			if p.as == int(in.As) {
				peer = p
			}
		}
	}
	if peer == nil && in.PeerAddress != nil {
		peer = s.bgp.peers[*in.PeerAddress]
	}
	if peer == nil && in.RouterId != nil {
		for _, p := range s.bgp.peers {
			if p.routerId.String() == *in.RouterId {
				peer = p
			}
		}
	}
	if peer == nil {
		return nil, ErrNeighborNotFound
	}
	return &pb.GetNeighborResponse{
		Neighbor: &pb.NeighborInfo{
			As:       uint32(peer.neighbor.as),
			Address:  peer.neighbor.addr.String(),
			Port:     uint32(peer.neighbor.port),
			RouterId: peer.neighbor.addr.String(),
		},
	}, nil
}

func (s *server) ListNeighbor(ctx context.Context, in *pb.ListNeighborRequest) (*pb.ListNeighborResponse, error) {
	neighbors := make([]*pb.NeighborInfo, 0)
	for _, p := range s.bgp.peers {
		info := &pb.NeighborInfo{
			As:       uint32(p.neighbor.as),
			Address:  p.neighbor.addr.String(),
			Port:     uint32(p.neighbor.port),
			RouterId: p.neighbor.addr.String(),
		}
		neighbors = append(neighbors, info)
	}
	return &pb.ListNeighborResponse{
		Neighbors: neighbors,
	}, nil
}

func (s *server) Network(ctx context.Context, in *pb.NetworkRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
