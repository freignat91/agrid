package main

import (
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var FileStatCmd = &cobra.Command{
	Use:   "stat",
	Short: "stat of a file",
	Long:  `stat of file`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.fileStat(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileStatCmd)
	FileStatCmd.Flags().String("user", "", `set user name`)
}

func (m *agridCLI) fileStat(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		m.Fatal("Error: need file name as first argument\n")
	}
	fileName := args[0]
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	api.SetUser(cmd.Flag("user").Value.String())

	stat, err := api.FileStat(fileName)
	if err != nil {
		return err
	}
	m.pSuccess("File %s: length=%d\n", fileName, stat.Length)
	return nil
}
