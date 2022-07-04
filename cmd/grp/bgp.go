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

var healthSubCmd = &cobra.Command{
	Use:   "health",
	Short: "health check",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", constants.ServiceApiServerMap["bgp"]), grpc.WithInsecure())
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer conn.Close()
		client := pb.NewBgpApiClient(conn)
		if _, err := client.Health(context.Background(), &pb.HealthRequest{}); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		fmt.Println("healthy")
	},
}

func bgpHealthCheck() bool {
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
