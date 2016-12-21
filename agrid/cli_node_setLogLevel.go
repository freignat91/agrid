package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeSetLogLevelCmd = &cobra.Command{
	Use:   "setLogLevel level [nodeName]",
	Short: "setLogLevel ERROR/WARN/INFO/DEBUG",
	Long:  `setLogLevel ERROR/WARN/INFO/DEBUG`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.setLogLevel(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeSetLogLevelCmd.Flags().StringP("node", "n", "*", "Target a specific node")
	NodeCmd.AddCommand(NodeSetLogLevelCmd)
}

func (m *agridCLI) setLogLevel(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Needs log level as first argument (error | warn | info | debug")
	}
	m.pInfo("Execute: setLogLevel %s\n", args[0])
	node := cmd.Flag("node").Value.String()
	api := agridapi.New(config.serverAddress)
	if err := api.NodeSetLogLevel(node, args[0]); err != nil {
		return err
	}
	if node == "*" {
		m.pSuccess("Log level set to %s for all nodes\n", args[0])
	} else {
		m.pSuccess("Log level set to %s for node %s\n", args[0], node)
	}
	return nil
}
