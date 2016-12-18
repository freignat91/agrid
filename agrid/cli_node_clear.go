package main

import (
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear an agrid node",
	Long:  `clear an agrid node`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.clear(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodeClearCmd)
}

func (m *ClientManager) clear(cmd *cobra.Command, args []string) error {
	node := "*"
	if len(args) >= 1 {
		node = args[0]
	}
	m.pInfo("Execute: clear node %s\n", node)
	client, err := m.getClient()
	if err != nil {
		return err
	}
	_, errc := client.createSendMessage(node, true, "clear")
	if errc != nil {
		return errc
	}
	m.pSuccess("done\n")
	return nil
}
