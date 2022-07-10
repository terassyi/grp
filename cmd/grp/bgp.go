package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pb"
	"github.com/terassyi/grp/pkg/constants"
	"google.golang.org/grpc"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "GRP BGP operating cli",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		as, err := strconv.Atoi(args[0])
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		bc := newBgpClient()
		defer bc.conn.Close()
		if _, err := bc.SetAS(context.Background(), &pb.SetASRequest{As: int32(as)}); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
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

var showSubCmd = &cobra.Command{
	Use:   "show",
	Short: "show bgp server information",
	Run: func(cmd *cobra.Command, args []string) {
		bc := newBgpClient()
		defer bc.conn.Close()
		res, err := bc.Show(context.Background(), &pb.ShowRequest{})
		if err != nil {
			os.Exit(1)
		}
		fmt.Println("BGP server information")
		fmt.Printf("  Running at %d\n", res.Port)
		fmt.Printf("  AS number %d\n", res.As)
		fmt.Printf("  Router id %s\n", res.RouterId)
	},
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
			log.Println(err)
			os.Exit(1)
		}
		fmt.Println("BGP Neighbors")
		for _, neighbor := range res.Neighbors {
			showNeighborInfo(neighbor)
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
			log.Println(err)
			os.Exit(1)
		}
		routerId, err := cmd.Flags().GetString("routerid")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		addr, err := cmd.Flags().GetString("address")
		if err != nil {
			log.Println(err)
			os.Exit(1)
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
			log.Println(err)
			os.Exit(1)
		}

		fmt.Println("BGP Neighbor")
		showNeighborInfo(res.Neighbor)
	},
}

func showNeighborInfo(info *pb.NeighborInfo) {
	fmt.Printf("  neighbor address %s\n", info.GetAddress())
	fmt.Printf("  remote AS %d\n", info.GetAs())
	fmt.Printf("  router id %s\n", info.GetRouterId())
}

var routerIdSubCmd = &cobra.Command{
	Use:   "router-id",
	Short: "set bgp router id",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		net.ParseIP(args[0])
		bc := newBgpClient()
		defer bc.conn.Close()
		if _, err := bc.RouterId(context.Background(), &pb.RouterIdRequest{RouterId: args[0]}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var remoteASSubCmd = &cobra.Command{
	Use:   "add",
	Short: "add remote AS(specify address and AS number)",
	Args:  cobra.MatchAll(cobra.MaximumNArgs(2), cobra.MaximumNArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			as   int
			addr string
		)
		for _, a := range args {
			if strings.Contains(a, ".") {
				addr = a
			} else {
				n, err := strconv.Atoi(a)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				as = n
			}
		}
		bc := newBgpClient()
		defer bc.conn.Close()
		if _, err := bc.RemoteAS(context.Background(), &pb.RemoteASRequest{
			As:   int32(as),
			Addr: addr,
		}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var networkSubCmd = &cobra.Command{
	Use:   "network",
	Short: "add network to advertise as a local originated route",
	Run: func(cmd *cobra.Command, args []string) {
		bc := newBgpClient()
		defer bc.conn.Close()
		if _, err := bc.Network(context.Background(), &pb.NetworkRequest{Networks: args}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
