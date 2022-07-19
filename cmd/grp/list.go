package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pkg/constants"
	"github.com/terassyi/grp/pkg/route"
)

var healthCheckMap = map[string]func() bool{
	"bgp":           bgpHealthCheck,
	"route-manager": route.RouteManagerHealthCheck,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GRP Service list")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Service", "Endpoint", "Status"})
		for name, f := range healthCheckMap {
			table.Append([]string{name, fmt.Sprintf("localhost:%d", constants.ServiceApiServerMap[name]), fmt.Sprintf("%v", f())})
		}
		table.Render()
	},
}
