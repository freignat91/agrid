package main

import (
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var FileLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list file",
	Long:  `list file`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.fileList(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileLsCmd)
	FileLsCmd.Flags().String("user", "", `set user name`)
	FileLsCmd.Flags().Bool("version", false, `if exist display all file versions`)
}

func (m *agridCLI) fileList(cmd *cobra.Command, args []string) error {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	api.SetUser(cmd.Flag("user").Value.String())
	version := false
	if cmd.Flag("version").Value.String() == "true" {
		version = true
	}
	lineList, err := api.FileLs(path, version)
	if err != nil {
		m.Fatal("%v\n", err)
	}
	for _, line := range lineList {
		m.pSuccess("%s\n", line)
	}
	return nil
}
