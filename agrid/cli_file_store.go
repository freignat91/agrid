package main

import (
	"github.com/freignat91/agrid/agridapi"
	"github.com/freignat91/agrid/server/gnode"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var FileStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "store file",
	Long:  `store file on the cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.fileStore(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileStoreCmd)
	FileStoreCmd.Flags().Int("thread", 1, "send thread number")
	FileStoreCmd.Flags().String("meta", "", "metadata folowing the file format: name:value, name:value, ...")
	FileStoreCmd.Flags().String("key", "", "AES key to encrypt file, 32 bybes")
	FileStoreCmd.Flags().String("user", "", `set user name`)
}

func (m *agridCLI) fileStore(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		m.Fatal("Error: need file name as first argument\n")
	}
	fileName := args[0]
	meta := strings.Split(cmd.Flag("meta").Value.String(), ",")
	m.pInfo("Execute: store file: %s\n", fileName)
	targetedPath := fileName
	if len(args) >= 2 {
		targetedPath = args[1]
	}
	targetedPath = strings.Trim(targetedPath, "/")
	if strings.Index(targetedPath, gnode.GNodeFileSuffixe) >= 0 {
		m.Fatal("Invalid path: containing %s\n", gnode.GNodeFileSuffixe)
	}
	key := cmd.Flag("key").Value.String()
	nbThread, err := strconv.Atoi(cmd.Flag("thread").Value.String())
	if err != nil {
		m.Fatal("Error option --thread is not a number: %s", cmd.Flag("thread").Value.String())
	}
	t0 := time.Now()
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	api.SetUser(cmd.Flag("user").Value.String())
	version, err := api.FileStore(fileName, targetedPath, meta, nbThread, key)
	if err != nil {
		return err
	}
	m.pSuccess("file %s stored as %s v%d (%dms)\n", fileName, targetedPath, version, time.Now().Sub(t0).Nanoseconds()/1000000)
	return nil
}
