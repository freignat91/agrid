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
	for thread := 0; thread < nbThread; thread++ {
		m.retrieveFileThread(clusterFile, thread, nbThread, key, file)
	}
	nbThreadEnded := 0
	for {
		ret := <-m.chanReceive
		if ret == "ok" {
			nbThreadEnded++
			m.api.info("NbThread completed: %d\n", nbThreadEnded)
			if nbThreadEnded == nbThread {
				return nil
			}
		} else {
			return fmt.Errorf("%s", ret)
		}
	}
}

func (m *fileReceiver) retrieveFileThread(clusterFile string, thread int, nbThread int, key string, file *os.File) {
	go func() {
		client, err := m.api.getClient()
		if err != nil {
			m.chanReceive <- err.Error()
			return
		}
		defer client.close()
		currentDuplicate := 1
		req := &gnode.RetrieveFileRequest{
			Name:      clusterFile,
			ClientId:  client.id,
			NbThread:  int32(nbThread),
			Thread:    int32(thread),
			Duplicate: int32(currentDuplicate),
			UserName:  m.api.userName,
			UserToken: m.api.userToken,
		}
		if _, err := client.client.RetrieveFile(context.Background(), req); err != nil {
			m.chanReceive <- err.Error()
			return
		}
		blockSize := int64(gnode.GNodeBlockSize * 1024)
		nbBlock := int64(0)
		totalBlock := int64(0)
		orderThreadMap := make(map[int]byte)
		timer := time.AfterFunc(time.Millisecond*time.Duration(30000), func() {
			for {
				currentDuplicate++
				if currentDuplicate > client.nbDuplicate {
					m.chanReceive <- "retrieve file timeout"
					return
				}
				if ret := m.nodeEndedOrTimeout(orderThreadMap, client, currentDuplicate, req, nbThread, thread, int(totalBlock)); ret {
					return
				}
				time.Sleep(10 * time.Second)
			}
		})
		nb := 0
		for {
			mes, _ := client.getNextAnswer(0)
			if nbBlock == 0 {
				totalBlock = mes.NbBlock
				nbBlock = totalBlock / int64(nbThread)
				if thread != 0 && totalBlock%int64(nbThread) >= int64(thread) {
					nbBlock++
				}
				m.api.info("Thread %d nbBlock to receive: %d/%d", thread, nbBlock, totalBlock)
				if nbThread >= int(nbBlock) {
					if (thread == 0 && nbBlock != 1) || thread > int(nbBlock) {
						m.api.info("Thread %d not useful concidering the number of blocks\n", thread)
						break
					}
				}
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
				m.writeLock.Unlock()
				m.chanReceive <- err.Error()
				return
			}
			if _, err := file.Write(mes.Data); err != nil {
				m.writeLock.Unlock()
				m.chanReceive <- err.Error()
				return
			}
			m.writeLock.Unlock()
			nb++
			if nb%100 == 0 {
				file.Sync()
			}
			orderThreadMap[int(mes.Order)] = 1
			if len(orderThreadMap) == int(nbBlock) {
				m.api.info("Thread %d received last mes %d/%d (%d)\n", thread, len(orderThreadMap), nbBlock, mes.Order)
				break
			}
			m.api.info("Thread %d received mes %d/%d (%d)\n", thread, len(orderThreadMap), nbBlock, mes.Order)
			timer.Reset(time.Millisecond * 20000)
		}
		timer.Stop()
		m.api.info("Thread %d completed\n", thread)
		m.chanReceive <- "ok"
		return
	}()
}

func (m *fileReceiver) nodeEndedOrTimeout(orderMap map[int]byte, client *gnodeClient, currentDuplicate int, req *gnode.RetrieveFileRequest, nbThread int, thread int, nbBlock int) bool {
	blockList := ""
	if nbBlock > 0 {
		blockList = m.getThreadBlockList(orderMap, nbThread, thread, int(nbBlock))
		if blockList == "#" {
			m.api.info("Thread %d completed\n", thread)
			m.chanReceive <- "thread ok"
			return true
		}
	}
	m.api.info("Thread %d recalls blocks duplicate=%d: %s", thread, currentDuplicate, blockList)
	req.Duplicate = int32(currentDuplicate)
	req.BlockList = blockList
	if _, err := client.client.RetrieveFile(context.Background(), req); err != nil {
		m.chanReceive <- err.Error()
		return true
	}
	return false
}

func (m *fileReceiver) getThreadBlockList(orderMap map[int]byte, nbThread int, thread int, totalBlock int) string {
	list := "#"
	for block := 1; block <= totalBlock; block++ {
		if block%nbThread == thread {
			if _, ok := orderMap[block]; !ok {
				list = fmt.Sprintf("%s%d#", list, block)
			}
		}
	}
	return list
}
