package main

import (
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeUpdateGridCmd = &cobra.Command{
	Use:   "updateGrid",
	Short: "update grid connections",
	Long:  `update grid connections`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.nodeUpdateGrid(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeUpdateGridCmd.Flags().StringP("node", "n", "*", "Target a specific node")
	NodeUpdateGridCmd.Flags().BoolP("force", "f", false, "Force to recompute")
	NodeCmd.AddCommand(NodeUpdateGridCmd)
}

func (m *ClientManager) nodeUpdateGrid(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: computeGrid\n")
	node := cmd.Flag("node").Value.String()
	force := cmd.Flag("force").Value.String()
	client, err := m.getClient()
	if err != nil {
		return err
	}
	if err := client.createSendMessageNoAnswer(node, "updateGrid", force); err != nil {
		return err
	}
	return nil
}
