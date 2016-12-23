package gnode

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const GNodeFileSuffixe = ".!fg!"
const GNodeBlockSize = 512

type FileManager struct {
	gnode          *GNode
	transferMap    map[string]*FileTransfer
	transferNumber int
}

type FileTransfer struct {
	clientId      string
	id            string
	name          string
	toBeReceived  int64
	nbBlockTotal  int64
	path          string
	orderMap      map[int]byte
	lockAck       sync.RWMutex
	lastClientMes time.Time
	terminated    bool
	blockSize     int64
	metadata      []string
}

func (f *FileManager) init(gnode *GNode) {
	f.gnode = gnode
	f.transferMap = make(map[string]*FileTransfer)
}

//-----------------------------------------------------------------------------------------------------------------------------------
// storeFile

func (f *FileManager) storeFile(req *StoreFileRequest) (*StoreFileRet, error) {
	f.transferNumber++
	transfer := &FileTransfer{
		clientId:      req.ClientId,
		id:            req.TransferId,
		name:          req.Name,
		path:          req.Path,
		blockSize:     req.BlockSize,
		toBeReceived:  req.NbBlock,
		nbBlockTotal:  req.NbBlockTotal,
		orderMap:      make(map[int]byte),
		lockAck:       sync.RWMutex{},
		lastClientMes: time.Now(),
		metadata:      req.Metadata,
	}
	f.transferMap[req.TransferId] = transfer
	logf.info("sendFile received: req=%+v\n", req)
	return &StoreFileRet{}, nil
}

func (f *FileManager) receiveBlocFromClient(mes *AntMes) error {
	transfer, ok := f.transferMap[mes.TransferId]
	if !ok {
		return fmt.Errorf("Received bloc from client on a not started transfert")
	}
	pos := int(rand.Int31n(int32(len(f.gnode.nodeNameList))))
	if pos == f.gnode.nodeIndex && len(f.gnode.nodeNameList) > 3 {
		pos = pos + int(rand.Int31n(int32(len(f.gnode.nodeNameList)-1)))
		if pos >= len(f.gnode.nodeNameList) {
			pos = pos - len(f.gnode.nodeNameList)
		}
	}
	data := mes.Data
	for nn := 0; nn < config.nbDuplicate; nn++ {
		block := &AntMes{
			Target:       f.gnode.nodeNameList[pos],
			Function:     "storeBlock",
			TransferId:   transfer.id,
			Order:        mes.Order,
			NbBlock:      transfer.toBeReceived,
			NbBlockTotal: transfer.nbBlockTotal,
			Size:         mes.Size,
			TargetedPath: transfer.path,
			Duplicate:    int32(nn + 1),
			Args:         transfer.metadata,
			Data:         data,
		}
		//logf.info("Send client order %d to %s\n", mes.Order, f.gnode.nodeNameList[pos])
		f.gnode.senderManager.sendMessage(block)
		pos++
		if pos >= len(f.gnode.nodeNameList) {
			pos = 0
		}
		if pos == f.gnode.nodeIndex && len(f.gnode.nodeNameList) > 3 {
			pos++
			if pos >= len(f.gnode.nodeNameList) {
				pos = 0
			}
		}
	}
	return nil
}

func (f *FileManager) storeBlockAck(mes *AntMes) error {
	transfer, ok := f.transferMap[mes.TransferId]
	if !ok {
		return fmt.Errorf("Received bloc ack on a not started transfert")
	}
	transfer.lockAck.Lock()
	defer transfer.lockAck.Unlock()
	order := int(mes.Order)
	transfer.orderMap[order] = byte(1)
	if int64(len(transfer.orderMap)) >= transfer.toBeReceived {
		transfer.terminated = true
		f.sendBackStoreMessageToClient(transfer)
		return nil
	} else {
		if mes.Order%100 == 0 {
			t0 := time.Now()
			if t0.Sub(transfer.lastClientMes).Seconds() > 1 {
				transfer.lastClientMes = t0
				f.sendBackStoreMessageToClient(transfer)
			}
		}
	}
	return nil
}

func (f *FileManager) sendBackStoreMessageToClient(transfer *FileTransfer) {
	if transfer.terminated {
		logf.info("All store block ack received. File store done\n")
		delete(f.transferMap, transfer.id)
		f.gnode.sendBackClient(transfer.clientId, &AntMes{
			Function: "FileSendAck",
		})
	} else {
		f.gnode.sendBackClient(transfer.clientId, &AntMes{
			Function:   "FileSendOngoing",
			NoBlocking: true,
		})
	}
}

