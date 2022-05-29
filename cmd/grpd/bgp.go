package main

import (
	"log"

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
		level, err := cmd.Flags().GetInt("log")
		if err != nil {
			log.Fatal(err)
		}
		out, err := cmd.Flags().GetString("log-path")
		if err != nil {
		}
		if file != "" {
			conf, err := config.Load(file)
			if err != nil {
				log.Fatal(err)
			}
			b, err := bgp.FromConfig(&conf.Bgp, conf.Level, conf.Out)
			if err != nil {
				log.Fatal(err)
			}
			if err := b.Poll(); err != nil {
				log.Fatal(err)
			}
		} else {
			b, err := bgp.New(bgp.PORT, level, out)
			if err != nil {
				log.Fatal(err)
			}
			if err := b.Poll(); err != nil {
				log.Fatal(err)
			}
		}
	},
}
