package agridapi

import (
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
)

type AFile struct {
	api      *AgridAPI
	client   *gnodeClient
	name     string
	key      string
	fileSeek int64
	blocks   map[int]*fBlock
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
	af.client = cli
	return &af, nil
}

func (af *AFile) init(api *AgridAPI, name string, key string) {
	af.api = api
	af.name = name
	af.key = key
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
	orderMin := int(af.fileSeek / gnode.GNodeBlockSize)
	min := int(af.fileSeek % gnode.GNodeBlockSize)
	orderMax := int((af.fileSeek + int64(len(data))) / gnode.GNodeBlockSize)
	max := int((af.fileSeek + int64(len(data))) % gnode.GNodeBlockSize)
	fmt.Printf("write orderMin=%d:%d orderMax=%d:%d\n", orderMin, min, orderMax, max)
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
		af.blocks[orderMin] = block
	} else {
		seek := 0
		for b := orderMin; b <= orderMax; b++ {
			block := af.getBlock(b)
			if b == orderMin {
				if block.min > min {
					block.min = min
				}
				cd := 0
				for c := min; c < gnode.GNodeBlockSize; c++ {
					block.data[c] = data[cd]
					cd++
				}
			} else if b == orderMax {
				if block.min < max {
					block.max = max
				}
				cd := seek
				for c := 0; c < max; c++ {
					block.data[c] = data[cd]
					cd++
				}
			} else {
				block.min = 0
				block.max = gnode.GNodeBlockSize - 1
				cd := seek
				for c := 0; c < gnode.GNodeBlockSize; c++ {
					block.data[c] = data[cd]
					cd++
				}
			}
			af.blocks[b] = block
			seek += gnode.GNodeBlockSize
		}
	}
	af.fileSeek += int64(len(data))
	return len(data), nil
}

func (af *AFile) WriteStringAt(data string, at int64) (int, error) {
	return af.WriteAt([]byte(data), at)
}

func (af *AFile) WriteAt(data []byte, at int64) (int, error) {
	if _, err := af.Seek(at); err != nil {
		return 0, err
	}
	return af.Write(data)
}

func (af *AFile) Seek(offset int64) (int64, error) {
	if err := af.Sync(); err != nil {
		return 0, err
	}
	af.fileSeek = offset
	return 0, nil
}

func (af *AFile) Sync() error {
	return af.saveBlocks()
}

func (af *AFile) getBlock(order int) *fBlock {
	block, ok := af.blocks[order]
	if !ok {
		block = &fBlock{}
		block.data = make([]byte, gnode.GNodeBlockSize, gnode.GNodeBlockSize)
		block.order = order
		block.min = gnode.GNodeBlockSize
	}
	return block
}

func (af *AFile) saveBlocks() error {
	if len(af.blocks) == 0 {
		return nil
	}
	nbSend := 0
	for _, block := range af.blocks {
		if !block.saved {
			nbSend++
			af.api.info("save block order=%d\n", block.order)
			_, err := af.client.sendMessage(&gnode.AntMes{
				Function:     "fileSaveBlock",
				TargetedPath: af.name,
				Target:       "",
				Order:        int64(block.order),
				Data:         block.data,
			}, false)
			if err != nil {
				return err
			}
		}
	}
	if nbSend == 0 {
		return nil
	}
	orderMap := make(map[int]byte)
	for {
		mes, err := af.client.getNextAnswer(3000)
		if err != nil {
			return err
		}
		order := int(mes.Order)
		orderMap[order] = 1
		block, _ := af.blocks[order]
		block.saved = true
		if len(orderMap) == nbSend {
			break
		}
	}
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
	fmt.Printf("file %s seek: %d nbBlock=%d\n", af.name, af.fileSeek, len(af.blocks))
	for i, block := range af.blocks {
		fmt.Printf("block %d order: %d [%d,%d] saved:%t\n", i, block.order, block.min, block.max, block.saved)
	}
}
