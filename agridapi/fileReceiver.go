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
		if ret == "ok" {
			return nil
		} else if ret != "thread ok" {
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
		currentDuplicate := 1
		nbNodeEnded := 0
		req := &gnode.RetrieveFileRequest{
			Name:      clusterFile,
			ClientId:  client.id,
			NbThread:  int32(nbThread),
			Thread:    int32(thread),
			Duplicate: int32(currentDuplicate),
		}
		if _, err := client.client.RetrieveFile(context.Background(), req); err != nil {
			m.chanReceive <- err.Error()
			return
		}
		blockSize := int64(gnode.GNodeBlockSize * 1024)
		nbBlock := int64(0)
		timer := time.AfterFunc(time.Millisecond*time.Duration(30000), func() {
			currentDuplicate++
			if currentDuplicate > client.nbDuplicate {
				m.chanReceive <- "retrieve file timeout"
			}
			if ret := m.nodeEndedOrTimeout(client, currentDuplicate, req, nbThread, thread, int(nbBlock)); ret {
				return
			}
		})
		nb := 0
		for {
			mes, _ := client.getNextAnswer(0)
			if mes.Eof {
				m.api.info("Received node %s EOF\n", mes.Origin)
				nbNodeEnded++
				if nbNodeEnded >= client.nbNode {
					m.api.info("Received nodes EOF\n")
					currentDuplicate++
					if currentDuplicate > client.nbDuplicate {
						if len(m.orderMap) == 0 {
							m.chanReceive <- "file doesn't exist"
						} else {
							m.chanReceive <- "missing blocks to complete file"
						}
						return
					}
					nbNodeEnded = 0
					if ret := m.nodeEndedOrTimeout(client, currentDuplicate, req, nbThread, thread, int(nbBlock)); ret {
						return
					}
				}
			} else {
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
			}
			timer.Reset(time.Millisecond * 30000)
		}
		timer.Stop()
		m.chanReceive <- "ok"
		return
	}()
}

func (m *fileReceiver) nodeEndedOrTimeout(client *gnodeClient, currentDuplicate int, req *gnode.RetrieveFileRequest, nbThread int, thread int, nbBlock int) bool {
	blockList := ""
	if nbBlock > 0 {
		blockList = m.getThreadBlockList(nbThread, thread, int(nbBlock))
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

func (m *fileReceiver) getThreadBlockList(nbThread int, thread int, nbBlock int) string {
	m.writeLock.Lock()
	list := "#"
	for block := 1; block <= nbBlock; block++ {
		if block%nbThread == thread {
			if _, ok := m.orderMap[block]; !ok {
				list = fmt.Sprintf("%s%d#", list, block)
			}
		}
	}
	m.writeLock.Unlock()
	return list
}
