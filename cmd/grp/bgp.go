package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pb"
	"github.com/terassyi/grp/pkg/constants"
	"google.golang.org/grpc"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "GRP BGP operating cli",
}

type bgpClient struct {
	pb.BgpApiClient
	conn *grpc.ClientConn
}

func newBgpClient() *bgpClient {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", constants.ServiceApiServerMap["bgp"]), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return &bgpClient{
		BgpApiClient: pb.NewBgpApiClient(conn),
		conn:         conn,
	}
}

var healthSubCmd = &cobra.Command{
	Use:   "health",
	Short: "health check",
	Run: func(cmd *cobra.Command, args []string) {
		if bgpHealthCheck() {
			fmt.Println("bgpd is healthy")
		} else {
			fmt.Println("bgpd is unhealthy")
		}
	},
}

func bgpHealthCheck() bool {
	bc := newBgpClient()
	defer bc.conn.Close()
	if _, err := bc.Health(context.Background(), &pb.HealthRequest{}); err != nil {
		return false
	}
	return true
}

var listNeighborSubCmd = &cobra.Command{
	Use:   "list neighbors",
	Short: "list up registered neighbors",
	Run: func(cmd *cobra.Command, args []string) {
		bc := newBgpClient()
		defer bc.conn.Close()
		_, err := bc.ListNeighbor(context.Background(), &pb.ListNeighborRequest{})
		if err != nil {
			panic(err)
		}
		fmt.Println("hoge")
	},
}
