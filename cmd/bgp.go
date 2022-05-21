package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/terassyi/grp/grp/bgp"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "BGP-4(Border Gateway Protocol Version 4 (RFC 1771))",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		level, err := cmd.Flags().GetInt("log")
		if err != nil {
			log.Println(err)
		}
		out, err := cmd.Flags().GetString("log-path")
		if err != nil {
			log.Println(err)
		}
		b, err := bgp.New(bgp.PORT, level, out)
		if err != nil {
			log.Println(err)
		}
		if err := b.Poll(); err != nil {
			log.Println(err)
			os.Exit(-1)
		}
	},
}
