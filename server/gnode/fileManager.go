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
const GNodeFileOnGoing = ".!og!"
const GNodeBlockSize = 512 * 1024

type FileManager struct {
	gnode          *GNode
	transferMap    secureMap //map[string]*FileTransfer
	transferNumber int
	lockRead       sync.RWMutex
	commitedMap    map[string]time.Time
	nbRead         int
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
	userName      string
	userToken     string
	version       int
}

//-----------------------------------------------------------------------------------------------------------------------------------
// init

func (f *FileManager) init(gnode *GNode) {
	f.gnode = gnode
	f.transferMap.init()
	f.commitedMap = make(map[string]time.Time)
	f.lockRead = sync.RWMutex{}
}

//-----------------------------------------------------------------------------------------------------------------------------------
// storeFile

func (f *FileManager) storeFile(req *StoreFileRequest) (*StoreFileRet, error) {
	if !f.gnode.checkUser(req.UserName, req.UserToken) {
		return nil, fmt.Errorf("Invalid user/token")
	}
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
		userName:      req.UserName,
		userToken:     req.UserToken,
		version:       int(req.Version),
	}
	f.transferMap.set(req.TransferId, transfer)
	logf.info("storeFile received: req=%v\n", req)
	return &StoreFileRet{}, nil
}

func (f *FileManager) receiveBlocFromClient(mes *AntMes) error {
	if !f.transferMap.exists(mes.TransferId) {
		return fmt.Errorf("Received bloc from client on a not started transfert")
	}
	transfer := f.transferMap.get(mes.TransferId).(*FileTransfer)
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
			Version:      int32(transfer.version),
			Duplicate:    int32(nn + 1),
			Args:         transfer.metadata,
			Data:         data,
			UserName:     transfer.userName,
			UserToken:    transfer.userToken,
		}
		//logf.info("Send client tf=%s order %d to %s\n", mes.TransferId, mes.Order, f.gnode.nodeNameList[pos])
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
	//logf.info("StoreBlockAck tf=%s order %d\n", mes.TransferId, mes.Order)
	if !f.transferMap.exists(mes.TransferId) {
		return fmt.Errorf("Received bloc ack on a not started transfert")
	}
	transfer := f.transferMap.get(mes.TransferId).(*FileTransfer)
	logf.info("storeBackAck: version=%d order=%d duplicate=%d (%d/%d)\n", mes.Version, mes.Order, mes.Duplicate, len(transfer.orderMap), transfer.toBeReceived)
	if mes.Duplicate == 1 {
		transfer.lockAck.Lock()
		defer transfer.lockAck.Unlock()
		order := int(mes.Order)
		transfer.orderMap[order] = 1
		if int64(len(transfer.orderMap)) >= transfer.toBeReceived {
			transfer.terminated = true
			f.sendBackStoreMessageToClient(mes, transfer)
			return nil
		}
	} else {
		if mes.Order%100 == 0 {
			t0 := time.Now()
			if t0.Sub(transfer.lastClientMes).Seconds() > 1 {
				transfer.lastClientMes = t0
				f.sendBackStoreMessageToClient(mes, transfer)
			}
		}
	}
	return nil
}

func (f *FileManager) sendBackStoreMessageToClient(mes *AntMes, transfer *FileTransfer) {
	if transfer.terminated {
		logf.info("All store block ack received. File store tf=%s done\n", transfer.id)
		f.gnode.sendBackClient(transfer.clientId, &AntMes{
			Function:   "FileSendAck",
			TransferId: transfer.id,
			Origin:     mes.Origin,
		})
		f.transferMap.del(transfer.id)
	} else {
		f.gnode.sendBackClient(transfer.clientId, &AntMes{
			Function:   "FileSendOngoing",
			TransferId: transfer.id,
			NoBlocking: true,
			Origin:     mes.Origin,
		})
	}
}

func (f *FileManager) storeBlock(mes *AntMes) error {
	//logf.info("StoreBlock tf=%s order %d\n", mes.TransferId, mes.Order)
	err := f.writeBlock(mes)
	if err != nil {
		logf.error("Error writting block order %d: %v\n", mes.Order, err)
	}
	f.gnode.senderManager.sendMessage(&AntMes{
		Function:     "storeBlocAck",
		Target:       mes.Origin,
		TransferId:   mes.TransferId,
		Duplicate:    mes.Duplicate,
		Version:      mes.Version,
		Order:        mes.Order,
		UserName:     mes.UserName,
		TargetedPath: mes.TargetedPath,
	})
	return nil
}

