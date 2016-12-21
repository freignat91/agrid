package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodePingCmd = &cobra.Command{
	Use:   "ping",
	Short: "ping  an agrid node",
	Long:  `ping an agrid node`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.ping(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodePingCmd)
}

func (m *agridCLI) ping(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Needs node name as first argument")
	}
	node := args[0]
	m.pInfo("Execute: ping %s\n", node)
	t0 := time.Now()
	api := agridapi.New(config.serverAddress)
	path, err := api.NodePing(node, false)
	if err != nil {
		return err
	}
	t1 := time.Now()
	m.pSuccess("Ping time=%d ms path: %s\n", t1.Sub(t0).Nanoseconds()/1000000, path)
	return nil
}
