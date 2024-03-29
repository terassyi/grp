package route

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pb"
	"github.com/terassyi/grp/pkg/route"
)

var RouteCmd = &cobra.Command{
	Use:   "route",
	Short: "route information",
}

func init() {
	RouteCmd.AddCommand(
		showSubCmd,
	)
}

var showSubCmd = &cobra.Command{
	Use:   "show",
	Short: "show routes in route-manager",
	Run: func(cmd *cobra.Command, args []string) {
		nilStr := "<nil>"
		client, err := route.NewRouteManagerClient(fmt.Sprintf("%s:%d", route.DefaultRouteManagerHost, route.DefaultRouteManagerPort))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		res, err := client.ListRoute(context.Background(), &pb.ListRouteRequest{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Destination", "Gateway", "Source", "Device", "Protocol"})
		fmt.Println("  GRP Routing Information Base ")
		for _, route := range res.Route {
			dst := ""
			gw := ""
			src := ""

			if *route.Gw != nilStr {
				gw = *route.Gw
			}
			if route.Destination == nilStr {
				dst = "default"
			} else {
				dst = route.Destination
			}
			if *route.Src != nilStr {
				src = *route.Src
			}
			table.Append([]string{dst, gw, src, route.Link, route.Protocol.String()})
		}
		table.Render()
	},
}
