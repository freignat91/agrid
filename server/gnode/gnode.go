package gnode

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"os"
	"sync"
	"time"
)

//TODO: ramasse miette sur les objet transfer non terminé apres moult

var (
	config GNodeConfig     = GNodeConfig{}
	ctx    context.Context = context.Background()
)

type GNode struct {
	host            string
	selfIP          *net.IP
	name            string
	nodeIndex       int
	conn            *grpc.ClientConn
	nbNode          int
	connectReady    bool
	targetMap       map[string]*gnodeTarget
	clientMap       map[string]*gnodeClient
	receiverManager ReceiverManager
	senderManager   SenderManager
	startupManager  *gnodeLeader
	mesNumber       int
	lastIndexTime   time.Time
	healthy         bool
	traceMap        map[string]*gnodeTrace
	nbRouted        int64
	idMap           gnodeIdMap
	nodeNameList    []string
	logMode         int
	updateNumber    int
	reduceMode      bool
	fileManager     *FileManager
	lockId          sync.RWMutex
	dataPath        string
}

type gnodeTarget struct {
	ready        bool
	closed       bool
	ip           string
	name         string
	host         string
	updateNumber int
	client       GNodeServiceClient
	conn         *grpc.ClientConn
	from         bool
}

type gnodeClient struct {
	name   string
	stream GNodeService_GetClientStreamServer
	usage  int
}

type gnodeTrace struct {
	creationTime time.Time
	nbUsed       int
	persistence  int
	target       *gnodeTarget
}

// Start gnode
func (g *GNode) Start(version string, build string) error {
	config.init(version, build)
	g.init()
	g.startupManager = &gnodeLeader{}
	if _, err := g.startupManager.init(g); err != nil {
		return err
	}
	for {
		//
		time.Sleep(3000 * time.Second)
	}

}

func (g *GNode) init() {
	g.lockId = sync.RWMutex{}
	g.traceMap = make(map[string]*gnodeTrace)
	g.clientMap = make(map[string]*gnodeClient)
	g.targetMap = make(map[string]*gnodeTarget)
	g.nbNode = config.nbNode
	g.dataPath = "/data"
	g.idMap.Init()
	g.fileManager = &FileManager{}
	g.fileManager.init(g)
	g.startRESTAPI()
	g.startGRPCServer()
	g.receiverManager.start(g, config.bufferSize, config.parallelReceiver)
	g.senderManager.start(g, config.bufferSize, config.parallelSender)
	g.host = os.Getenv("HOSTNAME")
	initFunctionMap()
	time.Sleep(3 * time.Second)
}

func (g *GNode) startGRPCServer() {
	s := grpc.NewServer()
	RegisterGNodeServiceServer(s, g)
	go func() {
		lis, err := net.Listen("tcp", ":"+config.grpcPort)
		if err != nil {
			logf.error("gnode is unable to listen on: %s\n%v", ":"+config.grpcPort, err)
		}
		logf.info("gnode is listening on port %s\n", ":"+config.grpcPort)
		if err := s.Serve(lis); err != nil {
			logf.error("Problem in gnode server: %s\n", err)
		}
	}()
}

func (g *GNode) clearConnection() {
	for _, target := range g.targetMap {
		target.closed = true
		if target.conn != nil {
			target.conn.Close()
		}
	}
	g.targetMap = make(map[string]*gnodeTarget)
	logf.printf("connections closed")
}

func (g *GNode) setSelfName(ip *net.IP, name string) {
	g.selfIP = ip
	g.name = name
}

func (g *GNode) connectTarget(updateNumber int, nodeName string, nodeIP net.IP) error {
	if targetOld, ok := g.targetMap[nodeName]; ok {
		targetOld.updateNumber = updateNumber
		logf.info("Still connected to %s (%s)\n", targetOld.name, targetOld.host)
		return nil
	}
	conn, err := g.startGRPCClient(nodeIP)
	if err != nil {
		return err
	}
	client := NewGNodeServiceClient(conn)
	ret, err2 := client.Ping(ctx, &AntMes{})
	/*
		ret, err2 := client.AskConnection(ctx, &AskConnectionRequest{
			Name: g.name,
			Host: g.host,
			Ip:   g.selfIP.String(),
		})
	*/
	if err2 != nil {
		return err2
	}
	target := &gnodeTarget{
		from:         true,
		name:         nodeName,
		host:         ret.Host,
		ip:           nodeIP.String(),
		client:       client,
		conn:         conn,
		updateNumber: updateNumber,
	}
	g.targetMap[nodeName] = target
	logf.info("Connected to %s (%s)\n", target.name, target.host)
	return nil
}

func (g *GNode) removeObsoletTarget(updateNumber int) {
	tmap := make(map[string]*gnodeTarget)
	for name, target := range g.targetMap {
		if target.updateNumber == updateNumber {
			tmap[name] = target
		} else {
			logf.info("Remove target %s (%s)\n", target.name, target.host)
			g.closeTarget(target)
		}
	}
	g.targetMap = tmap
}

func (g *GNode) closeTarget(target *gnodeTarget) {
	if target.conn != nil {
		target.conn.Close()
	}
	target.closed = true
	delete(g.targetMap, target.name)
}

func (g *GNode) displayConnection() {
	logf.printf("---------------------------------------------------------------------------------------\n")
	logf.printf("Node: %s\n", g.name)
	for _, target := range g.targetMap {
		logf.printf("Connected -> %s ip: %s (%s)\n", target.name, target.ip, target.host)
	}
	logf.printf("---------------------------------------------------------------------------------------\n")
}

// Connect to server
func (g *GNode) startGRPCClient(ip net.IP) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("%s:%s", ip.String(), config.grpcPort),
		grpc.WithInsecure(),
		grpc.WithBlock())
	//grpc.WithTimeout(time.Second*60))
}

func (g *GNode) getNewId(setAsAlreadySent bool) string {
	g.lockId.Lock()
	defer g.lockId.Unlock()
	g.mesNumber++
	id := fmt.Sprintf("%s-%d", g.host, g.mesNumber)
	if setAsAlreadySent {
		g.idMap.Add(id)
	}
	return id
}

func (g *GNode) sendBackClient(clientId string, mes *AntMes) {
	client, ok := g.clientMap[clientId]
	if !ok {
		logf.error("Send to client error: client %s doesn't exist mes=%v", clientId, mes.Id)
		return
	}
	//mes.Origin = g.name
	client.usage++
	if err := client.stream.Send(mes); err != nil {
		logf.error("Error trying to send message to client %s: mes=%s: %s\n", clientId, mes.toString(), err)
	}
	if client.usage%100 == 0 {
		//Seams to have a bug in grpc cg
		g.senderManager.sendMessage(&AntMes{
			Target:   "*",
			Function: "forceGC",
			Args:     []string{"false"},
		})
		forceGC(g, false)
	}

}

func (g *GNode) startReorganizer() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			g.fileManager.moveRandomBlock()
		}
	}()
}
