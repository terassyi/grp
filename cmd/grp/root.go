package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "grp <protocol> <command>",
	Short: "GRP(Go Routing Protocol) CLI Client",
}

func init() {
	// list command
	rootCmd.AddCommand(listCmd)

	// bgp command
	// neighbor sub command
	getNeighborSubCmd.Flags().IntP("as", "a", 0, "AS Number of the neighbor")
	getNeighborSubCmd.Flags().StringP("routerid", "r", "", "AS Number of the neighbor")
	getNeighborSubCmd.Flags().StringP("address", "d", "", "Peer IP Address")
	neighborSubCmd.AddCommand(
		getNeighborSubCmd,
		listNeighborSubCmd,
		remoteASSubCmd,
	)
	logSubCmd.Flags().BoolP("follow", "f", false, "follow logs")
	logSubCmd.Flags().BoolP("p", "plain-text", false, "plain text format")
	bgpCmd.AddCommand(
		logSubCmd,
		healthSubCmd,
		showSubCmd,
		neighborSubCmd,
		routerIdSubCmd,
		networkSubCmd,
	)
	rootCmd.AddCommand(bgpCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("GRP Error\n\n%s", err)
		os.Exit(1)
	}
}
