package agridapi

import (
	"crypto/md5"
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
	"golang.org/x/net/context"
	"io"
	"os"
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

func (m *fileSender) storeFile(fileName string, target string, meta []string, nbThread int, key string) error {
	key = m.api.formatKey(key)
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
	md5 := md5.New()
	io.WriteString(md5, fileName)
	tId := fmt.Sprintf("TF-%x-%d", md5.Sum(nil), time.Now().UnixNano())
	blockSize := int64(gnode.GNodeBlockSize) * 1024
	totalBlock := length / blockSize
	if length%blockSize > 0 {
		totalBlock++
	}
	m.api.info("store nbThread=%d nbBlock=%d\n", nbThread, totalBlock)
	if nbThread > int(totalBlock) {
		nbThread = int(totalBlock)
		m.api.info("nbThread ajusted to: %d\n", nbThread)
	}
	m.initClients(nbThread)
	defer m.close()
	transferIds := []string{}

	if key != "" {
		m.api.info("Encrypted transfer\n")
		m.cipher = &gCipher{}
		if err := m.cipher.init([]byte(key)); err != nil {
			return fmt.Errorf("Cipher init error: %v", err)
		}
	}

	for i, client := range m.clients {
		nbBlock := totalBlock / int64(nbThread)
		if totalBlock%int64(nbThread) >= int64(i+1) {
			nbBlock++
		}
		transferId := fmt.Sprintf("%s-%d", tId, i)
		transferIds = append(transferIds, transferId)
		m.api.info("client %d tf=%s nbBlock=%d\n", i, transferId, nbBlock)
		_, err := client.client.StoreFile(context.Background(), &gnode.StoreFileRequest{
			Name:         fileName,
			Path:         target,
			NbBlockTotal: totalBlock,
			NbBlock:      nbBlock,
			ClientId:     client.id,
			TransferId:   transferId,
			Metadata:     meta,
			BlockSize:    int64(blockSize * 1024),
			Key:          key,
			UserName:     m.api.userName,
			UserToken:    m.api.userToken,
		})
		if err != nil {
			return err
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
				return fmt.Errorf("file %s storage timeout", fileName)
			}
			if mes.Function == "FileSendAck" {
				nbOk++
			} else {
				m.api.info("File store ongoing\n")
			}
		}
		if nbOk >= nbThread {
			break
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
