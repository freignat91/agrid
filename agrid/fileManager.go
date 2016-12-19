package main

import (
	"crypto/md5"
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
	"golang.org/x/net/context"
	"io"
	"os"
	"time"
)

type fileManager struct {
	clientManager *ClientManager
	nbFile        int
	clients       []*gnodeClient
	currentClient int
	cipher        *gCipher
}

func (m *fileManager) init(clientManager *ClientManager) {
	m.clientManager = clientManager
	m.currentClient = 0
}

func (m *fileManager) send(fileName string, target string, meta []string, bSize int64, nbThread int, key string) error {
	key = m.clientManager.formatKey(key)
	t0 := time.Now()
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	st, errs := f.Stat()
	if errs != nil {
		return errs
	}
	length := st.Size()
	m.initClients(nbThread)
	defer m.close()
	md5 := md5.New()
	io.WriteString(md5, fileName)
	tId := fmt.Sprintf("TF-%x-%d", md5.Sum(nil), time.Now().UnixNano())
	blockSize := bSize * 1024
	totalBlock := length / blockSize
	if length%blockSize > 0 {
		totalBlock++
	}
	transferIds := []string{}

	if key != "" {
		m.clientManager.pInfo("Encrypted transfer\n")
		m.cipher = &gCipher{}
		if err := m.cipher.init([]byte(key)); err != nil {
			m.clientManager.pError("Cipher init error: %v\n", err)
			m.cipher = nil
		}
	}

	for i, client := range m.clients {
		nbBlock := totalBlock / int64(nbThread)
		if totalBlock%int64(nbThread) >= int64(i+1) {
			nbBlock++
		}
		transferId := fmt.Sprintf("%s-%d", tId, i)
		transferIds = append(transferIds, transferId)
		m.clientManager.pRegular("client %d tf=%s nbBlock=%d\n", i, transferId, nbBlock)
		_, err := client.client.SendFile(context.Background(), &gnode.SendFileRequest{
			Name:       fileName,
			Path:       target,
			NbBlock:    nbBlock,
			ClientId:   client.id,
			TransferId: transferId,
			Metadata:   meta,
			BlockSize:  int64(blockSize * 1024),
			Key:        key,
		})
		if err != nil {
			return err
		}
	}
	m.clientManager.pInfo("Bloc size: %d\n", blockSize)
	block := &gnode.AntMes{
		Target:   "",
		Function: "sendBlock",
		Data:     make([]byte, blockSize),
		Size:     int64(blockSize),
		Order:    0,
	}
	m.clientManager.pSuccess("Block size: %d\n", blockSize)
	m.currentClient = -1
	for {

		block.Data = block.Data[:cap(block.Data)]
		n, err := f.Read(block.Data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		block.Data = block.Data[:n]
		if m.cipher != nil {
			dat, err := m.cipher.encrypt(block.Data)
			if err != nil {
				return err
			}
			block.Data = dat
		}
		block.Size = int64(n)
		block.Order++
		client := m.getNextClient()
		block.TransferId = transferIds[m.currentClient]
		client.sendMessage(block, false)
	}
	nbOk := 0
	for {
		for _, client := range m.clients {
			mes, ok := client.getNextAnswer(30000)
			if !ok {
				m.clientManager.Fatal("file %s storage timeout\n", fileName)
			}
			if mes.Function == "FileSendAck" {
				nbOk++
			} else {
				m.clientManager.pRegular("File store ongoing\n")
			}
		}
		if nbOk >= nbThread {
			break
		}
	}
	m.clientManager.pSuccess("file %s stored as %s (%dms)\n", fileName, target, time.Now().Sub(t0).Nanoseconds()/1000000)
	return nil
}

func (m *fileManager) close() {
	for _, client := range m.clients {
		client.close()
	}
}

func (m *fileManager) initClients(nb int) error {
	m.clients = make([]*gnodeClient, nb, nb)
	for i, _ := range m.clients {
		if cli, err := m.clientManager.getClient(); err != nil {
			return err
		} else {
			m.clients[i] = cli
		}
	}
	return nil
}

func (m *fileManager) getNextClient() *gnodeClient {
	m.currentClient++
	if m.currentClient >= len(m.clients) {
		m.currentClient = 0
	}
	return m.clients[m.currentClient]
}

func (m *fileManager) get(clusterFile string, localFile string, key string) error {
	client, err := m.clientManager.getClient()
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
	return m.receiveFile(client, localFile, key)
}

func (m *fileManager) receiveFile(client *gnodeClient, localFile string, key string) error {
	key = m.clientManager.formatKey(key)
	file, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer file.Close()
	orderMap := make(map[int]byte)
	blockSize := int64(gnode.GNodeBlockSize * 1024)
	nbBlock := int64(0)
	timer := time.AfterFunc(time.Millisecond*time.Duration(30000), func() {
		m.clientManager.pError("get file timeout\n")
		os.Exit(1)
	})
	if key != "" {
		m.clientManager.pInfo("Encrypted transfer\n")
		m.cipher = &gCipher{}
		if err := m.cipher.init([]byte(key)); err != nil {
			m.clientManager.pError("Cipher init error: %v\n", err)
			m.cipher = nil
		}
	}

	for {
		mes, _ := client.getNextAnswer(0)
		m.clientManager.pInfo("received mes %d/%d (%d)\n", mes.Order, mes.NbBlock, len(orderMap))
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
	return nil
}
