package main

import (
	"github.com/spf13/cobra"
)

// PlatformCmd is the main command for attaching topic subcommands.
var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Node operations",
	Long:  `Manage node-related operations.`,
	//Aliases: []string{"pf"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(NodeCmd)
}
