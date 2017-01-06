package gnode

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path"
	"sync"
	"time"
)

//TODO: ramasse miette sur les objet transfer non terminé apres moult
//TODO: si duplicate x utilisé alors recréer duplicate x-1 si x>1
//TODO: passer toutes les requetes file par service
//TODO: utiliser targettedPath plutot que args[0] (nottament pour file rm)
//TODO: mettre implicite le received local pour les target *
//TODO: avoir un cancel global du transfer dsur transferId sur carch de noeud recepteru ou cancel client

var (
	config GNodeConfig     = GNodeConfig{}
	ctx    context.Context = context.Background()
)

type GNode struct {
	host              string
	selfIP            *net.IP
	name              string
	nodeIndex         int
	conn              *grpc.ClientConn
	nbNode            int
	connectReady      bool
	targetMap         map[string]*gnodeTarget
	clientMap         secureMap //map[string]*gnodeClient
	receiverManager   ReceiverManager
	senderManager     SenderManager
	startupManager    *gnodeLeader
	mesNumber         int
	lastIndexTime     time.Time
	healthy           bool
	traceMap          map[string]*gnodeTrace
	nbRouted          int64
	idMap             gnodeIdMap
	nodeNameList      []string
	logMode           int
	updateNumber      int
	reduceMode        bool
	fileManager       *FileManager
	lockId            sync.RWMutex
	dataPath          string
	nodeFunctions     *nodeFunctions
	userMap           map[string]string
	availableNodeList []string
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
	//g.clientMap = make(map[string]*gnodeClient)
	g.clientMap.init()
	g.targetMap = make(map[string]*gnodeTarget)
	g.nbNode = config.nbNode
	g.dataPath = config.rootDataPath
	g.loadUser()
	g.idMap.Init()
	g.nodeFunctions = &nodeFunctions{gnode: g}
	g.fileManager = &FileManager{}
	g.fileManager.init(g)
	g.startRESTAPI()
	g.startGRPCServer()
	g.receiverManager.start(g, config.bufferSize, config.parallelReceiver)
	g.senderManager.start(g, config.bufferSize, config.parallelSender)
	g.host = os.Getenv("HOSTNAME")
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

func (g *GNode) updateLocalNodeList() {
	list := []string{}
	for name, target := range g.targetMap {
		if g.isTargetAvailable(target) {
			list = append(list, name)
		}
	}
	g.availableNodeList = list
}

func (g *GNode) isTargetAvailable(target *gnodeTarget) bool {
	if _, err := target.client.Healthcheck(ctx, &HealthRequest{}); err != nil {
		return false
	}
	return true
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

func (g *GNode) createAnswer(mes *AntMes, withNodeList bool) *AntMes {
	ans := &AntMes{
		Function:     fmt.Sprintf("answer-%s", mes.Function),
		Target:       mes.Origin,
		OriginId:     mes.Id,
		FromClient:   mes.FromClient,
		IsAnswer:     true,
		Path:         mes.Path,
		PathIndex:    int32(len(mes.Path) - 1),
		ReturnAnswer: false,
		Debug:        mes.Debug,
		IsPathWriter: mes.IsPathWriter,
		AnswerWait:   mes.AnswerWait,
	}
	if withNodeList {
		ans.Nodes = g.availableNodeList
	}
	return ans
}

func (g *GNode) sendBackClient(clientId string, mes *AntMes) {
	//logf.info("sendBackClient tf=%s order=%d\n", mes.TransferId, mes.Order)
	if !g.clientMap.exists(clientId) {
		logf.error("Send to client error: client %s doesn't exist mes=%v", clientId, mes.Id)
		return
	}
	client := g.clientMap.get(clientId).(*gnodeClient)
	client.usage++
	if client.usage%100 == 0 {
		//Seams to have a bug in grpc cg
		g.nodeFunctions.forceGC()
	}
	//logf.info("sendBackClient eff tf=%s order=%d\n", mes.TransferId, mes.Order)
	if err := client.stream.Send(mes); err != nil {
		logf.error("Error trying to send message to client %s: mes=%s: %s\n", clientId, mes.toString(), err)
	}
}

func (g *GNode) startReorganizer() {
	go func() {
		nn := 0
		for {
			time.Sleep(10 * time.Second)
			g.nodeFunctions.forceGC()
			nn++
			if nn == 3 {
				nn = 0
				g.updateLocalNodeList()
				g.fileManager.moveRandomBlock()
			}
		}
	}()
}

func (g *GNode) getToken() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (g *GNode) createUser(userName string, token string) error {
	logf.info("Create user %s\n", userName)
	_, err := ioutil.ReadDir(path.Join(config.rootDataPath, userName))
	if err == nil {
		g.loadOneUser(userName)
		return fmt.Errorf("User %s : already exist", userName)
	}
	os.MkdirAll(path.Join(config.rootDataPath, userName), os.ModeDir)
	file, errc := os.Create(path.Join(config.rootDataPath, userName, "token"))
	if errc != nil {
		return errc
	}
	if _, err := file.WriteString(token); err != nil {
		return err
	}
	file.Close()
	logf.info("Save user: %s:[%s]\n", userName, token)
	g.userMap[userName] = token
	return nil
}

func (g *GNode) removeUser(userName string, token string, force bool) error {
	if !g.checkUser(userName, token) {
		return fmt.Errorf("Invalid user/token")
	}
	logf.warn("Remove user %s force mode=%t\n", userName, force)
	dir := path.Join(config.rootDataPath)
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	exist := false
	for _, fd := range fileList {
		if fd.Name() != "token" && fd.Name() != "meta" {
			exist = true
			break
		}
	}
	if exist && !force {
		return fmt.Errorf("Impossible to remove user %s files still exist (use --force", userName)
	}
	if err := os.RemoveAll(path.Join(config.rootDataPath, userName)); err != nil {
		return err
	}
	delete(g.userMap, userName)
	return nil
}

func (g *GNode) loadUser() error {
	g.userMap = make(map[string]string)
	fileList, err := ioutil.ReadDir(config.rootDataPath)
	if err != nil {
		return err
	}
	for _, fd := range fileList {
		g.loadOneUser(fd.Name())
	}
	return nil
}

func (g *GNode) loadOneUser(name string) {
	data, err := ioutil.ReadFile(path.Join(config.rootDataPath, name, "token"))
	if err != nil {
		logf.error("loadOneUser user=%s: %v\n", name, err)
		return
	}
	token := string(data)
	logf.info("Add user %s [%s]\n", name, token)
	g.userMap[name] = token
}

func (g *GNode) checkUser(user string, token string) bool {
	if user == "" {
		return false
	}
	if user == "common" {
		return true
	}
	check, ok := g.userMap[user]
	if !ok {
		logf.info("Check user %s: false\n", user)
		return false
	}
	if token == check {
		return true
	}
	logf.info("Check user %s false\n", user)
	return false
}
