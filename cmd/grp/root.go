package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/cmd/grp/bgp"
	"github.com/terassyi/grp/cmd/grp/rip"
)

var rootCmd = &cobra.Command{
	Use:   "grp <protocol> <command>",
	Short: "GRP(Go Routing Protocol) CLI Client",
}

func init() {
	// list command
	rootCmd.AddCommand(listCmd)

	rootCmd.AddCommand(
		rip.RipCmd,
		bgp.BgpCmd,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("GRP Error\n\n%s", err)
		os.Exit(1)
	}
}
