package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"strconv"
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
	FileStatCmd.Flags().String("version", "0", `internal file version, if 0, search the last one`)
}

func (m *agridCLI) fileStat(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		m.Fatal("Error: need file name as first argument\n")
	}
	version, err := strconv.Atoi(cmd.Flag("version").Value.String())
	if err != nil {
		m.Fatal("Error option --version is not a number: %s", cmd.Flag("version").Value.String())
	}
	fileName := args[0]
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	api.SetUser(cmd.Flag("user").Value.String())

	stat, exist, err := api.FileStat(fileName, version)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("File %s doesn't exist", fileName)
	}
	m.pSuccess("File %s: version=%d length=%d\n", fileName, stat.Version, stat.Length)
	return nil
}
