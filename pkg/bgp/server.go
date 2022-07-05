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
	signalCh chan os.Signal
}

func NewServer(config *Config, logLevel int, logOut string) (*server, error) {
	b, err := FromConfig(config, logLevel, logOut)
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	return &server{
		bgp:       b,
		apiServer: grpcServer,
	}, nil
}

func (s *server) Run(ctx context.Context) error {
	cctx, cancel := context.WithCancel(ctx)
	sigCh := make(chan os.Signal, 1)
	defer cancel()
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
			s.bgp.logger.Errorln(err)
		}
	}()
	go s.apiServer.Serve(listener)

	<-s.signalCh
	return nil
}