func (f *FileManager) writeBlock(mes *AntMes) error {
	f.storeFileNodeInit(mes)
	dir := f.getGNodeFilePath(mes.UserName, mes.TargetedPath, mes.Version, mes.Duplicate)
	os.MkdirAll(dir, os.ModeDir)
	name := f.getBlockName(mes.Order, mes.NbBlockTotal)
	//logf.info("writeblock: %s / %s\n", dir, name)
	if err := ioutil.WriteFile(path.Join(dir, name), mes.Data, 0666); err != nil {
		return err
	}
	return nil
}

func (f *FileManager) storeFileNodeInit(mes *AntMes) error {
	dir := f.getGNodeFilePath(mes.UserName, mes.TargetedPath, mes.Version, mes.Duplicate)
	if _, err := os.Stat(path.Join(dir, "meta")); err == nil {
		return nil
	}
	os.MkdirAll(dir, os.ModeDir)
	dirParent := path.Dir(dir)
	name := path.Base(mes.TargetedPath)
	if err := ioutil.WriteFile(path.Join(dirParent, fmt.Sprintf("%s%s", name, GNodeFileOnGoing)), []byte("ongoing"), 0666); err != nil {
		return err
	}
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
	return nil
}

func (f *FileManager) getGNodeFilePath(userName string, fileName string, version int32, duplicate int32) string {
	return fmt.Sprintf("%s.%d.%d%s", path.Join(f.gnode.dataPath, userName, fileName), version, duplicate, GNodeFileSuffixe)
}

func (f *FileManager) getBlockName(order int64, total int64) string {
	if order == total {
		return fmt.Sprintf("t.%s.%d", f.formatOrder(order), total)
	} else {
		return fmt.Sprintf("b.%s.%d", f.formatOrder(order), total)
	}
}

func (f *FileManager) formatOrder(order int64) string {
	if order < 10 {
		return fmt.Sprintf("0000%d", order)
	} else if order < 100 {
		return fmt.Sprintf("000%d", order)
	} else if order < 1000 {
		return fmt.Sprintf("00%d", order)
	} else if order < 10000 {
		return fmt.Sprintf("0%d", order)
	}
	return fmt.Sprintf("%d", order)
}

func (f *FileManager) commitFileStorage(mes *AntMes) error {
	logf.info("received commitFileStorage from client, tf=%s\n", mes.TransferId)
	f.commitedMap[mes.TransferId] = time.Now()
	dir := path.Dir(path.Join(config.rootDataPath, mes.UserName, mes.TargetedPath))
	name := path.Base(mes.TargetedPath)
	f.transferMap.del(mes.TransferId)
	os.Remove(path.Join(dir, fmt.Sprintf("%s%s", name, GNodeFileOnGoing)))
	return nil
}

//-----------------------------------------------------------------------------------------------------------------------------------
// removeFile

func (f *FileManager) removeFiles(mes *AntMes) error {
	if !f.gnode.checkUser(mes.UserName, mes.UserToken) {
		return fmt.Errorf("Invalid user/token")
	}
	logf.info("Received removeFiles: %v\n", mes)
	mes.Function = "removeNodeFiles"
	mes.Target = "*"
	f.gnode.receiverManager.receiveMessage(mes)
	f.gnode.senderManager.sendMessage(mes)
	return nil
}

func (f *FileManager) removeNodeFiles(mes *AntMes) error {
	//logf.info("Received removeNodeFiles: %v\n", mes)
	filename := ""
	if len(mes.Args) >= 1 {
		filename = mes.Args[0]
	}
	recursive := false
	if len(mes.Args) >= 2 && mes.Args[1] == "true" {
		recursive = true
	}
	fullName := path.Join(f.gnode.dataPath, mes.UserName, filename)
	//logf.info("fuleName: %s\n", fullName)
	if _, err := os.Stat(fullName); err == nil {
		//logf.info("does exist: %s\n", fullName)
		if !recursive {
			return fmt.Errorf("Trying to remove a directory %s without recusive flag set", filename)
		}
		//logf.info("remove all dir\n")
		if err := os.RemoveAll(fullName); err != nil {
			logf.warn("removeNodeFiles warn: %v\n", err)
			//return err
		}
	} else {
		//logf.info("remove all file\n")
		if err := f.removeAgridFiles(fullName, int(mes.Version)); err != nil {
			logf.warn("removeNodeFiles warn: %v\n", err)
			//return err
		}
	}
	f.gnode.senderManager.sendMessage(&AntMes{
		Target:     mes.Origin,
		Function:   "sendBackRemoveFilesToClient",
		FromClient: mes.FromClient,
		Eof:        true,
		Nodes:      f.gnode.availableNodeList,
	})
	return nil
}

