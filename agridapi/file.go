package agridapi

import (
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
)

type AFile struct {
	api          *AgridAPI
	client       *gnodeClient
	name         string
	key          string
	fileSeek     int64
	nbTotalBlock int
	length       int64
	isCreated    bool
	cipher       *gCipher
	blocks       map[int]*fBlock
	version      int
}

type fBlock struct {
	order int
	min   int
	max   int
	data  []byte
	saved bool
}

func (api *AgridAPI) CreateFile(name string, key string) (*AFile, error) {
	af := AFile{}
	af.init(api, name, key)
	cli, err := api.getClient()
	if err != nil {
		return nil, err
	}
	stat, exist, errv := api.getFileStat(cli, name, 0, true)
	if errv != nil {
		return nil, errv
	}
	if exist {
		af.version = stat.Version
	}
	af.version++
	af.client = cli
	af.isCreated = true
	api.info("Create file %s version=%d\n", af.name, af.version)
	return &af, nil
}

func (api *AgridAPI) OpenFile(name string, version int, key string) (*AFile, error) {
	af := AFile{}
	af.init(api, name, key)
	cli, err := api.getClient()
	if err != nil {
		return nil, err
	}
	if stat, exist, err := api.getFileStat(cli, name, version, false); err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("File %s doesn't exist", name)
	} else {
		af.version = stat.Version
		af.fileSeek = stat.Length
		af.length = stat.Length
		af.nbTotalBlock = int(af.length/int64(gnode.GNodeBlockSize)) + 1
	}
	af.client = cli
	api.info("Open file %s version=%d length=%d\n", af.name, af.version, af.length)
	return &af, nil
}

func (af *AFile) init(api *AgridAPI, name string, key string) {
	af.key = api.formatKey(key)
	af.api = api
	af.name = name
	af.blocks = make(map[int]*fBlock)
}

func (af *AFile) Close() error {
	if err := af.Sync(); err != nil {
		return err
	}
	af.client.close()
	return nil
}

func (af *AFile) WriteString(data string) (int, error) {
	return af.Write([]byte(data))
}

func (af *AFile) Write(data []byte) (int, error) {
	orderMin, min, orderMax, max := af.getBoundaries(len(data))
	af.api.info("write seek=%d lenData=%d orderMin=%d:%d orderMax=%d:%d\n", af.fileSeek, len(data), orderMin, min, orderMax, max)
	if err := af.loadBlocks(orderMin, orderMax); err != nil {
		return 0, err
	}
	if orderMin == orderMax {
		block := af.getBlock(orderMin)
		if block.min > min {
			block.min = min
		}
		if block.max < max {
			block.max = max
		}
		cd := 0
		for c := min; c < max; c++ {
			block.data[c] = data[cd]
			cd++
		}
		block.saved = false
	} else {
		dataIndex := 0
		for b := orderMin; b <= orderMax; b++ {
			block := af.getBlock(b)
			if b == orderMin {
				if block.min > min {
					block.min = min
				}
				block.max = gnode.GNodeBlockSize
				for c := min; c < gnode.GNodeBlockSize; c++ {
					block.data[c] = data[dataIndex]
					dataIndex++
				}
			} else if b == orderMax {
				block.min = 0
				block.max = gnode.GNodeBlockSize
				for c := 0; c < max; c++ {
					block.data[c] = data[dataIndex]
					dataIndex++
				}
			} else {
				block.min = 0
				block.max = gnode.GNodeBlockSize
				for c := 0; c < gnode.GNodeBlockSize; c++ {
					block.data[c] = data[dataIndex]
					dataIndex++
				}
			}
			block.saved = false
		}
	}
	af.fileSeek += int64(len(data))
	if af.nbTotalBlock < orderMax {
		af.nbTotalBlock = orderMax
	}
	if af.length < af.fileSeek {
		af.length = af.fileSeek
	}
	return len(data), nil
}

func (af *AFile) WriteStringAt(data string, at int64) (int, error) {
	return af.WriteAt([]byte(data), at)
}

func (af *AFile) WriteAt(data []byte, at int64) (int, error) {
	if _, err := af.Seek(at, 0); err != nil {
		return 0, err
	}
	return af.Write(data)
}

