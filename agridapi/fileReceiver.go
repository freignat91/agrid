package agridapi

import (
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
	"golang.org/x/net/context"
	"os"
	"sync"
	"time"
)

type fileReceiver struct {
	api         *AgridAPI
	nbFile      int
	cipher      *gCipher
	chanReceive chan string
	writeLock   sync.RWMutex
	orderMap    map[int]byte
}

func (m *fileReceiver) init(api *AgridAPI) {
	m.api = api
	m.chanReceive = make(chan string)
	m.writeLock = sync.RWMutex{}
}

func (m *fileReceiver) retrieveFile(clusterFile string, localFile string, nbThread int, key string) error {
	key = m.api.formatKey(key)
	file, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer file.Close()
	if key != "" {
		m.api.info("Encrypted transfer\n")
		m.cipher = &gCipher{}
		if err := m.cipher.init([]byte(key)); err != nil {
			return fmt.Errorf("Cipher init error: %v", err)
		}
	}
	m.orderMap = make(map[int]byte)
	for thread := 0; thread < nbThread; thread++ {
		m.retrieveFileThread(clusterFile, thread, nbThread, key, file)
	}
	for {
		ret := <-m.chanReceive
		if ret != "ok" {
			return fmt.Errorf("%s", ret)
		}
		return nil
	}
}

func (m *fileReceiver) retrieveFileThread(clusterFile string, thread int, nbThread int, key string, file *os.File) {
	go func() {
		client, err := m.api.getClient()
		if err != nil {
			m.chanReceive <- err.Error()
			return
		}
		_, errg := client.client.RetrieveFile(context.Background(), &gnode.RetrieveFileRequest{
			Name:     clusterFile,
			ClientId: client.id,
			NbThread: int32(nbThread),
			Thread:   int32(thread),
		})
		if errg != nil {
			m.chanReceive <- errg.Error()
			return
		}
		blockSize := int64(gnode.GNodeBlockSize * 1024)
		nbBlock := int64(0)
		timer := time.AfterFunc(time.Millisecond*time.Duration(130000), func() {
			m.chanReceive <- "retrieve file timeout"
		})
		nb := 0
		for {
			mes, _ := client.getNextAnswer(0)
			if nbBlock == 0 {
				nbBlock = mes.NbBlock
			}
			if m.cipher != nil {
				dat, err := m.cipher.decrypt(mes.Data)
				if err != nil {
					m.chanReceive <- err.Error()
					return
				}
				mes.Data = dat
			}
			m.writeLock.Lock()
			if _, err := file.Seek((mes.Order-1)*blockSize, 0); err != nil {
				m.chanReceive <- err.Error()
				return
			}
			if _, err := file.Write(mes.Data); err != nil {
				m.chanReceive <- err.Error()
				return
			}
			nb++
			if nb%100 == 0 {
				file.Sync()
			}
			m.orderMap[int(mes.Order)] = 1
			if len(m.orderMap) == int(nbBlock) {
				m.api.info("Thread %d received last mes %d/%d (%d)\n", thread, mes.Order, mes.NbBlock, len(m.orderMap))
				break
			}
			m.writeLock.Unlock()
			m.api.info("Thread %d received mes %d/%d (%d)\n", thread, mes.Order, mes.NbBlock, len(m.orderMap))
			timer.Reset(time.Millisecond * 130000)
		}
		timer.Stop()
		m.chanReceive <- "ok"
		return
	}()
}
