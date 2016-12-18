package main

import (
	"github.com/spf13/cobra"
	"sort"
	"strings"
	"time"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var FileLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list file",
	Long:  `list file`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.fileList(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileLsCmd)
}

func (m *ClientManager) fileList(cmd *cobra.Command, args []string) error {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	client, err := m.getClient()
	if err != nil {
		return err
	}
	if _, err := client.createSendMessage("*", false, "listFile", path); err != nil {
		return err
	}
	nbOk := 0
	listMap := make(map[string]string)
	t0 := time.Now()
	for {
		mes, ok := client.getNextAnswer(1000)
		if ok { //&& mes.Function == "fileListReturn" {
			nbOk++
			//m.pSuccess("nb=%d nbOk=%d mes %v\n", nbOk, client.nbNode, mes)
			for _, line := range strings.Split(mes.Args[0], "#") {
				listMap[line] = ""
			}
			if nbOk == client.nbNode {
				break
			}
		}
		if time.Now().Sub(t0).Seconds() > 3 {
			break
		}
	}
	m.pSuccess("nbNode: %d client=%d\n", nbOk, client.nbNode)
	lineList := []string{}
	for key, _ := range listMap {
		if key != "" {
			lineList = append(lineList, key)
		}
	}
	sort.Strings(lineList)
	for _, line := range lineList {
		m.pSuccess("%s\n", line)
	}
	return nil
}
