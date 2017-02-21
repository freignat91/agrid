package gnode

import (
	"fmt"
)

type gnodeListener struct {
	eventType string
	fileType  string
	clientMap map[string]string
}

func NewListener(eventType string) *gnodeListener {
	return &gnodeListener{
		eventType: eventType,
		clientMap: make(map[string]string),
	}
}

func (g *GNode) initEventListener() {
	g.eventListenerMap = make(map[string]*gnodeListener)
	g.eventListenerMap["TransferEvent"] = NewListener("TransferEvent")
}

func (g *GNode) setEventListener(eventType string, fileType string, userName string, clientId string) {
	logf.info("setEventListener eventType=%s fileType=%s userName=%s clientId=%s\n", eventType, fileType, userName, clientId)
	listener, ok := g.eventListenerMap[eventType]
	if !ok {
		logf.error("setEventListener eventType %s is not an eventType\n", eventType)
	}
	listener.clientMap[clientId] = userName
	listener.fileType = fileType
}

func (g *GNode) removeEventListener(clientId string) {
	logf.info("removeEventListener clientId=%s\n", clientId)
	for _, listener := range g.eventListenerMap {
		delete(listener.clientMap, clientId)
	}
}

func (g *GNode) sendBackEvent(mes *AntMes) error {
	logf.info("Received sendBackEvent: %v\n", mes.Args)
	if len(mes.Args) < 5 {
		return fmt.Errorf("sendBackEvent not enough parameters\n")
	}
	listener, ok := g.eventListenerMap[mes.Args[0]]
	if !ok {
		return fmt.Errorf("sendBackEvent eventType %s is not an eventType\n", mes.Args[0])
	}
	for clientId, userName := range listener.clientMap {
		if userName == mes.UserName && (listener.fileType == "" || mes.Args[4] == listener.fileType) {
			logf.info("sendBacked to %s\n", clientId)
			g.sendBackClient(clientId, mes)
		}
	}
	return nil
}
