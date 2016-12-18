package main

import (
	"github.com/freignat91/agrid/server/gnode"
	"github.com/spf13/cobra"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeVerifTraceCmd = &cobra.Command{
	Use:   "verifTrace",
	Short: "VerifTrace",
	Long:  `verifTrace`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.verifTrace(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodeVerifTraceCmd)
}

func (m *ClientManager) verifTrace(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: vertifTrace\n")
	client, err := m.getClient()
	if err != nil {
		return err
	}
	dec := "2"
	if len(args) >= 1 {
		dec = args[0]
	}
	m.pSuccess("From node %s (%s)\n", client.nodeName, client.nodeHost)
	target := ""
	if ret, err := client.createSendMessage("", true, "getNodeName", dec); err != nil {
		return err
	} else {
		target = ret.Args[0]
	}
	m.pSuccess("To node %s\n", target)
	t0 := time.Now()
	mes1 := gnode.CreateMessage(target, true, "ping", "client")
	if m.debug {
		mes1.Debug = true
	}
	ret1, err1 := client.sendMessage(mes1, true)
	if err1 != nil {
		return err1
	}
	t1 := time.Now()
	ret1.Path = append(ret1.Path, target)
	m.pSuccess("Answer1: %s\n", ret1.Args[0])
	m.pSuccess("Ping1 time=%d ms path: %v\n", t1.Sub(t0).Nanoseconds()/1000000, ret1.Path)

	t0 = time.Now()
	mes2 := gnode.CreateMessage(target, true, "ping", "client")
	if m.debug {
		mes2.Debug = true
	}
	ret2, err2 := client.sendMessage(mes2, true)
	if err2 != nil {
		return err2
	}
	t1 = time.Now()
	ret2.Path = append(ret2.Path, target)
	m.pSuccess("Answer2: %s\n", ret2.Args[0])
	m.pSuccess("Ping2 time=%d ms path: %v\n", t1.Sub(t0).Nanoseconds()/1000000, ret2.Path)
	return nil
}
