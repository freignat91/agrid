package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var UserRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove an user",
	Long:  `remove an user`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.userRemove(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	UserCmd.AddCommand(UserRemoveCmd)
}

func (m *agridCLI) userRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Needs user name as first argument")
	}
	user := args[0]
	m.pInfo("Execute: Remove user %s\n", user)
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	if err := api.UserRemove(user); err != nil {
		return err
	}
	m.pSuccess("User removed %s\n", user)
	return nil
}