func (af *AFile) Read(data []byte) (int, error) {
	orderMin, min, orderMax, max := af.getBoundaries(len(data))
	af.api.info("read seek=%d lenData=%d orderMin=%d:%d orderMax=%d:%d\n", af.fileSeek, len(data), orderMin, min, orderMax, max)
	if err := af.loadBlocks(orderMin, orderMax); err != nil {
		return 0, err
	}
	if orderMin == orderMax {
		block := af.getBlock(orderMin)
		if block.min > min {
			block.min = min
		}
		if block.max < max {
			block.max = max
		}
		cd := 0
		for c := min; c < max; c++ {
			data[cd] = block.data[c]
			cd++
		}
	} else {
		dataIndex := 0
		for b := orderMin; b <= orderMax; b++ {
			block := af.getBlock(b)
			if b == orderMin {
				if block.min > min {
					block.min = min
				}
				block.max = gnode.GNodeBlockSize
				for c := min; c < gnode.GNodeBlockSize; c++ {
					data[dataIndex] = block.data[c]
					dataIndex++
				}
			} else if b == orderMax {
				block.min = 0
				block.max = gnode.GNodeBlockSize
				for c := 0; c < max; c++ {
					data[dataIndex] = block.data[c]
					dataIndex++
				}
			} else {
				block.min = 0
				block.max = gnode.GNodeBlockSize
				for c := 0; c < gnode.GNodeBlockSize; c++ {
					data[dataIndex] = block.data[c]
					dataIndex++
				}
			}
		}
	}
	af.fileSeek += int64(len(data))
	return len(data), nil

}

func (af *AFile) ReadAt(data []byte, at int64) (int, error) {
	if _, err := af.Seek(at, 0); err != nil {
		return 0, err
	}
	return af.Read(data)
}

func (af *AFile) ReadString() (string, error) {
	return "", nil
}

func (af *AFile) Seek(offset int64, whence int) (int64, error) {
	if whence == 0 {
		af.fileSeek = offset
		return af.fileSeek, nil
	} else if whence == 1 {
		af.fileSeek += offset
		return af.fileSeek, nil
	}
	af.fileSeek = af.length + offset
	return af.fileSeek, nil
}

func (af *AFile) Sync() error {
	return af.saveBlocks()
}

func (af *AFile) getBoundaries(size int) (int, int, int, int) {
	orderMin := int(af.fileSeek/int64(gnode.GNodeBlockSize)) + 1
	min := int(af.fileSeek % int64(gnode.GNodeBlockSize))
	orderMax := int((af.fileSeek+int64(size))/int64(gnode.GNodeBlockSize)) + 1
	max := int((af.fileSeek + int64(size)) % int64(gnode.GNodeBlockSize))
	return orderMin, min, orderMax, max

}

func (af *AFile) getBlock(order int) *fBlock {
	block, ok := af.blocks[order]
	if !ok {
		block = &fBlock{}
		block.data = make([]byte, gnode.GNodeBlockSize, gnode.GNodeBlockSize)
		block.order = order
		block.min = gnode.GNodeBlockSize
		block.saved = af.isCreated
		af.blocks[order] = block
	}
	return block
}

func (af *AFile) loadBlocks(orderMin int, orderMax int) error {
	list := "#"
	nbBlock := 0
	for c := orderMin; c <= orderMax; c++ {
		block, ok := af.blocks[c]
		if !ok || !block.saved {
			list = fmt.Sprintf("%s%d#", list, c)
			nbBlock++
		}
	}
	if list == "#" {
		return nil
	}
	if af.key != "" {
		af.api.info("Encrypted load\n")
		af.cipher = &gCipher{}
		af.api.info("Key=%s\n", af.key)
		if err := af.cipher.init([]byte(af.key)); err != nil {
			return fmt.Errorf("Cipher init error: %v", err)
		}
	}
	af.api.info("load block order [%d,%d]\n", orderMin, orderMax)
	_, err := af.client.sendMessage(&gnode.AntMes{
		Function:     "fileLoadBlocks",
		TargetedPath: af.name,
		Target:       "*",
		Duplicate:    1,
		Version:      int32(af.version),
		UserName:     af.api.userName,
		UserToken:    af.api.userToken,
		Args:         []string{list},
	}, false)
	if err != nil {
		return nil
	}
	orderMap := make(map[int]byte)
	nodeMap := make(map[string]byte)
	nbReceived := 0
	nbOrigin := 0
	for {
		mes, err := af.client.getNextAnswer(3000)
		if err != nil {
			return err
		}
		af.api.info("received origin=%s order=%d orderMap=%d\n", mes.Origin, mes.Order, len(orderMap))
		if mes.Eof {
			nbOrigin++
			for _, nodeName := range mes.Nodes {
				nodeMap[nodeName] = 1
			}
			nbReceived += int(mes.NbBlock)
			if len(nodeMap) > 0 && nbOrigin == len(nodeMap) {
				if nbReceived == 0 {
					af.api.info("No block found\n")
					return nil
				}
			}
		} else {
			order := int(mes.Order)
			block := af.getBlock(order)
			block.data = make([]byte, gnode.GNodeBlockSize, gnode.GNodeBlockSize)
			if af.cipher != nil {
				dat, err := af.cipher.decrypt(mes.Data)
				if err != nil {
					return err
				}
				mes.Data = dat
			}
			for c := 0; c < len(mes.Data); c++ {
				block.data[c] = mes.Data[c]
			}
			block.saved = true
			block.min = 0
			block.max = len(mes.Data)
			orderMap[order] = 1
			if len(orderMap) == nbBlock {
				break
			}
		}
	}
	return nil
}

