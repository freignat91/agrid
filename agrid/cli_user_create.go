package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var UserCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create an user",
	Long:  `create an user`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.userCreate(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	UserCmd.AddCommand(UserCreateCmd)
	UserCreateCmd.Flags().String("token", "", `force user token`)
}

func (m *agridCLI) userCreate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Needs user name as first argument")
	}
	token := cmd.Flag("token").Value.String()
	user := args[0]
	m.pInfo("Execute: Create user %s\n", user)
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	token, err := api.UserCreate(user, token)
	if err != nil {
		return err
	}
	m.pSuccess("User create: %s:%s\n", user, token)
	return nil
}
