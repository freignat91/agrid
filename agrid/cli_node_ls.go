package main

import (
	"github.com/spf13/cobra"
	"sort"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "get the agrid nodes list",
	Long:  `get the agrid nodes list`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.getNodeList(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodeLsCmd)
}

func (m *ClientManager) getNodeList(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: getNodeList\n")
	t0 := time.Now()
	client, err := m.getClient()
	if err != nil {
		return err
	}
	_, errp := client.createSendMessage("*", false, "getConnections", "client")
	if errp != nil {
		return errp
	}
	rep := []string{}
	nb := 0
	for {
		mes, ok := client.getNextAnswer(100)
		if ok {
			nb++
			rep = append(rep, mes.Args[0])
		}
		if time.Now().Sub(t0) > time.Second*5 {
			break
		}
		if nb == client.nbNode {
			break
		}
	}
	sort.Strings(rep)
	for _, line := range rep {
		m.pSuccess("%s\n", line)
	}
	m.pSuccess("number=%d\n", len(rep))
	return nil
}
