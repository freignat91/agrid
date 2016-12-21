package agridapi

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// FileLs return list of file stored under <folder>
func (api *AgridAPI) FileLs(folder string) ([]string, error) {
	lineList := []string{}
	client, err := api.getClient()
	if err != nil {
		return lineList, err
	}
	if _, err := client.createSendMessage("*", false, "listFile", folder); err != nil {
		return lineList, err
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
	for key, _ := range listMap {
		if key != "" {
			lineList = append(lineList, key)
		}
	}
	sort.Strings(lineList)
	return lineList, nil
}

// FileStore store a file in cluster
func (api *AgridAPI) FileStore(localFile string, clusterPathname string, meta *[]string, nbThread int, key string) error {
	fileSender := fileSender{}
	fileSender.init(api)
	if err := fileSender.send(localFile, clusterPathname, meta, nbThread, key); err != nil {
		return err
	}
	return nil
}

// FileGet get a file from cluster
func (api *AgridAPI) FileGet(clusterPathname string, localFile string, key string) error {
	fileReceiver := fileReceiver{}
	fileReceiver.init(api)
	if err := fileReceiver.get(clusterPathname, localFile, key); err != nil {
		return err
	}
	return nil
}

// FileRm remove a file from cluster, returned bool = true if file removed, false if file doesn't exist
func (api *AgridAPI) FileRm(clusterPathname string, recursive bool) (error, bool) {
	client, err := api.getClient()
	if err != nil {
		return err, false
	}
	if _, err := client.createSendMessage("*", false, "removeFile", clusterPathname, fmt.Sprintf("%t", recursive)); err != nil {
		return err, false
	}
	t0 := time.Now()
	nbOk := 0
	retMes := "nofile"
	for {
		mes, ok := client.getNextAnswer(1000)
		if ok {
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
	if retMes == "done" {
		return nil, true
	} else if retMes == "nofile" {
		return nil, false
	}
	return fmt.Errorf("remove file %s error: %s\n", clusterPathname, retMes), false
}