func (f *FileManager) removeAgridFiles(fullname string, version int) error {
	dir := path.Dir(fullname)
	name := path.Base(fullname)
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range fileList {
		if strings.HasPrefix(file.Name(), name) && strings.HasSuffix(file.Name(), GNodeFileSuffixe) {
			//logf.info("remove: %s/%s\n", dir, file.Name())
			ok := false
			if version == 0 {
				ok = true
			} else {
				vers, _ := f.extractDataFromGNodeFileName(file.Name())
				if vers == version {
					ok = true
				}
			}
			if ok {
				if err := os.RemoveAll(path.Join(dir, file.Name())); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (f *FileManager) sendBackRemoveFilesToClient(mes *AntMes) error {
	f.gnode.sendBackClient(mes.FromClient, mes)
	return nil
}

//-----------------------------------------------------------------------------------------------------------------------------------
// getStat

func (f *FileManager) getFileStat(mes *AntMes) error {
	logf.info("Received getNodeFileStat: %v\n", mes)
	found := false
	if mes.Version == 0 {
		v, exist, err := f.getLastVersion(mes.UserName, mes.TargetedPath, mes.Version)
		if err != nil {
			return err
		} else if exist {
			mes.Version = int32(v)
			if v > 0 {
				found = true
			}
		}
	} else {
		if exist, err := f.isVersionExist(mes.UserName, mes.TargetedPath, mes.Version); err != nil {
			return err
		} else if exist {
			found = true
		}
	}
	orderMax := 0
	length := int64(0)
	if mes.Version > 0 {
		if len(mes.Args) == 0 || mes.Args[0] != "versionOnly" {
			fileName := f.getGNodeFilePath(mes.UserName, mes.TargetedPath, mes.Version, mes.Duplicate)
			logf.info("ReadDir fullName: %s\n", fileName)
			fileList, err := ioutil.ReadDir(fileName)
			if err == nil {
				for _, file := range fileList {
					if strings.HasPrefix(file.Name(), "t.") {
						found = true
						order, _, err := f.extractDataFromBlockName(file.Name())
						if err == nil {
							orderMax = order
							length = file.Size()
						}
					}
				}
			}
		}
	}
	ans := f.gnode.createAnswer(mes, true)
	ans.Size = int64(orderMax-1)*int64(GNodeBlockSize) + length
	ans.Version = mes.Version
	ans.Nodes = f.gnode.availableNodeList
	ans.Args = []string{fmt.Sprintf("%t", found)}
	logf.info("stat send answer: %v\n", ans)
	f.gnode.senderManager.sendMessage(ans)
	return nil
}

func (f *FileManager) getLastVersion(userName string, filePath string, version int32) (int, bool, error) {
	fullName := f.getGNodeFilePath(userName, filePath, version, 1)
	dir := path.Dir(fullName)
	filePath = path.Base(filePath)
	//logf.info("getLastVersion dir=%s filename=%s\n", dir, filePath)
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		//logf.error("ReadDir error: %v\n", err)
		return 0, false, nil
	}
	lastVersion := 0
	filePath += "."
	for _, file := range fileList {
		//logf.info("fileRead: %s\n", file.Name())
		if strings.HasPrefix(file.Name(), filePath) {
			//logf.info("found %s\n", file.Name())
			version, err := f.extractDataFromGNodeFileName(file.Name())
			logf.info("version=%d", version)
			if err == nil && lastVersion < version {
				lastVersion = version
			}
		}
	}
	return lastVersion, true, nil
}

func (f *FileManager) isVersionExist(userName string, filePath string, version int32) (bool, error) {
	fullName := f.getGNodeFilePath(userName, filePath, version, 1)
	dir := path.Dir(fullName)
	filePath = path.Base(filePath)
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, nil
	}
	filePath += "."
	for _, file := range fileList {
		if strings.HasPrefix(file.Name(), filePath) {
			vers, err := f.extractDataFromGNodeFileName(file.Name())
			if err == nil && vers == int(version) {
				return true, nil
			}
		}
	}
	return false, nil
}

//gnode file name format is [baseName].[version].[duplicate].[GNodeFilePrefix]
func (f *FileManager) extractDataFromGNodeFileName(name string) (int, error) {
	if name == "meta" {
		return 0, nil
	}
	datas := strings.Split(name, ".")
	if len(datas) < 4 {
		return 0, fmt.Errorf("Invalid gnode file name: %s", name)
	}
	version, err := strconv.Atoi(datas[len(datas)-3])
	if err != nil {
		return 0, fmt.Errorf("Invalid version format in name: %s", name)
	}
	/* no need for duplicate for now
	duplicate, err := strconv.Atoi(datas[len(data)-2])
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid duplicate format in name: %s", name)
	}
	*/
	return version, nil
}

//-----------------------------------------------------------------------------------------------------------------------------------
// listFile

func (f *FileManager) listFiles(mes *AntMes) error {
	//logf.info("Received listFile: %v\n", mes)
	if !f.gnode.checkUser(mes.UserName, mes.UserToken) {
		return fmt.Errorf("Invalid user/token")
	}
	mes.Function = "listNodeFiles"
	mes.Target = "*"
	f.gnode.receiverManager.receiveMessage(mes)
	f.gnode.senderManager.sendMessage(mes)
	return nil
}

func (f *FileManager) listNodeFiles(mes *AntMes) error {
	//logf.info("Received listNodeFile: %v\n", mes)
	folder := mes.UserName
	if len(mes.Args) >= 1 {
		folder = path.Join(mes.UserName, mes.Args[0])
	}
	version := false
	if len(mes.Args) >= 2 {
		if mes.Args[1] == "withVersion" {
			version = true
		}
	}
	//logf.info("Receive listFiles on path %s\n", folder)
	pathname := path.Join(config.rootDataPath, folder)
	args, order := f.listFolder(mes, pathname, 0, []string{}, version)
	f.sendListFilesBack(mes, order+1, args, true)
	return nil
}

func (f *FileManager) listFolder(mes *AntMes, pathname string, order int, args []string, version bool) ([]string, int) {
	files, _ := ioutil.ReadDir(pathname)
	for _, fl := range files {
		name := path.Join(pathname, fl.Name())
		if strings.HasSuffix(name, GNodeFileSuffixe) {
			line := f.getTrueName(name, version)[len(f.gnode.dataPath):]
			args = append(args, line)
			if len(args) >= 3 {
				order++
				f.sendListFilesBack(mes, order, args, false)
				args = []string{}
			}
		} else {
			args, order = f.listFolder(mes, name, order, args, version)
		}
	}
	return args, order
}

func (f *FileManager) sendListFilesBack(mes *AntMes, order int, args []string, eof bool) {
	//logf.info("sendListFilesBack: %v\n", mes)
	ans := &AntMes{
		Function:   "sendBackListFilesToClient",
		Target:     mes.Origin,
		Origin:     f.gnode.name,
		FromClient: mes.FromClient,
		Order:      int64(order),
		Eof:        eof,
		Args:       args,
	}
	if eof {
		ans.Nodes = f.gnode.availableNodeList
	}
	f.gnode.senderManager.sendMessage(ans)
}

func (f *FileManager) getTrueName(name string, version bool) string {
	list := strings.Split(name, ".")
	tname := list[0]
	if len(list) > 4 {
		for _, part := range list[1 : len(list)-3] {
			tname += "." + part
		}
	}
	if version {
		return fmt.Sprintf("%s (v%s)", tname, list[len(list)-3])
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
	if !f.gnode.checkUser(req.UserName, req.UserToken) {
		return nil, fmt.Errorf("Invalid user/token")
	}
	f.transferNumber++
	md5 := md5.New()
	io.WriteString(md5, req.Name)
	transferId := fmt.Sprintf("TF-%x-%d", md5.Sum(nil), time.Now().UnixNano())
	logf.info("retrieveFile received: transferId=%s req=%v\n", transferId, req)
	transfer := &FileTransfer{
		clientId: req.ClientId,
		id:       transferId,
		name:     req.Name,
		version:  int(req.Version),
		orderMap: make(map[int]byte),
	}
	f.transferMap.set(transferId, transfer)
	mes := &AntMes{
		Target:       "*",
		Function:     "getFileBlocks",
		TargetedPath: req.Name,
		TransferId:   transferId,
		NbThread:     req.NbThread,
		Thread:       req.Thread,
		Duplicate:    req.Duplicate,
		UserName:     req.UserName,
		UserToken:    req.UserToken,
		Version:      req.Version,
	}
	mes.Origin = f.gnode.name
	mes.Args = []string{req.BlockList}
	f.gnode.receiverManager.receiveMessage(mes)
	f.gnode.senderManager.sendMessage(mes)
	return &RetrieveFileRet{TransferId: transferId}, nil
}

func (f *FileManager) sendBlocks(mes *AntMes) error {
	//logf.info("sendBlocks tf:%s  cl: %s order:%d\n", mes.TransferId, mes.FromClient, mes.Order)
	fileName := f.getGNodeFilePath(mes.UserName, mes.TargetedPath, mes.Version, mes.Duplicate)
	nbThread := int(mes.NbThread)
	thread := int(mes.Thread)
	blockList := mes.Args[0]
	files, _ := ioutil.ReadDir(fileName)
	for _, fl := range files {
		order, nbBlock, err := f.extractDataFromBlockName(fl.Name())
		if err != nil {
			logf.error("Error extracting data from name: %v\n", err)
		} else {
			if order == 0 || (order%nbThread == thread && (blockList == "" || strings.Index(blockList, fmt.Sprintf("#%d#", order)) >= 0)) {
				name := path.Join(fileName, fl.Name())
				f.lockRead.Lock() //only for multiple local nodes install: TODO: to be removed
				data, err := ioutil.ReadFile(name)
				f.lockRead.Unlock()
				if err != nil {
					logf.error("Error reading file %s\n", name)
				} else {
					mesr := &AntMes{
						Origin:     f.gnode.name,
						Target:     mes.Origin,
						Function:   "sendBackBlock",
						TransferId: mes.TransferId,
						Order:      int64(order),
						NbBlock:    int64(nbBlock),
						Data:       data,
					}
					f.gnode.senderManager.sendMessage(mesr)
				}
			}
		}
	}
	return nil
}

//block name format is b.[order].[TotalBlockNumber]
func (f *FileManager) extractDataFromBlockName(name string) (int, int, error) {
	if name == "meta" {
		return 0, 0, nil
	}
	datas := strings.Split(name, ".")
	if len(datas) != 3 {
		return 0, 0, fmt.Errorf("Invalid block file name: %s", name)
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
	//logf.info("receivedBackBlock tf:%s  cl: %s order:%d\n", mes.TransferId, mes.FromClient, mes.Order)
	if !f.transferMap.exists(mes.TransferId) {
		logf.error("Received back bloc on a not started transfert id=%s", mes.TransferId)
		return fmt.Errorf("Received back bloc on a not started transfert id=%s", mes.TransferId)
	}
	transfer := f.transferMap.get(mes.TransferId).(*FileTransfer)
	transfer.lockAck.Lock()
	defer transfer.lockAck.Unlock()
	_, ok := transfer.orderMap[int(mes.Order)]
	if !ok || mes.Order == 0 {
		transfer.orderMap[int(mes.Order)] = 1
		f.gnode.sendBackClient(transfer.clientId, mes)
	}
	return nil
}

func (f *FileManager) moveRandomBlock() {
	//TODO
}

//----------------------------------------------------------------------------------------------
// direct file save block

func (f *FileManager) fileSaveBlock(mes *AntMes) error {
	if !f.gnode.checkUser(mes.UserName, mes.UserToken) {
		return fmt.Errorf("Invalid user/token")
	}
	//logf.info("Received fileSaveBlock client=%s file=%s order=%d\n", mes.FromClient, mes.TargetedPath, mes.Order)
	pos := int(rand.Int31n(int32(len(f.gnode.nodeNameList))))
	if pos == f.gnode.nodeIndex && len(f.gnode.nodeNameList) > 3 {
		pos = pos + int(rand.Int31n(int32(len(f.gnode.nodeNameList)-1)))
		if pos >= len(f.gnode.nodeNameList) {
			pos = pos - len(f.gnode.nodeNameList)
		}
	}
	for nn := 0; nn < config.nbDuplicate; nn++ {
		block := &AntMes{
			Target:       f.gnode.nodeNameList[pos],
			Function:     "fileNodeSaveBlock",
			Order:        mes.Order,
			TargetedPath: mes.TargetedPath,
			NbBlockTotal: mes.NbBlockTotal,
			Duplicate:    int32(nn + 1),
			FromClient:   mes.FromClient,
			UserName:     mes.UserName,
			UserToken:    mes.UserToken,
			Data:         mes.Data,
		}
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

func (f *FileManager) fileNodeSaveBlock(mes *AntMes) error {
	//logf.info("Received fileNodeSaveBlock file=%s order=%d\n", mes.TargetedPath, mes.Order)
	for duplicate := 1; duplicate <= config.nbDuplicate; duplicate++ {
		fileName := fmt.Sprintf("%s.%d%s", path.Join(f.gnode.dataPath, mes.UserName, mes.TargetedPath), mes.Duplicate, GNodeFileSuffixe)
		fmt.Printf("fileName: %s\n", fileName)
		files, _ := ioutil.ReadDir(fileName)
		for _, fl := range files {
			if fl.Name() != "meta" {
				order, nbBlock, err := f.extractDataFromBlockName(fl.Name())
				fmt.Printf("find file %s: %d-%d\n", fl.Name(), order, nbBlock)
				if err == nil {
					if nbBlock < int(mes.NbBlockTotal) {
						fmt.Printf("rename %s -> %s\n", fl.Name(), f.getBlockName(int64(order), int64(mes.NbBlockTotal)))
						os.Rename(path.Join(fileName, fl.Name()), path.Join(fileName, f.getBlockName(int64(order), int64(mes.NbBlockTotal))))
					}
				}
			}
		}
	}
	if err := f.writeBlock(mes); err != nil {
		return err
	}
	f.gnode.senderManager.sendMessage(&AntMes{
		Target:       mes.Origin,
		Function:     "fileSaveBlockReturnClient",
		FromClient:   mes.FromClient,
		TargetedPath: mes.TargetedPath,
		Order:        mes.Order,
	})
	return nil
}

func (f *FileManager) fileSaveBlockReturnClient(mes *AntMes) error {
	//logf.info("Received fileSaveBlockReturnClient client=%s file=%s order=%d\n", mes.FromClient, mes.TargetedPath, mes.Order)
	f.gnode.sendBackClient(mes.FromClient, mes)
	return nil
}

//----------------------------------------------------------------------------------------------
// direct file save block

func (f *FileManager) fileLoadBlocks(mes *AntMes) error {
	//logf.info("Received fileLoadBlocks: %v\n", mes)
	if len(mes.Args) == 0 {
		return nil
	}
	fileName := fmt.Sprintf("%s.%d%s", path.Join(f.gnode.dataPath, mes.UserName, mes.TargetedPath), mes.Duplicate, GNodeFileSuffixe)
	blockList := mes.Args[0]
	files, _ := ioutil.ReadDir(fileName)
	nbSent := 0
	for _, fl := range files {
		if fl.Name() != "meta" {
			order, _, err := f.extractDataFromBlockName(fl.Name())
			if err == nil {
				if strings.Index(blockList, fmt.Sprintf("#%d#", order)) >= 0 {
					name := path.Join(fileName, fl.Name())
					f.lockRead.Lock() //only for multiple local nodes install: TODO: to be removed
					data, err := ioutil.ReadFile(name)
					f.lockRead.Unlock()
					if err != nil {
						logf.error("Error reading file %s\n", name)
					} else {
						ans := f.gnode.createAnswer(mes, false)
						ans.Order = int64(order)
						ans.Data = data
						f.gnode.senderManager.sendMessage(ans)
						nbSent++
					}
				}
			}
		}
	}
	ans := f.gnode.createAnswer(mes, true)
	ans.Order = 0
	ans.Eof = true
	ans.NbBlock = int64(nbSent)
	//logf.info("send back answer: %v\n", ans)
	f.gnode.senderManager.sendMessage(ans)
	return nil
}
