package main

import (
	"github.com/spf13/cobra"
)

// PlatformCmd is the main command for attaching topic subcommands.
var FileCmd = &cobra.Command{
	Use:   "file",
	Short: "file operations",
	Long:  `Manage file-related operations.`,
	//Aliases: []string{"pf"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(FileCmd)
}
