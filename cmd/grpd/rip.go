package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pkg/rip"
)

var ripCmd = &cobra.Command{
	Use:   "rip",
	Short: "RIP(Routing Information Protocol)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		links, err := cmd.Flags().GetStringSlice("if")
		if err != nil {
			fmt.Println(err)
		}
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			fmt.Println(err)
		}
		timeout, err := cmd.Flags().GetUint64("timeout")
		if err != nil {
			fmt.Println(err)
		}
		gcTime, err := cmd.Flags().GetUint64("gc")
		if err != nil {
			fmt.Println(err)
		}
		level, err := cmd.Flags().GetInt("log")
		if err != nil {
			fmt.Println(err)
		}
		out, err := cmd.Flags().GetString("log-path")
		if err != nil {
			fmt.Println(err)
		}
		r, err := rip.New(links, port, timeout, gcTime, level, out)
		if err != nil {
			fmt.Println(err)
		}
		if err := r.Poll(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	},
}
