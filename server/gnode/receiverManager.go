package gnode

import (
	"fmt"
	"io"
	"time"
)

type ReceiverManager struct {
	usage        int
	gnode        *GNode
	buffer       MessageBuffer
	receiverList []*MessageReceiver
	ioChan       chan *AntMes
	nbReceiver   int
	receiver     MessageReceiver
	answerMap    map[string]*AntMes
	getChan      chan string
	functionMap  map[string]*agridInternalFunction
}

type agridInternalFunction struct {
	function    interface{}
	returnError bool
}

func (m *ReceiverManager) loadFunctions() {
	m.functionMap = make(map[string]*agridInternalFunction)
	m.functionMap["storeBlock"] = &agridInternalFunction{function: m.gnode.fileManager.storeBlock, returnError: true}
	m.functionMap["storeBlocAck"] = &agridInternalFunction{function: m.gnode.fileManager.storeBlockAck, returnError: true}
	m.functionMap["getFileBlocks"] = &agridInternalFunction{function: m.gnode.fileManager.sendBlocks, returnError: true}
	m.functionMap["sendBackBlock"] = &agridInternalFunction{function: m.gnode.fileManager.receivedBackBlock, returnError: true}
	m.functionMap["listFiles"] = &agridInternalFunction{function: m.gnode.fileManager.listFiles, returnError: true}
	m.functionMap["listNodeFiles"] = &agridInternalFunction{function: m.gnode.fileManager.listNodeFiles, returnError: true}
	m.functionMap["sendBackListFilesToClient"] = &agridInternalFunction{function: m.gnode.fileManager.sendBackListFilesToClient, returnError: true}
	m.functionMap["removeFiles"] = &agridInternalFunction{function: m.gnode.fileManager.removeFiles, returnError: true}
	m.functionMap["removeNodeFiles"] = &agridInternalFunction{function: m.gnode.fileManager.removeNodeFiles, returnError: true}
	m.functionMap["sendBackRemoveFilesToClient"] = &agridInternalFunction{function: m.gnode.fileManager.sendBackRemoveFilesToClient, returnError: true}
}

func (m *ReceiverManager) start(gnode *GNode, bufferSize int, maxGoRoutine int) {
	m.gnode = gnode
	m.loadFunctions()
	m.nbReceiver = maxGoRoutine
	m.buffer.init(bufferSize)
	m.ioChan = make(chan *AntMes)
	m.getChan = make(chan string)
	m.answerMap = make(map[string]*AntMes)
	m.receiverList = []*MessageReceiver{}
	if maxGoRoutine <= 0 {
		m.receiver.gnode = gnode
		return
	}
	for i := 0; i < maxGoRoutine; i++ {
		routine := &MessageReceiver{
			id:              i,
			gnode:           m.gnode,
			receiverManager: m,
		}
		m.receiverList = append(m.receiverList, routine)
		routine.start()
	}
	go func() {
		for {
			mes, ok := m.buffer.get(true)
			if ok && mes != nil {
				m.ioChan <- mes
			}
		}
	}()

}

func (m *ReceiverManager) waitForAnswer(id string, timeoutSecond int) (*AntMes, error) {
	if mes, ok := m.answerMap[id]; ok {
		return mes, nil
	}
	timer := time.AfterFunc(time.Second*time.Duration(timeoutSecond), func() {
		m.getChan <- "timeout"
	})
	logf.info("Waiting for answer originId=%s\n", id)
	for {
		retId := <-m.getChan
		if retId == "timeout" {
			return nil, fmt.Errorf("Timeout wiating for message answer id=%s", id)
		}
		if mes, ok := m.answerMap[id]; ok {
			logf.info("Found answer originId=%s\n", id)
			timer.Stop()
			return mes, nil
		}
	}
}

func (m *ReceiverManager) receiveMessage(mes *AntMes) bool {
	m.usage++
	logf.debugMes(mes, "receive message: %s\n", mes.toString())
	if m.nbReceiver <= 0 {
		m.receiver.executeMessage(mes)
		return true
	}
	return m.buffer.put(mes)
}

func (m *ReceiverManager) stats() {
	fmt.Printf("Receiver: nb=%d maxbuf=%d\n", m.usage, m.buffer.max)
	execVal := ""
	for _, exec := range m.receiverList {
		execVal = fmt.Sprintf("%s %d", execVal, exec.usage)
	}
	fmt.Printf("Receivers: %s\n", execVal)
}

func (m *ReceiverManager) startClientReader(stream GNodeService_GetClientStreamServer) {
	clientName := fmt.Sprintf("client-%d", len(m.gnode.clientMap)+1)
	m.gnode.clientMap[clientName] = &gnodeClient{
		name:   clientName,
		stream: stream,
	}
	stream.Send(&AntMes{
		Function:   "ClientAck",
		FromClient: clientName,
	})
	logf.info("Client stream open: %s\n", clientName)
	for {
		mes, err := stream.Recv()
		if err == io.EOF {
			logf.error("Client reader %s: EOF\n", clientName)
			delete(m.gnode.clientMap, clientName)
			m.gnode.senderManager.sendMessage(&AntMes{
				Target:   "*",
				Function: "forceGC",
				Args:     []string{"true"},
			})
			forceGC(m.gnode, true)
			return
		}
		if err != nil {
			logf.error("Client reader %s: Failed to receive message: %v\n", clientName, err)
			delete(m.gnode.clientMap, clientName)
			m.gnode.senderManager.sendMessage(&AntMes{
				Target:   "*",
				Function: "forceGC",
				Args:     []string{"true"},
			})
			forceGC(m.gnode, true)
			return
		}
		if mes.Function == "sendBlock" {
			m.gnode.fileManager.receiveBlocFromClient(mes)
		} else {
			mes.Id = m.gnode.getNewId(false)
			mes.Origin = m.gnode.name
			mes.FromClient = clientName
			if mes.Debug {
				logf.debugMes(mes, "-------------------------------------------------------------------------------------------------------------\n")
				logf.debugMes(mes, "Receive mes from client %s : %v\n", clientName, mes)
			}
			m.gnode.idMap.Add(mes.Id)
			if ok := m.gnode.receiverManager.receiveMessage(mes); !ok {
				logf.error("Message put error id=%s\n", mes.Id)
			}
			if !mes.ReturnAnswer {
				stream.Send(&AntMes{Origin: m.gnode.name})
			}
		}
	}
}
