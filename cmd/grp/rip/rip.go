package rip

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	"github.com/terassyi/grp/pb"
	"github.com/terassyi/grp/pkg/constants"
	grpLog "github.com/terassyi/grp/pkg/log"
	"github.com/terassyi/grp/pkg/rip"
	"google.golang.org/grpc"
)

var RipCmd = &cobra.Command{
	Use:   "rip",
	Short: "GRP Rip operating cli",
}

func init() {
	// lgos
	logSubCmd.Flags().BoolP("follow", "f", false, "follow logs")
	logSubCmd.Flags().BoolP("plain-text", "p", false, "plain text format")
	RipCmd.AddCommand(
		healthSubCmd,
		showSubCmd,
		logSubCmd,
	)
}

type ripClient struct {
	pb.RipApiClient
	conn *grpc.ClientConn
}

func newRipClient() *ripClient {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", constants.ServiceApiServerMap["rip"]), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return &ripClient{
		RipApiClient: pb.NewRipApiClient(conn),
		conn:         conn,
	}
}

var healthSubCmd = &cobra.Command{
	Use:   "health",
	Short: "health check",
	Run: func(cmd *cobra.Command, args []string) {
		if rip.HealthCheck() {
			fmt.Println("ripd is healthy")
		} else {
			fmt.Println("ripd is unhealthy")
		}
	},
}

var showSubCmd = &cobra.Command{
	Use:   "show",
	Short: "show RIP server information",
	Run: func(cmd *cobra.Command, args []string) {
		rc := newRipClient()
		defer rc.conn.Close()
		res, err := rc.Show(context.Background(), &pb.RipShowRequest{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("RIP server information")
		fmt.Printf("  timeout %d\n", res.Timeout)
		fmt.Printf("  gc_time %d\n", res.Gc)
		fmt.Printf("  networks  ")
		for _, network := range res.Network {
			fmt.Printf("            %s\n", network)
		}
	},
}

var logSubCmd = &cobra.Command{
	Use:   "logs",
	Short: "show rip logs",
	Run: func(cmd *cobra.Command, args []string) {
		type LogJson struct {
			Time     string `json:"time"`
			Level    string `json:"level"`
			Protocol string `json:"protocol"`
			Message  string `json:"message"`
		}
		rc := newRipClient()
		defer rc.conn.Close()
		res, err := rc.GetLogPath(context.Background(), &pb.GetLogPathRequest{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if res.Path == "stdout" {
			fmt.Printf("Logs are output to standard output with level %s.\n", grpLog.Level(res.Level))
			os.Exit(0)
		}
		if res.Path == "stderr" {
			fmt.Printf("Logs are output to standard error(stderr) with level %s\n.", grpLog.Level(res.Level))
			os.Exit(0)
		}
		if res.Path == "" {
			fmt.Println("Logs are not output")
			os.Exit(0)
		}
		follow, err := cmd.Flags().GetBool("follow")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		plain, err := cmd.Flags().GetBool("plain-text")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("RIPv1 Logs Output %s with level %s\n\n", res.Path, grpLog.Level(res.Level))
		t, err := tail.TailFile(res.Path, tail.Config{
			ReOpen: true,
			Poll:   true,
			Follow: follow,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ctrlC := make(chan os.Signal, 1)
		signal.Notify(ctrlC,
			syscall.SIGINT,
			syscall.SIGTERM)
		formatPlainText := func(lj *LogJson) (string, error) {
			return fmt.Sprintf("%s|%s|%s|%s", lj.Time, lj.Level, lj.Protocol, lj.Message), nil
		}
		go func() {
			for line := range t.Lines {
				lj := &LogJson{}
				if err := json.Unmarshal([]byte(line.Text), lj); err != nil {
					fmt.Printf("parse json formated log:%v\n", err)
					continue
				}
				if plain {
					l, err := formatPlainText(lj)
					if err != nil {
						fmt.Println(err)
						continue
					}
					fmt.Println(l)
				} else {
					fmt.Println(line.Text)
				}
			}
		}()
		<-ctrlC
	},
}
