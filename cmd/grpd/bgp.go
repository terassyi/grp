package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pkg/bgp"
	"github.com/terassyi/grp/pkg/config"
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
		// level, err := cmd.Flags().GetInt("log")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// out, err := cmd.Flags().GetString("log-path")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		if file == "" {
			fmt.Println("please specify config file.")
			os.Exit(1)
		}
		conf, err := config.Load(file)
		if err != nil {
			log.Fatal(err)
		}
		server, err := bgp.NewServer(&conf.Bgp, conf.Level, conf.Out)
		if err != nil {
			log.Fatal(err)
		}
		if err := server.Run(context.Background()); err != nil {
			log.Fatal(err)
		}
	},
}
