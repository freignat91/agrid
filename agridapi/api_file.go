package agridapi

import (
	"fmt"
	"sort"
	"time"
)

// FileLs return list of file stored under <folder>
func (api *AgridAPI) FileLs(folder string) ([]string, error) {
	lineList := []string{}
	client, err := api.getClient()
	if err != nil {
		return lineList, err
	}
	defer client.close()
	if _, err := client.createSendMessage("*", false, "listFiles", folder); err != nil {
		return lineList, err
	}
	nbOk := 0
	listMap := make(map[string]byte)
	t0 := time.Now()
	nbWaited := 0
	nbReceived := 0
	for {
		mes, err := client.getNextAnswer(1000)
		if err != nil {
			return lineList, err
		} else {
			nbReceived++
			//api.info("From node %s: order:%d w:%d r:%d\n", mes.Origin, mes.Order, nbWaited, nbReceived)
			for _, line := range mes.Args {
				listMap[line] = 1
			}
			if mes.Eof {
				nbWaited += int(mes.Order)
				nbOk++
				//api.info("Node EOF %s: nbMes:%d  w:%d r:%d  wok=%d rok=%d\n", mes.Origin, mes.Order, nbWaited, nbReceived, client.nbNode, nbOk)
			}
			if nbOk == client.nbNode && nbWaited == nbReceived {
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
func (api *AgridAPI) FileStore(localFile string, clusterPathname string, meta []string, nbThread int, key string) error {
	fileSender := fileSender{}
	fileSender.init(api)
	if err := fileSender.storeFile(localFile, clusterPathname, meta, nbThread, key); err != nil {
		return err
	}
	return nil
}

// FileGet get a file from cluster
func (api *AgridAPI) FileRetrieve(clusterPathname string, localFile string, nbThread int, key string) error {
	fileReceiver := fileReceiver{}
	fileReceiver.init(api)
	if err := fileReceiver.retrieveFile(clusterPathname, localFile, nbThread, key); err != nil {
		return err
	}
	return nil
}

// FileRm remove a file from cluster
func (api *AgridAPI) FileRm(clusterPathname string, recursive bool) error {
	client, err := api.getClient()
	if err != nil {
		return err
	}
	defer client.close()
	if _, err := client.createSendMessage("*", false, "removeFiles", clusterPathname, fmt.Sprintf("%t", recursive)); err != nil {
		return err
	}
	t0 := time.Now()
	nbOk := 0
	for {
		mes, err := client.getNextAnswer(1000)
		if err != nil {
			return err
		} else {
			api.info("Receive answer: %v\n", mes.Origin)
			if len(mes.Args) > 0 {
				return fmt.Errorf("%s", mes.Args[0])
			}
			nbOk++
			if nbOk == client.nbNode {
				break
			}
		}
		if time.Now().Sub(t0) > time.Second*3 {
			return fmt.Errorf("Remove file %s timeout", clusterPathname)
		}
	}
	return nil
}
