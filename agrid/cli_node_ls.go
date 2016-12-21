package main

import (
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var NodeLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "get the agrid nodes list",
	Long:  `get the agrid nodes list`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.getNodeList(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	NodeCmd.AddCommand(NodeLsCmd)
}

func (m *agridCLI) getNodeList(cmd *cobra.Command, args []string) error {
	m.pInfo("Execute: getNodeList\n")

	api := agridapi.New(config.serverAddress)
	list, err := api.NodeLs()
	if err != nil {
		return err
	}
	for _, line := range list {
		m.pSuccess("%s\n", line)
	}
	m.pSuccess("number=%d\n", len(list))
	return nil
}
