package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var FileGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get file",
	Long:  `get file from the cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.fileGet(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileGetCmd)
	FileGetCmd.Flags().String("meta", "", "metadata folowing the file format: name:value, name:value, ...")
	FileGetCmd.Flags().String("key", "", "AES key to encrypt file, 32 bybes")
}

func (m *ClientManager) fileGet(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		m.Fatal("Error missing arguments: usage: get [cluster file] [local file]\n")
	}
	clusterFile := args[0]
	localFile := args[1]
	m.pInfo("Execute: get file: %s to %d\n", clusterFile, localFile)
	key := cmd.Flag("key").Value.String()
	if key != "" {
		for len(key) < 32 {
			key = fmt.Sprintf("%s%s", key, key)
		}
		key = key[0:32]
	}
	t0 := time.Now()
	fileManager := fileManager{}
	fileManager.init(m)
	if err := fileManager.get(clusterFile, localFile, key); err != nil {
		return err
	}
	m.pSuccess("file %s received (%dms)\n", localFile, time.Now().Sub(t0).Nanoseconds()/1000000)
	return nil
}
