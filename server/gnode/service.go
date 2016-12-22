package gnode

import (
	"fmt"
	"golang.org/x/net/context"
	"net"
)

// Send send message to grid
func (g *GNode) ExecuteFunction(ctx context.Context, mes *AntMes) (*AntRet, error) {
	if mes.Id == "" {
		mes.Id = g.getNewId(false)
		mes.Origin = g.name
		//logf.debugMes(mes, "Received message from client: %v\n", mes)
	} else {
		if ok := g.idMap.Exists(mes.Id); ok {
			//logf.info("execute store bloc ack doublon id=%s order=%d\n", mes.Id, mes.Order)
			return g.getRet(mes.Id, true), nil
		}
		g.idMap.Add(mes.Id)
	}
	if ok := g.receiverManager.receiveMessage(mes); !ok {
		logf.error("Message put error id=%s\n", mes.Id)
		return g.getRet(mes.Id, false), nil
	}
	return g.getRet(mes.Id, true), nil
}

func (g *GNode) getRet(id string, ack bool) *AntRet {
	return &AntRet{
		Id:  id,
		Ack: ack,
	}
}

func (g *GNode) Ping(ctx context.Context, mes *AntMes) (*PingRet, error) {
	return &PingRet{
		Name:         g.name,
		Host:         g.host,
		NbNode:       int32(g.nbNode),
		ClientNumber: int32(len(g.clientMap)),
	}, nil
}

func (g *GNode) GetClientStream(stream GNodeService_GetClientStreamServer) error {
	if !g.healthy {
		return fmt.Errorf("Node %s not yet ready", g.name)
	}
	g.receiverManager.startClientReader(stream)
	return nil
}

func (g *GNode) AskConnection(ctx context.Context, req *AskConnectionRequest) (*PingRet, error) {
	ret := &PingRet{
		Name: g.name,
		Host: g.host,
	}
	ip := net.ParseIP(req.Ip)
	if ip == nil {
		return nil, fmt.Errorf("IP addresse parse error %s", req.Ip)
	}
	if _, ok := g.targetMap[req.Name]; ok {
		return ret, nil
	}
	conn, err := g.startGRPCClient(ip)
	if err != nil {
		return nil, fmt.Errorf("Start GRPC error: %v", err)
	}
	client := NewGNodeServiceClient(conn)
	target := &gnodeTarget{
		name:         req.Name,
		host:         req.Host,
		ip:           req.Ip,
		conn:         conn,
		client:       client,
		updateNumber: g.updateNumber,
	}
	g.targetMap[req.Name] = target
	return ret, nil
}

func (g *GNode) StoreFile(ctx context.Context, req *StoreFileRequest) (*StoreFileRet, error) {
	return g.fileManager.storeFile(req)
}

func (g *GNode) RetrieveFile(ctx context.Context, req *RetrieveFileRequest) (*RetrieveFileRet, error) {
	return g.fileManager.retrieveFile(req)
}
