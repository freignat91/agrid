package main

import (
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeUpdateGridCmd = &cobra.Command{
	Use:   "updateGrid",
	Short: "update grid connections",
	Long:  `update grid connections`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.nodeUpdateGrid(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeUpdateGridCmd.Flags().StringP("node", "n", "*", "Target a specific node")
	NodeUpdateGridCmd.Flags().BoolP("force", "f", false, "Force to recompute")
	NodeCmd.AddCommand(NodeUpdateGridCmd)
}

func (m *agridCLI) nodeUpdateGrid(cmd *cobra.Command, args []string) error {
	node := "*"
	if len(args) >= 1 {
		node = args[0]
	}
	force := false
	if cmd.Flag("force").Value.String() == "true" {
		force = true
	}

	if len(args) >= 1 {
		node = args[0]
	}
	m.pInfo("Execute: update grid\n")
	api := agridapi.New(config.serverAddress)
	if err := api.NodeUpdateGrid(node, force); err != nil {
		return err
	}
	m.pSuccess("Grid updated for node %s\n", node)
	return nil
}