func (af *AFile) saveBlocks() error {
	if len(af.blocks) == 0 {
		return nil
	}
	if af.key != "" {
		af.api.info("Encrypted write\n")
		af.cipher = &gCipher{}
		if err := af.cipher.init([]byte(af.key)); err != nil {
			return fmt.Errorf("Cipher init error: %v", err)
		}
	}
	nbSend := 0
	for _, block := range af.blocks {
		if !block.saved {
			nbSend++
			size := block.max - block.min
			if size > 0 {
				data := make([]byte, size, size)
				for c := 0; c < size; c++ {
					data[c] = block.data[block.min+c]
				}
				if af.cipher != nil {
					dat, err := af.cipher.encrypt(data)
					if err != nil {
						return err
					}
					data = dat
				}
				af.api.info("save block order=%d\n", block.order)
				_, err := af.client.sendMessage(&gnode.AntMes{
					Function:     "fileSaveBlock",
					TargetedPath: af.name,
					Target:       "",
					Version:      int32(af.version),
					Order:        int64(block.order),
					NbBlockTotal: int64(af.nbTotalBlock),
					UserName:     af.api.userName,
					UserToken:    af.api.userToken,
					Data:         data,
				}, false)
				if err != nil {
					return err
				}
			}
		}
	}
	if nbSend == 0 {
		return nil
	}
	af.api.info("Waiting cluster ack\n")
	orderMap := make(map[int]byte)
	for {
		mes, err := af.client.getNextAnswer(3000)
		if err != nil {
			return err
		}
		af.api.info("Received ack order=%d\n", mes.Order)
		order := int(mes.Order)
		block := af.getBlock(order)
		block.saved = true
		block.min = 0
		block.max = len(mes.Data)
		orderMap[order] = 1
		if len(orderMap) == nbSend {
			break
		}
	}
	af.api.info("cluster ack received\n")
	af.api.info("send commit\n")
	if _, err := af.client.sendMessage(&gnode.AntMes{
		Function:     "fileNodeSaveCommit",
		TargetedPath: af.name,
		Target:       "*",
		Version:      int32(af.version),
		NbBlockTotal: int64(af.nbTotalBlock),
		UserName:     af.api.userName,
	}, false); err != nil {
		return err
	}
	af.api.info("waiting commit ack\n")
	nodeMap := make(map[string]byte)
	nbOk := 0
	for {
		mes, err := af.client.getNextAnswer(3000)
		if err != nil {
			return err
		}
		if mes.Function == "fileNodeSaveCommitAck" {
			for _, nodeName := range mes.Nodes {
				nodeMap[nodeName] = 1
			}
			nbOk++
			af.api.info("Receive node ack from %s: (%d/%d)\n", mes.Origin, nbOk, len(nodeMap))
			if len(nodeMap) > 0 && nbOk >= len(nodeMap) {
				break
			}
		}
	}
	af.api.info("blocks saved commited\n")
	return nil
}

func (af *AFile) copyBlocksInData(data []byte) {
	min := int(af.fileSeek % int64(gnode.GNodeBlockSize))
	max := int((int64(af.fileSeek) + int64(len(data))) % int64(gnode.GNodeBlockSize))
	if len(af.blocks) == 1 {

		af.copyBlock(data, af.blocks[0], min, max)
	}

}

func (af *AFile) copyBlock(data []byte, block *fBlock, min int, max int) {
	i := 0
	for c := min; c <= max; c++ {
		data[i] = block.data[c]
		i++
	}
}

func (af *AFile) Display() {
	af.api.info("file %s seek: %d nbBlock=%d\n", af.name, af.fileSeek, len(af.blocks))
	list := make([]*fBlock, len(af.blocks), len(af.blocks))
	for i, block := range af.blocks {
		list[i-1] = block
	}
	for i, block := range list {
		af.api.info("block %d order: %d [%d,%d]:%s\n", i, block.order, block.min, block.max, string(block.data))
	}
}