func (f *FileManager) storeBlock(mes *AntMes) error {
	//logf.info("bloc data stored order=%d\n", mes.Order)
	err := f.writeBlock(mes)
	if err != nil {
		logf.error("Error writting block order %d: %v\n", mes.Order, err)
	}
	f.gnode.senderManager.sendMessage(&AntMes{
		Function:   "storeBlocAck",
		Target:     mes.Origin,
		TransferId: mes.TransferId,
		Order:      mes.Order,
	})
	return nil
}

func (f *FileManager) writeBlock(mes *AntMes) error {
	dir := fmt.Sprintf("%s/%s.%d%s", f.gnode.dataPath, mes.TargetedPath, mes.Duplicate, GNodeFileSuffixe)
	os.MkdirAll(dir, os.ModeDir)
	name := fmt.Sprintf("b.%d.%d", mes.Order, mes.NbBlockTotal)
	//logf.info("writeblock: %s / %s\n", dir, name)
	file, err := os.Create(path.Join(dir, name))
	if err != nil {
		return err
	}
	defer file.Close()
	_, errw := file.Write(mes.Data)
	if errw != nil {
		return errw
	}
	//return file.Sync()
	if len(mes.Args) > 0 {
		if _, err := os.Stat(path.Join(dir, "meta")); os.IsNotExist(err) {
			filed, err := os.Create(path.Join(dir, "meta"))
			if err != nil {
				return err
			}
			defer filed.Close()
			for _, line := range mes.Args {
				_, err := filed.WriteString(line + "\n")
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//-----------------------------------------------------------------------------------------------------------------------------------
// removetFile

func (f *FileManager) removeFile(fileName string, recursive bool) string {
	fullName := path.Join(f.gnode.dataPath, fileName)
	dir := path.Dir(fullName)
	files, _ := ioutil.ReadDir(dir)
	logf.debug("removeFile searching in dir %s\n", dir)
	for _, fl := range files {
		name := path.Join(dir, fl.Name())
		logf.debug("removeFile found file: %s\n", name)
		tname := f.getTrueName(name)
		if tname == fullName {
			if !recursive && !strings.HasSuffix(name, GNodeFileSuffixe) {
				ret := fmt.Sprintf("Trying to remove a directory %s without recusive flag set", name)
				logf.warn("%s\n", ret)
				return ret
			}
			if err := os.RemoveAll(name); err != nil {
				ret := fmt.Sprintf("Error removing file %s: %v", name, err)
				logf.error("%s\n", ret)
				return ret
			}
			logf.info("file %s removed\n", name)
			return "done"
		}
	}
	return "nofile"
}

//-----------------------------------------------------------------------------------------------------------------------------------
// listFile

func (f *FileManager) listFiles(mes *AntMes) error {
	//logf.info("Received listFile: %v\n", mes)
	mes.Function = "listNodeFiles"
	f.gnode.receiverManager.receiveMessage(mes)
	f.gnode.senderManager.sendMessage(mes)
	return nil
}

func (f *FileManager) listNodeFiles(mes *AntMes) error {
	//logf.info("Received listNodeFile: %v\n", mes)
	folder := "/"
	if len(mes.Args) >= 1 {
		folder = mes.Args[0]
	}
	logf.info("Receive listFiles on path %s\n", folder)
	pathname := path.Join(config.rootDataPath, folder)
	args := f.listFolder(mes, pathname, []string{})
	f.sendListFilesBack(mes, args, true)
	return nil
}

func (f *FileManager) listFolder(mes *AntMes, pathname string, args []string) []string {
	files, _ := ioutil.ReadDir(pathname)
	for _, fl := range files {
		name := path.Join(pathname, fl.Name())
		if strings.HasSuffix(name, GNodeFileSuffixe) {
			line := f.getTrueName(name)[len(f.gnode.dataPath):]
			args = append(args, line)
			if len(args) >= 100 {
				f.sendListFilesBack(mes, args, false)
				args = []string{}
			}
		} else {
			args = f.listFolder(mes, name, args)
		}
	}
	return args
}

func (f *FileManager) sendListFilesBack(mes *AntMes, args []string, eof bool) {
	//logf.info("sendListFilesBack: %v\n", mes)
	f.gnode.senderManager.sendMessage(&AntMes{
		Function:   "sendBackListFilesToClient",
		Target:     mes.Origin,
		Origin:     f.gnode.name,
		FromClient: mes.FromClient,
		Eof:        eof,
		Args:       args,
	})
}

func (f *FileManager) getTrueName(name string) string {
	list := strings.Split(name, ".")
	tname := list[0]
	if len(list) > 3 {
		for _, part := range list[1 : len(list)-2] {
			tname += "." + part
		}
	}
	return tname
}

func (f *FileManager) sendBackListFilesToClient(mes *AntMes) error {
	//logf.info("sendBackListFilesToClient: %v\n", mes)
	f.gnode.sendBackClient(mes.FromClient, mes)
	return nil
}

//-----------------------------------------------------------------------------------------------------------------------------------
// RetrieveFile

func (f *FileManager) retrieveFile(req *RetrieveFileRequest) (*RetrieveFileRet, error) {
	f.transferNumber++
	md5 := md5.New()
	io.WriteString(md5, req.Name)
	transferId := fmt.Sprintf("TF-%x-%d", md5.Sum(nil), time.Now().UnixNano())
	logf.info("Received getFile: transferId=%s req=%v\n", transferId, req)
	transfer := &FileTransfer{
		clientId:      req.ClientId,
		id:            transferId,
		name:          req.Name,
		lastClientMes: time.Now(),
		orderMap:      make(map[int]byte),
	}
	f.transferMap[transferId] = transfer
	mes := &AntMes{
		Target:       "*",
		Function:     "getFileBlocks",
		TargetedPath: req.Name,
		TransferId:   transferId,
		NbThread:     req.NbThread,
		Thread:       req.Thread,
		Duplicate:    req.Duplicate,
	}
	mes.Origin = f.gnode.name
	mes.Args = []string{req.BlockList}
	f.gnode.receiverManager.receiveMessage(mes)
	f.gnode.senderManager.sendMessage(mes)
	return &RetrieveFileRet{TransferId: transferId}, nil
}

func (f *FileManager) sendBlocks(mes *AntMes) error {
	//logf.info("sendBlock: %s\n", mes.toString())
	fileName := fmt.Sprintf("%s/%s.%d%s", f.gnode.dataPath, mes.TargetedPath, mes.Duplicate, GNodeFileSuffixe)
	nbThread := int(mes.NbThread)
	thread := int(mes.Thread)
	blockList := mes.Args[0]
	files, _ := ioutil.ReadDir(fileName)
	for _, fl := range files {
		if fl.Name() != "meta" {
			order, nbBlock, err := f.extractDataFromName(fl.Name())
			if err != nil {
				logf.error("Error extracting data from name: %v\n", err)
			} else {
				if order%nbThread == thread && (blockList == "" || strings.Index(blockList, fmt.Sprintf("#%d#", order)) >= 0) {
					name := path.Join(fileName, fl.Name())
					rfile, err := os.Open(name)
					if err != nil {
						logf.error("Error opening file %s\n", name)
					} else {
						defer rfile.Close()
						mesr := &AntMes{
							Origin:     f.gnode.name,
							Target:     mes.Origin,
							Function:   "sendBackBlock",
							TransferId: mes.TransferId,
							Order:      int64(order),
							NbBlock:    int64(nbBlock),
						}
						mesr.Data = make([]byte, GNodeBlockSize*1024, GNodeBlockSize*1024)
						nn, err := rfile.Read(mesr.Data)
						if err != nil {
							logf.error("Error reading file %s\n", name)
						}
						mesr.Data = mesr.Data[:nn]
						//logf.info("sendBlock order=%d\n", order)
						f.gnode.senderManager.sendMessage(mesr)
					}
				}
			}
		}
	}
	mesr := &AntMes{
		Origin:     f.gnode.name,
		Target:     mes.Origin,
		Function:   "sendBackBlock",
		TransferId: mes.TransferId,
		Eof:        true,
	}
	f.gnode.senderManager.sendMessage(mesr)
	return nil
}

func (f *FileManager) extractDataFromName(name string) (int, int, error) {
	datas := strings.Split(name, ".")
	if len(datas) != 3 {
		return 0, 0, fmt.Errorf("Invalid block file name: %s", name)
	}
	if datas[0] != "b" {
		return 0, 0, fmt.Errorf("Invalid block file name header: %s", name)
	}
	order, err := strconv.Atoi(datas[1])
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid order format in name: %s", name)
	}
	nbBlock, err := strconv.Atoi(datas[2])
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid nbBlock format in name: %s", name)
	}
	return order, nbBlock, nil
}

func (f *FileManager) receivedBackBlock(mes *AntMes) error {
	transfer, ok := f.transferMap[mes.TransferId]
	if !ok {
		return fmt.Errorf("Received back bloc on a not started transfert id=%s", mes.TransferId)
	}
	transfer.lockAck.Lock()
	defer transfer.lockAck.Unlock()
	if _, ok := transfer.orderMap[int(mes.Order)]; !ok {
		if !mes.Eof {
			transfer.orderMap[int(mes.Order)] = 1
		}
		f.gnode.sendBackClient(transfer.clientId, mes)
	}
	return nil
}

func (f *FileManager) moveRandomBlock() {
	//TODO
}
