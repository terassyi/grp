package main

import (
	"context"
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
		if file == "" {
			server, err := rip.NewServer(level, out)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if err := server.Run(context.Background()); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return
		}
		config, err := config.Load(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if links != nil {
			config.Rip.Interfaces = links
		}
		if port != 0 {
			config.Rip.Port = port
		}
		if timeout != 0 {
			config.Rip.Timeout = int(timeout)
		}
		if gcTime != 0 {
			config.Rip.Gc = int(gcTime)
		}
		server, err := rip.NewServerWithConfig(config.Rip, config.Level, config.Out)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := server.Run(context.Background()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
