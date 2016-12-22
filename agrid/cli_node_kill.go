package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeKillCmd = &cobra.Command{
	Use:   "kill",
	Short: "kill an agrid node",
	Long:  `kill an agrid node`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.kill(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodeKillCmd)
}

func (m *agridCLI) kill(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Needs node name as first argument")
	}
	node := args[0]
	m.pInfo("Execute: kill node %s\n", node)
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	if err := api.NodeKill(node); err != nil {
		return err
	}
	m.pSuccess("Container killed node %s\n", node)
	return nil
}
