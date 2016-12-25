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
	if _, err := client.createSendMessage("*", false, "listFiles", folder); err != nil {
		return lineList, err
	}
	nbOk := 0
	listMap := make(map[string]byte)
	t0 := time.Now()
	for {
		mes, ok := client.getNextAnswer(1000)
		if ok {
			if mes.ErrorMes != "" {
				return lineList, fmt.Errorf("%s", mes.ErrorMes)
			}
			for _, line := range mes.Args {
				listMap[line] = 1
			}
			if mes.Eof {
				api.info("Node EOF %s: %v\n", mes.Origin, mes)
				nbOk++
				if nbOk == client.nbNode {
					break
				}
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
	if _, err := client.createSendMessage("*", false, "removeFiles", clusterPathname, fmt.Sprintf("%t", recursive)); err != nil {
		return err
	}
	t0 := time.Now()
	nbOk := 0
	for {
		mes, ok := client.getNextAnswer(1000)
		if ok {
			if mes.ErrorMes != "" {
				return fmt.Errorf("%s", mes.ErrorMes)
			}
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
