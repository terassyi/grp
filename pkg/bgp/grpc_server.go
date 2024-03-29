package bgp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/terassyi/grp/pb"
	"github.com/terassyi/grp/pkg/constants"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrNeighborNotFound error = errors.New("Neighbor is not found")
)

type Request struct {
	code requestCode
	req  gRPCRequest
}

type requestCode uint8

const (
	requestSetAS       = iota
	requestSetRouterId = iota
	requestAddNeighbor = iota
	requestAddNetwork  = iota
)

type gRPCRequest interface {
	Reset()
	String() string
	ProtoMessage()
}

func (s *server) Health(ctx context.Context, in *pb.HealthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) GetLogPath(ctx context.Context, in *pb.GetLogPathRequest) (*pb.GetLogPathResponse, error) {
	return &pb.GetLogPathResponse{
		Level: int32(s.bgp.logger.Level()),
		Path:  s.bgp.logger.Path(),
	}, nil
}

func (s *server) Show(ctx context.Context, in *pb.BgpShowRequest) (*pb.BgpShowResponse, error) {
	return &pb.BgpShowResponse{
		As:       int32(s.bgp.as),
		Port:     int32(s.bgp.port),
		RouterId: s.bgp.routerId.String(),
	}, nil
}

func (s *server) ShowRoute(ctx context.Context, in *pb.BgpShowRouteRequest) (*pb.BgpShowRouteResponse, error) {
	pathes := s.bgp.adjRibIn.GetAll()
	routes := make([]*pb.BgpRoute, 0, len(pathes))
	for _, p := range pathes {
		var nexhop string
		var reason string
		if p.nextHop == nil || p.nextHop.Equal(net.ParseIP("0.0.0.0")) {
			nexhop = "self"
		} else {
			nexhop = p.nextHop.String()
		}
		asPath := make([]int32, 0)
		for _, a := range p.asPath.GetSequence() {
			asPath = append(asPath, int32(a))
		}
		if p.best {
			reason = p.reason.String()
		} else {
			reason = ""
		}
		routes = append(routes, &pb.BgpRoute{
			Network:   p.nlri.String(),
			Nexthop:   nexhop,
			Metric:    0,
			LocalPref: int32(p.localPref),
			Path:      asPath,
			Best:      p.best,
			Reason:    reason,
		})
	}
	return &pb.BgpShowRouteResponse{Routes: routes}, nil
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

func (s *server) SetAS(ctx context.Context, in *pb.SetASRequest) (*emptypb.Empty, error) {
	if in.As == 0 {
		return &emptypb.Empty{}, errors.New("Invalid AS Number")
	}
	if s.bgp.as != 0 {
		return &emptypb.Empty{}, errors.New("AS Number is already set")
	}
	s.bgp.requestQueue <- &Request{
		code: requestSetAS,
		req:  in,
	}
	return &emptypb.Empty{}, nil
}

func (s *server) RouterId(ctx context.Context, in *pb.RouterIdRequest) (*emptypb.Empty, error) {
	if !strings.Contains(in.RouterId, ".") {
		return &emptypb.Empty{}, errors.New("Invalid router id format")
	}
	split := strings.Split(in.RouterId, ".")
	if len(split) != 4 {
		return &emptypb.Empty{}, errors.New("Invalid router id format")
	}
	s.bgp.requestQueue <- &Request{
		code: requestSetRouterId,
		req:  in,
	}
	return &emptypb.Empty{}, nil
}

func (s *server) RemoteAS(ctx context.Context, in *pb.RemoteASRequest) (*emptypb.Empty, error) {
	addr := net.ParseIP(in.Addr)
	_, _, err := lookupLocalAddr(addr)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	_, ok := s.bgp.peers[in.Addr]
	if ok {
		return &emptypb.Empty{}, fmt.Errorf("Peer(addr=%s) is already registered", in.Addr)
	}
	_, ook := s.bgp.LookupPeerWithAS(int(in.As))
	if ook {
		return &emptypb.Empty{}, fmt.Errorf("Peer(AS=%d) is already registered", in.As)
	}
	s.bgp.requestQueue <- &Request{
		code: requestAddNeighbor,
		req:  in,
	}
	return &emptypb.Empty{}, nil
}

func (s *server) Network(ctx context.Context, in *pb.NetworkRequest) (*emptypb.Empty, error) {
	for _, network := range in.Networks {
		_, _, err := net.ParseCIDR(network)
		if err != nil {
			return &emptypb.Empty{}, err
		}
	}
	s.bgp.requestQueue <- &Request{
		code: requestAddNetwork,
		req:  in,
	}
	return &emptypb.Empty{}, nil
}

func HealthCheck() bool {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", constants.ServiceApiServerMap["bgp"]), grpc.WithInsecure())
	if err != nil {
		return false
	}
	defer conn.Close()
	client := pb.NewBgpApiClient(conn)
	if _, err := client.Health(context.Background(), &pb.HealthRequest{}); err != nil {
		return false
	}
	return true
}
