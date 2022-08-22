package main

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pkg/bgp"
	"github.com/terassyi/grp/pkg/config"
	grpLog "github.com/terassyi/grp/pkg/log"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "BGP-4(Border Gateway Protocol Version 4 (RFC 1771))",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatal(err)
		}
		level, err := cmd.Flags().GetInt("log")
		if err != nil {
			log.Fatal(err)
		}
		out, err := cmd.Flags().GetString("log-path")
		if err != nil {
			log.Fatal(err)
		}
		if file == "" {
			logLevel := 1
			logOut := grpLog.BGP_PATH
			if level != -1 {
				logLevel = level
			}
			if out != "" {
				logOut = out
			}
			server, err := bgp.NewServer(logLevel, logOut)
			if err != nil {
				log.Fatal(err)
			}
			if err := server.Run(context.Background()); err != nil {
				log.Fatal(err)
			}
			return
		}
		conf, err := config.Load(file)
		if err != nil {
			log.Fatal(err)
		}
		if conf.Bgp.Log == nil {
			conf.Bgp.Log = &grpLog.Log{
				Level: 1,
				Out:   grpLog.BGP_PATH,
			}
		}
		if level != -1 {
			conf.Bgp.Log.Level = level
		}
		if out != "" {
			conf.Bgp.Log.Out = out
		}
		server, err := bgp.NewServerWithConfig(conf.Bgp, conf.Bgp.Log.Level, conf.Bgp.Log.Out)
		if err != nil {
			log.Fatal(err)
		}
		if err := server.Run(context.Background()); err != nil {
			log.Fatal(err)
		}
	},
}
