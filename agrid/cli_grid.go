package main

import (
	"github.com/spf13/cobra"
)

// PlatformCmd is the main command for attaching topic subcommands.
var GridCmd = &cobra.Command{
	Use:   "grid",
	Short: "grid operations",
	Long:  `Manage grid-related operations.`,
	//Aliases: []string{"pf"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(GridCmd)
}
