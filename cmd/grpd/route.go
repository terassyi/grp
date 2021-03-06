package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pkg/route"
)

var routeCmd = &cobra.Command{
	Use:   "route-manager",
	Short: "run Grp route manager",
	Run: func(cmd *cobra.Command, args []string) {
		routeServer, err := route.New(route.DefaultRouteManagerHost, route.DefaultRouteManagerPort)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := routeServer.Serve(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
