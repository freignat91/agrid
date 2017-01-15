package agridapi

import (
	"crypto/md5"
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
	"golang.org/x/net/context"
	"io"
	"os"
	"strconv"
	"time"
)

type fileSender struct {
	api           *AgridAPI
	nbFile        int
	clients       []*gnodeClient
	currentClient int
	cipher        *gCipher
}

func (m *fileSender) init(api *AgridAPI) {
	m.api = api
	m.currentClient = 0
}

func (m *fileSender) storeFile(fileName string, target string, meta []string, nbThread int, key string) (int, error) {
	key = m.api.formatKey(key)
	f, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	st, errs := f.Stat()
	if errs != nil {
		return 0, errs
	}
	length := st.Size()
	md5 := md5.New()
	io.WriteString(md5, target)
	tId := fmt.Sprintf("TF-%x-%d", md5.Sum(nil), time.Now().Unix())
	blockSize := int64(gnode.GNodeBlockSize)
	totalBlock := length / blockSize
	if length%blockSize > 0 {
		totalBlock++
	}
	m.api.info("store nbThread=%d nbBlock=%d\n", nbThread, totalBlock)
	if nbThread > int(totalBlock) {
		nbThread = int(totalBlock)
		m.api.info("nbThread ajusted to: %d\n", nbThread)
	}
	if err := m.initClients(nbThread); err != nil {
		return 0, err
	}
	defer m.close()
	transferIds := []string{}

	if key != "" {
		m.api.info("Encrypted transfer\n")
		m.cipher = &gCipher{}
		if err := m.cipher.init([]byte(key)); err != nil {
			return 0, fmt.Errorf("Cipher init error: %v", err)
		}
	}
	stat, exist, errv := m.api.getFileStat(m.clients[0], target, 0, true)
	if errv != nil {
		return 0, fmt.Errorf("getFileStat: %v", errv)
	}
	version := 0
	if exist {
		version = stat.Version + 1
		m.api.info("Found file %s version %d, store version %d\n", target, version-1, version)
	} else {
		m.api.info("No file %s version found, store version 1\n", target)
		version = 1
	}
	for i, client := range m.clients {
		nbBlock := totalBlock / int64(nbThread)
		if totalBlock%int64(nbThread) >= int64(i+1) {
			nbBlock++
		}
		transferId := fmt.Sprintf("%s-%d", tId, i)
		transferIds = append(transferIds, transferId)
		m.api.info("client %d tf=%s nbBlock=%d\n", i, transferId, nbBlock)
		req := &gnode.StoreFileRequest{
			Name:         fileName,
			Path:         target,
			NbBlockTotal: totalBlock,
			NbBlock:      nbBlock,
			ClientId:     client.id,
			TransferId:   transferId,
			Metadata:     meta,
			BlockSize:    int64(blockSize),
			Key:          key,
			Version:      int32(version),
			UserName:     m.api.userName,
			UserToken:    m.api.userToken,
		}
		_, err := client.client.StoreFile(context.Background(), req)
		if err != nil {
			return 0, err
		}
	}
	m.api.info("Bloc size: %d\n", blockSize)
	block := &gnode.AntMes{
		Target:   "",
		Function: "sendBlock",
		Data:     make([]byte, blockSize),
		Size:     int64(blockSize),
		Order:    0,
	}
	m.currentClient = -1
	for {

		block.Data = block.Data[:cap(block.Data)]
		n, err := f.Read(block.Data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		block.Data = block.Data[:n]
		if m.cipher != nil {
			dat, err := m.cipher.encrypt(block.Data)
			if err != nil {
				return 0, err
			}
			block.Data = dat
		}
		block.Size = int64(n)
		block.Order++
		client := m.getNextClient()
		block.TransferId = transferIds[m.currentClient]
		m.api.info("Send block order %d/%d\n", block.Order, totalBlock)
		client.sendMessage(block, false)
	}
	okMap := make(map[string]byte)
	m.api.info("Waiting for cluster ack\n")
	for {
		for _, client := range m.clients {
			mes, err := client.getNextAnswer(120000)
			if err != nil {
				return 0, err
			}
			//m.api.info(" waiting cluster ack received: %v\n", mes)
			if mes.Function == "blockAsking" {
				m.api.info("ReSent %d blocks\n", len(mes.Args))
				m.sendBlocks(f, block, client, transferIds[m.currentClient], mes.Args)
			} else if mes.Function == "FileStoreAck" {
				okMap[mes.TransferId] = 1
			}
		}
		if len(okMap) >= nbThread {
			break
		}
	}
	m.api.info("Cluster ack received\n")
	for i := 0; i < nbThread; i++ {
		client := m.getNextClient()
		client.sendMessage(&gnode.AntMes{
			Function:     "commitFileStorage",
			Target:       "*",
			UserName:     m.api.userName,
			TargetedPath: target,
			Version:      int32(version),
			TransferId:   transferIds[m.currentClient],
			FromClient:   client.id,
		}, false)
	}
	m.api.info("Waiting for cluster commit ack\n")
	okMap = make(map[string]byte)
	for {
		for _, client := range m.clients {
			mes, err := client.getNextAnswer(120000)
			if err != nil {
				return 0, err
			}
			//m.api.info(" waiting commit ack received: %v\n", mes)
			if mes.Function == "FileCommitAck" {
				okMap[mes.TransferId] = 1
			}
		}
		if len(okMap) >= nbThread {
			break
		}
	}
	m.api.info("Storage commited\n")
	return version, nil
}

func (m *fileSender) sendBlocks(file *os.File, block *gnode.AntMes, client *gnodeClient, transferId string, list []string) error {
	for _, sorder := range list {
		order, err := strconv.ParseInt(sorder, 10, 64)
		if err == nil {
			block.Data = block.Data[:cap(block.Data)]
			file.Seek((order-1)*int64(cap(block.Data)), 0)
			n, err := file.Read(block.Data)
			if err != nil {
				return err
			}
			block.Data = block.Data[:n]
			block.Order = order
			block.Eof = true
			if m.cipher != nil {
				dat, err := m.cipher.encrypt(block.Data)
				if err == nil {
					block.Data = dat
				} else {
					block.Data = nil
				}
			}
			if block.Data != nil {
				block.TransferId = transferId
				//m.api.info("send block order:%d\n", order)
				client.sendMessage(block, false)
			}
		}
	}
	return nil
}

func (m *fileSender) close() {
	for _, client := range m.clients {
		client.close()
	}
}

func (m *fileSender) initClients(nb int) error {
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

func (m *fileSender) getNextClient() *gnodeClient {
	m.currentClient++
	if m.currentClient >= len(m.clients) {
		m.currentClient = 0
	}
	return m.clients[m.currentClient]
}
