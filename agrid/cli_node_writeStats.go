package main

import (
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeWriteStatsCmd = &cobra.Command{
	Use:   "writeStats",
	Short: "write stats in log file",
	Long:  `write stats in log file`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.writeStats(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodeWriteStatsCmd)
	NodeWriteStatsCmd.Flags().StringP("node", "n", "*", "Target a specific node")
}

func (m *ClientManager) writeStats(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: writeStats\n")
	node := cmd.Flag("node").Value.String()
	client, err := m.getClient()
	if err != nil {
		return err
	}
	if err := client.createSendMessageNoAnswer(node, "writeStatsInLog"); err != nil {
		return err
	}
	m.pSuccess("done\n")
	return nil
}
