package main

import (
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeSetLogLevelCmd = &cobra.Command{
	Use:   "setLogLevel level [nodeName]",
	Short: "setLogLevel ERROR/WARN/INFO/DEBUG",
	Long:  `setLogLevel ERROR/WARN/INFO/DEBUG`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.setLogLevel(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeSetLogLevelCmd.Flags().StringP("node", "n", "*", "Target a specific node")
	NodeCmd.AddCommand(NodeSetLogLevelCmd)
}

func (m *ClientManager) setLogLevel(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: setLogLevel %s\n", args[0])
	node := cmd.Flag("node").Value.String()
	client, err := m.getClient()
	if err != nil {
		return err
	}
	if err := client.createSendMessageNoAnswer(node, "setLogLevel", args[0]); err != nil {
		return err
	}
	return nil
}
