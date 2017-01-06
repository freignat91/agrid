package main

import (
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"strconv"
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
	FileRmCmd.Flags().String("user", "", `set user name`)
	FileRmCmd.Flags().String("version", "", `to remove a specific version only`)
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
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	api.SetUser(cmd.Flag("user").Value.String())
	version, err := strconv.Atoi(cmd.Flag("version").Value.String())
	if err != nil {
		m.Fatal("Error option --version is not a number: %s", cmd.Flag("version").Value.String())
	}

	errr := api.FileRm(fileName, version, recursive)
	t1 := time.Now()
	if errr != nil {
		return errr
	}
	m.pSuccess("File %s removed time=%dms\n", fileName, t1.Sub(t0).Nanoseconds()/1000000)
	return nil
}
