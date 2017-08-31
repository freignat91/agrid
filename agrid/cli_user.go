package main

import (
	"github.com/spf13/cobra"
)

// PlatformCmd is the main command for attaching topic subcommands.
var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "user operations",
	Long:  `Manage user-related operations.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(UserCmd)
}
