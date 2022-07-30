package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pkg/config"
	"github.com/terassyi/grp/pkg/log"
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
		conf, err := config.Load(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if port != 0 {
			conf.Rip.Port = port
		}
		if timeout != 0 {
			conf.Rip.Timeout = int(timeout)
		}
		if gcTime != 0 {
			conf.Rip.Gc = int(gcTime)
		}
		if conf.Rip.Log == nil {
			conf.Rip.Log = &log.Log{
				Level: 1,
				Out:   "/var/log/grp/rip",
			}
		}
		if level != 0 {
			conf.Rip.Log.Level = level
		}
		if out != "" {
			conf.Rip.Log.Out = out
		}
		server, err := rip.NewServerWithConfig(conf.Rip, conf.Rip.Log.Level, conf.Rip.Log.Out)
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
