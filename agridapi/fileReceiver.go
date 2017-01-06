package agridapi

import (
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
	"golang.org/x/net/context"
	"os"
	"strings"
	"sync"
	"time"
)

type fileReceiver struct {
	api         *AgridAPI
	nbFile      int
	cipher      *gCipher
	chanReceive chan string
	writeLock   sync.RWMutex
	clients     []*gnodeClient
	meta        map[string]string
}

type receiveSession struct {
	client           *gnodeClient
	orderMap         map[int]byte
	currentDuplicate int
	req              *gnode.RetrieveFileRequest
	nbThread         int
	thread           int
	nbBlock          int
	lastBlockList    string
	blocListTry      int
}

func (m *fileReceiver) init(api *AgridAPI) {
	m.api = api
	m.chanReceive = make(chan string)
	m.writeLock = sync.RWMutex{}
}

func (m *fileReceiver) retrieveFile(clusterFile string, localFile string, version int, nbThread int, key string) (map[string]string, int, error) {
	key = m.api.formatKey(key)
	m.meta = make(map[string]string)
	file, err := os.Create(localFile)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()
	if key != "" {
		m.api.info("Encrypted transfer\n")
		m.cipher = &gCipher{}
		if err := m.cipher.init([]byte(key)); err != nil {
			return nil, 0, fmt.Errorf("Cipher init error: %v", err)
		}
	}
	if err := m.initClients(nbThread); err != nil {
		return nil, 0, err
	}
	defer m.close()
	stat, exist, errv := m.api.getFileStat(m.clients[0], clusterFile, version, true)
	if errv != nil {
		return nil, 0, errv
	}
	if !exist {
		if version != 0 {
			return nil, 0, fmt.Errorf("File %s version %d doesn't exist", clusterFile, version)
		} else {
			return nil, 0, fmt.Errorf("File %s doesn't exist", clusterFile)
		}
	}
	for thread := 0; thread < nbThread; thread++ {
		m.retrieveFileThread(clusterFile, thread, nbThread, stat.Version, key, file)
	}
	nbThreadEnded := 0
	for {
		ret := <-m.chanReceive
		if ret == "ok" {
			nbThreadEnded++
			m.api.info("NbThread completed: %d\n", nbThreadEnded)
			if nbThreadEnded == nbThread {
				return m.meta, stat.Version, nil
			}
		} else {
			return nil, 0, fmt.Errorf("%s", ret)
		}
	}
}

