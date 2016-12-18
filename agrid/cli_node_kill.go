package main

import (
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeKillCmd = &cobra.Command{
	Use:   "kill",
	Short: "kill an agrid node",
	Long:  `kill an agrid node`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.kill(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodeKillCmd)
}

func (m *ClientManager) kill(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: kill node %s\n", args[0])
	client, err := m.getClient()
	if err != nil {
		return err
	}
	mes, err := client.createSendMessage(args[0], true, "killNode")
	if err != nil {
		return err
	}
	m.pSuccess("Container killed: %s\n", mes.Args[0])
	return nil
}
