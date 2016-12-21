package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeVerifTraceCmd = &cobra.Command{
	Use:   "verifTrace",
	Short: "VerifTrace",
	Long:  `verifTrace`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.verifTrace(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodeVerifTraceCmd)
}

func (m *agridCLI) verifTrace(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: vertifTrace\n")
	if len(args) < 2 {
		return fmt.Errorf("Needs two arguements: sender node name and targeted node name")
	}
	t0 := time.Now()
	api := agridapi.New(config.serverAddress)
	path, err := api.NodePingFrom(args[0], args[1], m.debug)
	if err != nil {
		return err
	}
	t1 := time.Now()
	m.pSuccess("Ping %s -> %s time=%d ms path: %v\n", t1.Sub(t0).Nanoseconds()/1000000, args[0], args[1], path)
	return nil
}
