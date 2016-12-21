package main

import (
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var FileRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove file",
	Long:  `remove file`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.fileRemove(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileRmCmd)
	FileRmCmd.Flags().BoolP("recursive", "r", false, `remomve all files under a directory`)
}

func (m *agridCLI) fileRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		m.Fatal("Error: need file name as first argument\n")
	}
	fileName := args[0]
	srecursive := cmd.Flag("recursive").Value.String()
	recursive := false
	if srecursive == "true" {
		recursive = true
	}
	t0 := time.Now()
	api := agridapi.New(config.serverAddress)

	err, done := api.FileRm(fileName, recursive)
	t1 := time.Now()
	if err != nil {
		return err
	} else if done {
		m.pSuccess("File %s removed time=%dms\n", fileName, t1.Sub(t0).Nanoseconds()/1000000)
	} else {
		m.pWarn("File %s not found time=%dms\n", fileName, t1.Sub(t0).Nanoseconds()/1000000)
	}
	return nil
}
