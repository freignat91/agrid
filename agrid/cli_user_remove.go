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
	UserRemoveCmd.Flags().Bool("force", false, `WARNING: force to removce user with its associated files`)
}

func (m *agridCLI) userRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Error number of argument, needs [user] format userName:token")
	}
	user := args[0]
	force := false
	if cmd.Flag("force").Value.String() == "true" {
		force = true
	}
	m.pInfo("Execute: Remove user %s\n", user)
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	if err := api.UserRemove(user, force); err != nil {
		return err
	}
	m.pSuccess("User removed %s\n", user)
	return nil
}
