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
	bgpCmd.AddCommand(healthSubCmd)
	rootCmd.AddCommand(bgpCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("GRP Error\n\n%s", err)
		os.Exit(1)
	}
}
