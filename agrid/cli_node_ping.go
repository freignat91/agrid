package main

import (
	"github.com/spf13/cobra"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodePingCmd = &cobra.Command{
	Use:   "ping",
	Short: "ping  an agrid node",
	Long:  `ping an agrid node`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.ping(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodePingCmd)
}

func (m *ClientManager) ping(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: ping %s\n", args[0])
	t0 := time.Now()
	client, err := m.getClient()
	if err != nil {
		return err
	}
	mes, err := client.createSendMessage(args[0], true, "ping", "client")
	if err != nil {
		return err
	}
	t1 := time.Now()
	mes.Path = append(mes.Path, args[0])
	m.pSuccess("Ping time=%d ms path: %v\n", t1.Sub(t0).Nanoseconds()/1000000, mes.Path)
	return nil
}
