package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pkg/config"
	"github.com/terassyi/grp/pkg/rip"
)

var ripCmd = &cobra.Command{
	Use:   "rip",
	Short: "RIP(Routing Information Protocol)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, err := cmd.Flags().GetString("config")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		links, err := cmd.Flags().GetStringSlice("if")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		timeout, err := cmd.Flags().GetUint64("timeout")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		gcTime, err := cmd.Flags().GetUint64("gc")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		level, err := cmd.Flags().GetInt("log")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		out, err := cmd.Flags().GetString("log-path")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var r *rip.Rip
		if file == "" {
			r, err = rip.New(links, port, timeout, gcTime, level, out)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		config, err := config.Load(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		r, err = rip.FromConfig(config.Rip, level, out)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		if err := r.Poll(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	},
}
