package agridapi

import (
	"fmt"
	"sort"
	"time"
)

type AFileStat struct {
	Name    string
	User    string
	Length  int64
	Version int
}

// FileLs return list of file stored under <folder>
func (api *AgridAPI) FileLs(folder string, version bool) ([]string, error) {
	lineList := []string{}
	client, err := api.getClient()
	if err != nil {
		return lineList, err
	}
	defer client.close()
	sversion := ""
	if version {
		sversion = "withVersion"
	}
	if _, err := client.createSendMessage("*", false, "listFiles", folder, sversion); err != nil {
		return lineList, err
	}
	nodeMap := make(map[string]byte)
	listMap := make(map[string]byte)
	t0 := time.Now()
	nbWaited := 0
	nbReceived := 0
	nbOk := 0
	for {
		mes, err := client.getNextAnswer(1000)
		if err != nil {
			return lineList, err
		} else {
			//api.info("Received mes with nodes: %v (%d/%d)\n", mes.Nodes, nbOk, len(nodeMap))
			nbReceived++
			//api.info("From node %s: order:%d w:%d r:%d\n", mes.Origin, mes.Order, nbWaited, nbReceived)
			for _, line := range mes.Args {
				listMap[line] = 1
			}
			if mes.Eof {
				nbWaited += int(mes.Order)
				for _, nodeName := range mes.Nodes {
					nodeMap[nodeName] = 1
				}
				nbOk++
				//api.info("Node EOF %s: nbMes:%d  w:%d r:%d  wok=%d rok=%d\n", mes.Origin, mes.Order, nbWaited, nbReceived, client.nbNode, nbOk)
			}
			if len(nodeMap) > 0 && nbOk >= len(nodeMap) && nbWaited == nbReceived {
				break
			}
		}
		if time.Now().Sub(t0).Seconds() > 3 {
			break
		}
	}
	for key, _ := range listMap {
		if key != "" {
			filename := key[len(api.userName)+1:]
			lineList = append(lineList, filename)
		}
	}
	sort.Strings(lineList)
	return lineList, nil
}

// FileStore store a file in cluster
func (api *AgridAPI) FileStore(localFile string, clusterPathname string, meta []string, nbThread int, key string) (int, error) {
	fileSender := fileSender{}
	fileSender.init(api)
	version, err := fileSender.storeFile(localFile, clusterPathname, meta, nbThread, key)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// FileGet get a file from cluster, return metadata and error
func (api *AgridAPI) FileRetrieve(clusterPathname string, localFile string, version int, nbThread int, key string) (map[string]string, int, error) {
	fileReceiver := fileReceiver{}
	fileReceiver.init(api)
	meta, version, err := fileReceiver.retrieveFile(clusterPathname, localFile, version, nbThread, key)
	if err != nil {
		return nil, 0, err
	}
	return meta, version, nil
}

// FileRm remove a file from cluster
func (api *AgridAPI) FileRm(clusterPathname string, version int, recursive bool) error {
	client, err := api.getClient()
	if err != nil {
		return err
	}
	defer client.close()
	mes := client.createMessage("*", false, "removeFiles", clusterPathname, fmt.Sprintf("%t", recursive))
	mes.Version = int32(version)
	if _, err := client.sendMessage(mes, false); err != nil {
		return err
	}
	nodeMap := make(map[string]byte)
	nbOk := 0
	for {
		mes, err := client.getNextAnswer(1000)
		if err != nil {
			return err
		}
		api.info("Receive answer: %v\n", mes.Origin)
		if len(mes.Args) > 0 {
			return fmt.Errorf("%s", mes.Args[0])
		}
		for _, nodeName := range mes.Nodes {
			nodeMap[nodeName] = 1
		}
		nbOk++
		if len(nodeMap) > 0 && nbOk >= len(nodeMap) {
			break
		}
	}
	return nil
}

// FileStat return stat of file name
func (api *AgridAPI) FileStat(name string, version int) (*AFileStat, bool, error) {
	client, err := api.getClient()
	if err != nil {
		return nil, false, err
	}
	defer client.close()
	return api.getFileStat(client, name, version, false)
}

func (api *AgridAPI) getFileStat(client *gnodeClient, name string, version int, versionOnly bool) (*AFileStat, bool, error) {
	api.info("getFileStat for %s\n", name)
	mes := client.createMessage("*", false, "getFileStat", name)
	mes.TargetedPath = name
	mes.Duplicate = 1
	mes.Version = int32(version) //if 0, get last one
	if versionOnly {
		mes.Args = []string{"versionOnly"}
	}
	if _, err := client.sendMessage(mes, false); err != nil {
		return nil, false, err
	}
	nbOk := 0
	nodeMap := make(map[string]byte)
	nbNotFound := 0
	stat := AFileStat{Name: name, User: api.userName}
	for {
		mes, err := client.getNextAnswer(1000)
		if err != nil {
			return nil, false, err
		}
		stat.Version = int(mes.Version)
		if mes.Size > stat.Length {
			stat.Length = mes.Size
		}
		if mes.Args[0] == "false" {
			nbNotFound++
		}
		for _, nodeName := range mes.Nodes {
			nodeMap[nodeName] = 1
		}
		nbOk++
		api.info("Receive answer: %v (%d/%d) found=%s nodes=%v\n", mes.Origin, nbOk, len(nodeMap), mes.Args[0], mes.Nodes)
		if len(nodeMap) > 0 && nbOk >= len(nodeMap) {
			break
		}
	}
	if nbNotFound >= len(nodeMap) {
		return nil, false, nil
	}
	api.info("stat=%+v\n", stat)
	return &stat, true, nil
}
