package bgp

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
)

type server struct {
	bgp       *Bgp
	apiServer *grpc.Server
	pb.UnimplementedBgpApiServer
}

func NewServer(level int, out string) (*server, error) {
	b, err := New(PORT, level, out)
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()
	return &server{
		bgp:       b,
		apiServer: grpcServer,
	}, nil
}

func NewServerWithConfig(config *Config, logLevel int, logOut string) (*server, error) {
	b, err := FromConfig(config, logLevel, logOut)
	if err != nil {
		return nil, err
	}
	b.logger.Info("gRPC server is not created")
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	b.logger.Info("gRPC server created")
	return &server{
		bgp:       b,
		apiServer: grpcServer,
	}, nil
}

func (s *server) Run(ctx context.Context) error {
	s.bgp.logger.Info("GRP BGP Server Start")
	cctx, cancel := context.WithCancel(ctx)
	sigCh := make(chan os.Signal, 1)
	defer func() {
		cancel()
		s.bgp.logger.Info("BGP server stopped.")
	}()
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGTERM)
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", constants.ServiceApiServerMap["bgp"]))
	if err != nil {
		return err
	}
	pb.RegisterBgpApiServer(s.apiServer, s)
	go func() {
		if err := s.bgp.PollWithContext(cctx); err != nil {
			s.bgp.logger.Err("%v", err)
		}
	}()
	go func() {
		if err := s.apiServer.Serve(listener); err != nil {
			s.bgp.logger.Err("%v", err)
		}
	}()

	<-sigCh
	return nil
}
