package main

import (
	"github.com/spf13/cobra"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var FileRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove file",
	Long:  `remove file`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.fileRemove(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileRmCmd)
	FileRmCmd.Flags().BoolP("recursive", "r", false, `remomve all files under a directory`)
}

func (m *ClientManager) fileRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		m.Fatal("Error: need file name as first argument\n")
	}
	fileName := args[0]
	t0 := time.Now()
	client, err := m.getClient()
	if err != nil {
		return err
	}
	recursive := cmd.Flag("recursive").Value.String()
	if _, err := client.createSendMessage("*", false, "removeFile", fileName, recursive); err != nil {
		return err
	}
	nbOk := 0
	retMes := "nofile"
	for {
		mes, ok := client.getNextAnswer(1000)
		if ok {
			//m.pSuccess("ret: %v\n", mes)
			nbOk++
			if mes.Args[0] == "done" && retMes == "nofile" {
				retMes = "done"
			} else if mes.Args[0] != "nofile" {
				retMes = mes.Args[0]
			}
			if nbOk == client.nbNode {
				break
			}
		}
		if time.Now().Sub(t0) > time.Second*3 {
			break
		}
	}
	t1 := time.Now()
	if retMes == "done" {
		m.pSuccess("File %s removed time=%dms\n", fileName, t1.Sub(t0).Nanoseconds()/1000000)
	} else if retMes == "nofile" {
		m.pWarn("File %s not found time=%dms\n", fileName, t1.Sub(t0).Nanoseconds()/1000000)
	} else {
		m.pWarn("remove file %s error: %s\n", fileName, retMes)
	}
	return nil
}