func (m *fileReceiver) retrieveFileThread(clusterFile string, thread int, nbThread int, version int, key string, file *os.File) {
	go func() {
		client := m.clients[thread]
		req := &gnode.RetrieveFileRequest{
			Name:      clusterFile,
			ClientId:  client.id,
			NbThread:  int32(nbThread),
			Thread:    int32(thread),
			Duplicate: 1,
			Version:   int32(version),
			UserName:  m.api.userName,
			UserToken: m.api.userToken,
		}
		m.api.info("req: %v\n", req)
		session := &receiveSession{
			client:           client,
			orderMap:         make(map[int]byte),
			currentDuplicate: 1,
			req:              req,
			nbThread:         nbThread,
			thread:           thread,
		}
		if _, err := client.client.RetrieveFile(context.Background(), req); err != nil {
			m.chanReceive <- err.Error()
			return
		}
		blockSize := int64(gnode.GNodeBlockSize)
		totalBlock := int64(0)
		timer := time.AfterFunc(time.Second*time.Duration(120), func() {
			m.chanReceive <- "retrieve file timeout"
		})
		nb := 0
		metaDataReceived := false
		timeoutDelay := 1000
		for {
			mes, err := client.getNextAnswer(timeoutDelay)
			if err != nil {
				if err.Error() == "Error: timeout" {
					if ended := m.receivedTimeout(session); ended {
						break
					}
					timeoutDelay = 10000
				} else {
					m.chanReceive <- err.Error()
					return
				}
			} else {
				timeoutDelay = 1000
				session.blocListTry = 0
				if session.nbBlock == 0 && mes.Order > 0 {
					totalBlock = mes.NbBlock
					session.nbBlock = int(totalBlock / int64(nbThread))
					if thread != 0 && totalBlock%int64(nbThread) >= int64(thread) {
						session.nbBlock++
					}
					m.api.info("Thread %d nbBlock to receive: %d/%d", thread, session.nbBlock, totalBlock)
					if nbThread >= session.nbBlock && session.nbBlock > 0 {
						if thread == 0 || thread > session.nbBlock {
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
				if mes.Order == 0 {
					if !metaDataReceived {
						metaDataReceived = true
						if err := m.loadMetadata(mes); err != nil {
							m.api.warn("Warning: error loading metadata: %v\n", err)
						} else {
							m.api.warn("Thread %d received metadata\n", thread)
						}
					}
					if session.nbBlock > 0 && len(session.orderMap) == session.nbBlock {
						break
					}
				} else {
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
					session.orderMap[int(mes.Order)] = 1
					if len(session.orderMap) == session.nbBlock && metaDataReceived {
						m.api.info("Thread %d received last mes %d/%d (%d)\n", thread, len(session.orderMap), session.nbBlock, mes.Order)
						break
					}
					m.api.info("Thread %d received mes %d/%d (%d)\n", thread, len(session.orderMap), session.nbBlock, mes.Order)
				}
				timer.Reset(time.Second * 120)
			}
		}
		timer.Stop()
		m.api.info("Thread %d completed\n", thread)
		m.chanReceive <- "ok"
		return
	}()
}

func (m *fileReceiver) receivedTimeout(session *receiveSession) bool {
	blockList, nn := m.getThreadBlockList(session.orderMap, session.nbThread, session.thread, session.nbBlock)
	if blockList == "#" {
		m.api.info("Thread %d completed\n", session.thread)
		m.chanReceive <- "thread ok"
		return true
	}
	if blockList == session.lastBlockList {
		session.blocListTry++
		if session.blocListTry > session.client.nbDuplicate {
			m.chanReceive <- fmt.Sprintf("still missing %d blocks on all duplicate", nn)
			return true
		}
		session.currentDuplicate++
		if session.currentDuplicate > session.client.nbDuplicate {
			session.currentDuplicate = 1
		}
	} else {
		session.blocListTry = 1
	}
	session.lastBlockList = blockList
	m.api.info("Thread %d recalls %d blocks duplicate=%d\n", session.thread, nn, session.currentDuplicate)
	//m.api.info("Thread %d recalls blocks duplicate=%d: %s\n", session.thread, session.currentDuplicate, blockList)
	session.req.Duplicate = int32(session.currentDuplicate)
	session.req.BlockList = blockList
	if _, err := session.client.client.RetrieveFile(context.Background(), session.req); err != nil {
		m.chanReceive <- err.Error()
		return true
	}
	return false
}

func (m *fileReceiver) getThreadBlockList(orderMap map[int]byte, nbThread int, thread int, totalBlock int) (string, int) {
	list := "#"
	size := 0
	for block := 1; block <= totalBlock; block++ {
		if block%nbThread == thread {
			if _, ok := orderMap[block]; !ok {
				size++
				if size >= 100 {
					break
				}
				list = fmt.Sprintf("%s%d#", list, block)
			}
		}
	}
	return list, size
}

func (m *fileReceiver) loadMetadata(mes *gnode.AntMes) error {
	metaList := strings.Split(string(mes.Data), "\n")
	for _, line := range metaList {
		if line != "" {
			meta := strings.Split(line, "=")
			if len(meta) != 2 {
				return fmt.Errorf("Metadata bad format: %s", line)
			}
			m.meta[meta[0]] = meta[1]
		}
	}
	return nil
}

func (m *fileReceiver) close() {
	for _, client := range m.clients {
		client.close()
	}
}

func (m *fileReceiver) initClients(nb int) error {
	m.clients = make([]*gnodeClient, nb, nb)
	for i, _ := range m.clients {
		if cli, err := m.api.getClient(); err != nil {
			return err
		} else {
			m.clients[i] = cli
		}
	}
	return nil
}
