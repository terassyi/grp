package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

var neighborSubCmd = &cobra.Command{
	Use:   "neighbor",
	Short: "neighbor operationg commands",
}

var listNeighborSubCmd = &cobra.Command{
	Use:   "list neighbors",
	Short: "list up registered neighbors",
	Run: func(cmd *cobra.Command, args []string) {
		bc := newBgpClient()
		defer bc.conn.Close()
		res, err := bc.ListNeighbor(context.Background(), &pb.ListNeighborRequest{})
		if err != nil {
			panic(err)
		}
		fmt.Println("BGP Neighbors")
		for _, neighbor := range res.Neighbors {
			showNeighborInfor(neighbor)
			fmt.Println()
		}
	},
}

var getNeighborSubCmd = &cobra.Command{
	Use:   "get neighbor",
	Short: "get neighbor",
	Run: func(cmd *cobra.Command, args []string) {
		asn, err := cmd.Flags().GetInt("as")
		if err != nil {
			log.Fatal(err)
		}
		routerId, err := cmd.Flags().GetString("routerid")
		if err != nil {
			log.Fatal(err)
		}
		addr, err := cmd.Flags().GetString("address")
		if err != nil {
			log.Fatal(err)
		}
		if asn == 0 && routerId == "" && addr == "" {
			fmt.Println("Please specify at least an identifier(AS or router id or address)")
			os.Exit(1)
		}
		bc := newBgpClient()
		defer bc.conn.Close()
		res, err := bc.GetNeighbor(context.Background(), &pb.GetNeighborRequest{
			As:          uint32(asn),
			RouterId:    &routerId,
			PeerAddress: &addr,
		})
		if err != nil {
			panic(err)
		}

		fmt.Println("BGP Neighbor")
		showNeighborInfor(res.Neighbor)
	},
}

func showNeighborInfor(info *pb.NeighborInfo) {
	fmt.Printf("  neighbor address %s\n", info.GetAddress())
	fmt.Printf("  remote AS %d\n", info.GetAs())
	fmt.Printf("  router id %s\n", info.GetRouterId())
}
