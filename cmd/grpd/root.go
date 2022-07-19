package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pkg/rip"
)

var rootCmd = &cobra.Command{
	Use:   "grpd [command]",
	Short: "GRP(Go Routing Protocol) daemon",
}

func init() {
	// rip
	ripCmd.Flags().StringSliceP("if", "i", []string{}, "Interfaces to handle RIP.")
	ripCmd.Flags().IntP("port", "p", 520, "RIP running port.")
	ripCmd.Flags().Uint64P("timeout", "t", rip.DEFALUT_TIMEOUT, "RIP timeout time.")
	ripCmd.Flags().Uint64P("gc", "g", rip.DEFALUT_GC_TIME, "RIP gc time.")
	ripCmd.Flags().IntP("log", "l", 0, "log level")
	ripCmd.Flags().StringP("log-path", "o", "", "log output path")

	// bgp
	bgpCmd.Flags().IntP("log", "l", 0, "log level")
	bgpCmd.Flags().StringP("log-path", "o", "", "log output path")
	bgpCmd.Flags().StringP("config", "c", "", "configuration file path")

	// route manager

	rootCmd.AddCommand(
		routeCmd,
		ripCmd,
		bgpCmd,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("GRPd Error\n\n%s", err)
		os.Exit(1)
	}
}
