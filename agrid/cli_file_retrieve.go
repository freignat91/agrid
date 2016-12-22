package main

import (
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var FileRetrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "retrieve file",
	Long:  `retrieve file from the cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.fileRetrieve(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileRetrieveCmd)
	FileRetrieveCmd.Flags().Int("thread", 1, "send thread number")
	FileRetrieveCmd.Flags().String("meta", "", "metadata folowing the file format: name:value, name:value, ...")
	FileRetrieveCmd.Flags().String("key", "", "AES key to encrypt file, 32 bybes")
}

func (m *agridCLI) fileRetrieve(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		m.Fatal("Error missing arguments: usage: agrid file retrieve [cluster file] [local file]\n")
	}
	clusterFile := args[0]
	localFile := args[1]
	nbThread, err := strconv.Atoi(cmd.Flag("thread").Value.String())
	if err != nil {
		m.Fatal("Error option --thread is not a number: %s", cmd.Flag("thread").Value.String())
	}
	m.pInfo("Execute: retrieve file: %s to %s\n", clusterFile, localFile)
	key := cmd.Flag("key").Value.String()
	t0 := time.Now()

	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	if err := api.FileRetrieve(clusterFile, localFile, nbThread, key); err != nil {
		return err
	}
	m.pSuccess("file %s received (%dms)\n", localFile, time.Now().Sub(t0).Nanoseconds()/1000000)
	return nil
}
