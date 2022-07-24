package rip

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/terassyi/grp/pb"
	"github.com/terassyi/grp/pkg/constants"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedRipApiServer
	rip       *Rip
	apiServer *grpc.Server
}

func NewServer(level int, out string) (*server, error) {
	r, err := New([]string{}, PORT, DEFALUT_TIMEOUT, DEFALUT_GC_TIME, level, out)
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()
	return &server{
		rip:       r,
		apiServer: grpcServer,
	}, nil
}

func NewServerWithConfig(config *Config, logLevel int, logOut string) (*server, error) {
	r, err := FromConfig(config, logLevel, logOut)
	if err != nil {
		return nil, err
	}
	r.logger.Info("gRPC server is not created")
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	r.logger.Info("gRPC server created")
	return &server{
		rip:       r,
		apiServer: grpcServer,
	}, nil
}

func (s *server) Run(ctx context.Context) error {
	s.rip.logger.Info("GRP RIPv1 Server Start")
	cctx, cancel := context.WithCancel(ctx)
	sigCh := make(chan os.Signal, 1)
	defer func() {
		cancel()
		s.rip.logger.Info("RIP server stopped.")
	}()
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGTERM)
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", constants.ServiceApiServerMap["rip"]))
	if err != nil {
		return err
	}
	pb.RegisterRipApiServer(s.apiServer, s)
	go func() {
		if err := s.rip.PollWithContext(cctx); err != nil {
			s.rip.logger.Err("%v", err)
		}
	}()
	go func() {
		if err := s.apiServer.Serve(listener); err != nil {
			s.rip.logger.Err("%v", err)
		}
	}()

	<-sigCh
	return nil
}

func (s *server) Health(ctx context.Context, in *pb.HealthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) GetVersion(ctx context.Context, in *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	return &pb.GetVersionResponse{Version: 1}, nil
}

func (s *server) GetLogPath(ctx context.Context, in *pb.GetLogPathRequest) (*pb.GetLogPathResponse, error) {
	return &pb.GetLogPathResponse{
		Level: int32(s.rip.logger.Level()),
		Path:  s.rip.logger.Path(),
	}, nil
}

func (s *server) Show(ctx context.Context, in *pb.RipShowRequest) (*pb.RipShowResponse, error) {
	ifnames := make([]string, 0, len(s.rip.links))
	for _, iface := range s.rip.links {
		ifnames = append(ifnames, iface.Attrs().Name)
	}
	return &pb.RipShowResponse{
		Timeout:   int32(s.rip.timeout),
		Gc:        int32(s.rip.gcTime),
		Interface: ifnames,
	}, nil
}

func (s *server) Network(ctx context.Context, in *pb.NetworkRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
