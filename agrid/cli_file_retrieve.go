package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"os"
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
	FileRetrieveCmd.Flags().String("meta", "", "meta file name (default no metadata file)")
	FileRetrieveCmd.Flags().String("key", "", "AES key to encrypt file, 32 bybes")
	FileRetrieveCmd.Flags().String("user", "", `set user name`)
	FileRetrieveCmd.Flags().String("version", "0", `to retrieve a specific version, default last one`)
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
	version, err := strconv.Atoi(cmd.Flag("version").Value.String())
	if err != nil {
		m.Fatal("Error option --version is not a number: %s", cmd.Flag("thread").Value.String())
	}
	m.pInfo("Execute: retrieve file: %s to %s\n", clusterFile, localFile)
	key := cmd.Flag("key").Value.String()
	metaFileName := cmd.Flag("meta").Value.String()
	t0 := time.Now()

	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	api.SetUser(cmd.Flag("user").Value.String())
	metaMap, version, err := api.FileRetrieve(clusterFile, localFile, version, nbThread, key)
	if err != nil {
		return err
	}
	if metaFileName != "" {
		file, err := os.Create(metaFileName)
		if err != nil {
			return fmt.Errorf("Metadata file creation error: %v\n", err)
		}
		for key, val := range metaMap {
			if _, err := file.WriteString(fmt.Sprintf("%s=%s\n", key, val)); err != nil {
				return fmt.Errorf("Metadata file writing error: %v\n", err)
			}
		}
		file.Close()
	}
	m.pSuccess("file %s v%d received (%dms)\n", localFile, version, time.Now().Sub(t0).Nanoseconds()/1000000)
	return nil
}
