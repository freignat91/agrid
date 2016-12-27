package gnode

import (
	"fmt"
	"io"
	"sync"
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
	lockClient   sync.RWMutex
	functionMap  map[string]interface{}
}

func (m *ReceiverManager) loadFunctions() {
	m.functionMap = make(map[string]interface{})
	//file functions
	m.functionMap["storeBlock"] = m.gnode.fileManager.storeBlock
	m.functionMap["storeBlocAck"] = m.gnode.fileManager.storeBlockAck
	m.functionMap["getFileBlocks"] = m.gnode.fileManager.sendBlocks
	m.functionMap["sendBackBlock"] = m.gnode.fileManager.receivedBackBlock
	m.functionMap["storeClientAck"] = m.gnode.fileManager.storeClientAck
	m.functionMap["listFiles"] = m.gnode.fileManager.listFiles
	m.functionMap["listNodeFiles"] = m.gnode.fileManager.listNodeFiles
	m.functionMap["sendBackListFilesToClient"] = m.gnode.fileManager.sendBackListFilesToClient
	m.functionMap["removeFiles"] = m.gnode.fileManager.removeFiles
	m.functionMap["removeNodeFiles"] = m.gnode.fileManager.removeNodeFiles
	m.functionMap["sendBackRemoveFilesToClient"] = m.gnode.fileManager.sendBackRemoveFilesToClient
	//node Functions
	m.functionMap["ping"] = m.gnode.nodeFunctions.ping
	m.functionMap["pingFromTo"] = m.gnode.nodeFunctions.pingFromTo
	m.functionMap["setLogLevel"] = m.gnode.nodeFunctions.setLogLevel
	m.functionMap["killNode"] = m.gnode.nodeFunctions.killNode
	m.functionMap["updateGrid"] = m.gnode.nodeFunctions.updateGrid
	m.functionMap["writeStatsInLog"] = m.gnode.nodeFunctions.writeStatsInLog
	m.functionMap["clear"] = m.gnode.nodeFunctions.clear
	m.functionMap["forceGC"] = m.gnode.nodeFunctions.forceGCMes
	m.functionMap["getConnections"] = m.gnode.nodeFunctions.getConnections
	m.functionMap["createUser"] = m.gnode.nodeFunctions.createUser
	m.functionMap["createNodeUser"] = m.gnode.nodeFunctions.createNodeUser
	m.functionMap["removeUser"] = m.gnode.nodeFunctions.removeUser
	m.functionMap["removeNodeUser"] = m.gnode.nodeFunctions.removeNodeUser
}

func (m *ReceiverManager) start(gnode *GNode, bufferSize int, maxGoRoutine int) {
	m.gnode = gnode
	m.loadFunctions()
	m.lockClient = sync.RWMutex{}
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
			//logf.info("Receive message ok=%t %v\n", ok, mes.toString())
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

func (m *ReceiverManager) receiveMessage(mes *AntMes) {
	refused := false
	for {
		m.usage++
		logf.debugMes(mes, "recceive message: %s\n", mes.toString())
		if m.nbReceiver <= 0 {
			m.receiver.executeMessage(mes)
			return
		}
		if m.buffer.put(mes) {
			if refused {
				logf.warn("Received message: message re-accepted: %v\n", mes.Id)
			}
			refused = false
			return
		}
		logf.warn("Received message: buffer full, message temporary refused: %v\n", mes.toString())
		refused = true
		time.Sleep(1 * time.Second)
	}
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
	m.lockClient.Lock()
	clientName := fmt.Sprintf("client-%d-%d", time.Now().UnixNano(), m.gnode.clientMap.len()+1)
	m.gnode.clientMap.set(clientName, &gnodeClient{
		name:   clientName,
		stream: stream,
	})
	stream.Send(&AntMes{
		Function:   "ClientAck",
		FromClient: clientName,
	})
	logf.info("Client stream open: %s\n", clientName)
	m.lockClient.Unlock() //unlock far to be sure to have several nano
	for {
		mes, err := stream.Recv()
		if err == io.EOF {
			logf.error("Client reader %s: EOF\n", clientName)
			m.gnode.clientMap.del(clientName)
			m.gnode.senderManager.sendMessage(&AntMes{
				Target:   "*",
				Function: "forceGC",
				Args:     []string{"true"},
			})
			m.gnode.nodeFunctions.forceGC()
			return
		}
		if err != nil {
			logf.error("Client reader %s: Failed to receive message: %v\n", clientName, err)
			m.gnode.clientMap.del(clientName)
			m.gnode.senderManager.sendMessage(&AntMes{
				Target:   "*",
				Function: "forceGC",
				Args:     []string{"true"},
			})
			m.gnode.nodeFunctions.forceGC()
			return
		}
		if mes.Function == "sendBlock" {
			if err := m.gnode.fileManager.receiveBlocFromClient(mes); err != nil {
				stream.Send(&AntMes{
					Function: "sendBlock",
					ErrorMes: err.Error(),
				})
			}
		} else {
			mes.Id = m.gnode.getNewId(false)
			mes.Origin = m.gnode.name
			mes.FromClient = clientName
			if mes.Debug {
				logf.debugMes(mes, "-------------------------------------------------------------------------------------------------------------\n")
				logf.debugMes(mes, "Receive mes from client %s : %v\n", clientName, mes)
			}
			m.gnode.idMap.Add(mes.Id)
			m.gnode.receiverManager.receiveMessage(mes)
			/*
				if !mes.ReturnAnswer {
					stream.Send(&AntMes{Origin: m.gnode.name})
				}
			*/
		}
	}
}
