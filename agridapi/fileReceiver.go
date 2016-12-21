package agridapi

import (
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
	"golang.org/x/net/context"
	"os"
	"time"
)

type fileReceiver struct {
	api         *AgridAPI
	nbFile      int
	cipher      *gCipher
	chanReceive chan string
}

func (m *fileReceiver) init(api *AgridAPI) {
	m.api = api
	m.chanReceive = make(chan string)
}

func (m *fileReceiver) get(clusterFile string, localFile string, key string) error {
	client, err := m.api.getClient()
	if err != nil {
		return err
	}
	_, errg := client.client.GetFile(context.Background(), &gnode.GetFileRequest{
		Name:     clusterFile,
		ClientId: client.id,
	})
	if errg != nil {
		return err
	}
	go func() {
		m.receiveFile(client, localFile, key)
	}()
	ret := <-m.chanReceive
	if ret != "ok" {
		return fmt.Errorf("%s", ret)
	}
	return nil
}

func (m *fileReceiver) receiveFile(client *gnodeClient, localFile string, key string) error {
	key = m.api.formatKey(key)
	file, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer file.Close()
	orderMap := make(map[int]byte)
	blockSize := int64(gnode.GNodeBlockSize * 1024)
	nbBlock := int64(0)
	timer := time.AfterFunc(time.Millisecond*time.Duration(30000), func() {
		m.chanReceive <- "get file timeout"
	})
	if key != "" {
		m.api.info("Encrypted transfer\n")
		m.cipher = &gCipher{}
		if err := m.cipher.init([]byte(key)); err != nil {
			return fmt.Errorf("Cipher init error: %v", err)
		}
	}

	for {
		mes, _ := client.getNextAnswer(0)
		m.api.info("received mes %d/%d (%d)\n", mes.Order, mes.NbBlock, len(orderMap))
		if nbBlock == 0 {
			nbBlock = mes.NbBlock
			//fmt.Printf("nbBlock set to %d\n", mes.NbBlock)
		}
		if m.cipher != nil {
			dat, err := m.cipher.decrypt(mes.Data)
			if err != nil {
				return err
			}
			mes.Data = dat
		}
		if _, err := file.Seek((mes.Order-1)*blockSize, 0); err != nil {
			return err
		}
		if _, err := file.Write(mes.Data); err != nil {
			return err
		}
		timer.Reset(time.Millisecond * 30000)
		orderMap[int(mes.Order)] = 1
		if len(orderMap) == int(nbBlock) {
			break
		}
	}
	timer.Stop()
	m.chanReceive <- "ok"
	return nil
}
