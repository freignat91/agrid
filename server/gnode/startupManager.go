package gnode

import (
	"fmt"
	"log"
	"net"
	"sort"
	//"strconv"
	"strings"
	"time"
)

type gnodeLeader struct {
	gnode            *GNode
	udpServer        *udpServer
	leaderIP         *net.IP
	selfIP           *net.IP
	commonIP         *net.IP
	nodeIPList       []net.IP
	nodeAckGridMap   map[string]string
	nodeClearGridMap map[string]string
	leader           bool
	lastAck          *time.Time
	startupInit      bool
	updateNumber     int
	nbLineConnect    int
	nbCrossConnect   int
	ref              [][2]int
}

func (g *gnodeLeader) init(gnode *GNode) (bool, error) {
	g.startupInit = true
	g.gnode = gnode
	g.udpServer = &udpServer{}
	if err := g.udpServer.start(g); err != nil {
		return false, err
	}
	if err := g.getSelfAddr(); err != nil {
		return false, err
	}
	g.computeGrid()
	return g.leader, g.waitReady()

}

func (g *gnodeLeader) getSelfAddr() error {
	err := g.getCommonAddr()
	if err != nil {
		return err
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("getSelfAddr net.Interface error: %v", err)
	}
	var selfInterface net.Interface
	found := false
	for _, inter := range interfaces {
		addrList, err := inter.Addrs()
		if err != nil {
			return fmt.Errorf("getSelfAddr inter.Addrs error: %v", err)
		}
		for _, addr := range addrList {
			if strings.HasPrefix(addr.String(), g.commonIP.String()) {
				selfInterface = inter
				found = true
				break
			}
		}
	}
	if !found {
		return fmt.Errorf("Error common ip interface not found")
	}
	var selfAddr net.Addr
	found = false
	addrList, err := selfInterface.Addrs()
	if err != nil {
		return fmt.Errorf("Error getting addrs from selfInterface")
	}
	log.Printf("SelfInterface addresses: %v\n", addrList)
	for _, addr := range addrList {
		if strings.HasPrefix(addr.String(), g.commonIP.String()[0:4]) {
			selfAddr = addr
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Error no self addr found")
	}
	ips := strings.Split(selfAddr.String(), "/")
	if len(ips) == 0 {
		return fmt.Errorf("Error spliting self addr: %s: ", selfAddr.String())
	}
	ip := net.ParseIP(ips[0])
	if ip == nil {
		return fmt.Errorf("Error parcing self addr: %s: ", selfAddr.String())
	}
	g.gnode.setSelfName(&ip, g.buildNodeName(ip))
	g.selfIP = &ip
	log.Printf("Self IP: %v\n", g.selfIP)
	return nil
}

func (g *gnodeLeader) getCommonAddr() error {
	ips, err := net.LookupIP("agrid")
	if err != nil || len(ips) == 0 {
		return fmt.Errorf("getCommonAddr error in lookipIP: %v", err)
	}
	g.commonIP = &ips[0]
	log.Printf("common IP: %s\n", g.commonIP.String())
	return nil
}

func (g *gnodeLeader) waitReady() error {
	for !g.gnode.healthy {
		time.Sleep(3 * time.Second)
	}
	return nil
}

func (g *gnodeLeader) computeGrid() error {
	nb := 0
	logf.info("Building grid\n")
	for {
		ipcs, err := net.LookupIP("tasks.agrid")
		if err != nil || len(ipcs) == 0 {
			return fmt.Errorf("getCommonAddr error in tasks lookipIP: %v", err)
		}
		//logf.info("IPlist: %v\n", ipcs)
		if nb == len(ipcs) {
			g.nodeIPList = ipcs
			break
		}
		nb = len(ipcs)
		time.Sleep(20 * time.Second)
	}
	g.gnode.nbNode = len(g.nodeIPList)
	//logf.info("IPlist: %v\n", g.nodeIPList)

	g.updateNumber++
	g.gnode.healthy = false
	g.setIpList()
	//g.sendUpdateGrid()
	g.connectNodes()
	return nil
}

func (g *gnodeLeader) setIpList() {
	ipList := gIpList{list: g.nodeIPList}
	sort.Sort(ipList)
	logf.info("IPlist: %v\n", g.nodeIPList)
	g.gnode.nodeNameList = []string{}
	for i, ip := range g.nodeIPList {
		name := g.buildNodeName(ip)
		if name == g.gnode.name {
			g.gnode.nodeIndex = i
			g.gnode.dataPath = config.rootDataPath //fmt.Sprintf("%s/node%d", config.rootDataPath, i)
			logf.info("Set dataPath: %s\n", g.gnode.dataPath)
		}
		g.gnode.nodeNameList = append(g.gnode.nodeNameList, name)
	}
	logf.info("NodeNamelist: %v\n", g.gnode.nodeNameList)
	logf.info("Node index: %d\n", g.gnode.nodeIndex)
	if config.nbDuplicate > len(g.gnode.nodeNameList) {
		config.nbDuplicate = len(g.gnode.nodeNameList)
		logf.info("Nb duplicate set to %d\n", config.nbDuplicate)
	}
}

func (g *gnodeLeader) connectNodes() {
	g.gnode.updateNumber = g.updateNumber
	logf.info("connecNodes nbLineConnect=%d, nbCrossConnect=%d\n", g.nbLineConnect, g.nbCrossConnect)
	grid := CreateGrid(g.gnode.nbNode, g.nbLineConnect, g.nbCrossConnect, true)
	connectionArray := grid.Nodes[g.gnode.nodeIndex]
	for _, node := range connectionArray {
		ip := g.nodeIPList[node]
		name := g.buildNodeName(ip)
		if err := g.gnode.connectTarget(g.updateNumber, name, ip); err != nil {
			logf.error("Connection to %s error: %v\n", name, err)
		}
	}

	g.gnode.connectReady = true
	g.gnode.healthy = true
	g.startupInit = false
	time.AfterFunc(time.Second*20, func() {
		//g.gnode.removeObsoletTarget(g.updateNumber)
		g.gnode.displayConnection()
	})
}

func (g *gnodeLeader) buildNodeName(ip net.IP) string {
	name := "N"
	list := strings.Split(ip.String(), ".")
	for _, val := range list {
		if len(val) == 1 {
			val = "00" + val
		} else if len(val) == 2 {
			val = "0" + val
		}
		name += val
	}
	return name
	//return fmt.Sprintf("g-%s", ip.String())
}

func (g *gnodeLeader) sendUpdateGrid() {
	mes := CreateMessage("*", true, "updateGrid", "false")
	g.gnode.senderManager.sendMessage(mes)
}

func (g *gnodeLeader) updateGrid(wait bool, force bool) {
	logf.printf("ok\n")
	if g.startupInit {
		return
	}
	logf.printf("updateGrid\n")
	g.startupInit = true
	g.gnode.healthy = false
	if !wait {
		g.computeGrid()
	} else {
		time.AfterFunc(time.Second*60, func() {
			g.computeGrid()
		})
	}
}

//---------------------------------------------------------------------------------------------------------------
// net.IP list sort

type gIpList struct {
	list []net.IP
}

func (a gIpList) Len() int {
	return len(a.list)
}

func (a gIpList) Swap(i, j int) {
	a.list[i], a.list[j] = a.list[j], a.list[i]
}

func (a gIpList) Less(i, j int) bool {
	ret := strings.Compare(a.list[i].String(), a.list[j].String())
	if ret == 0 {
		return false
	} else if ret == -1 {
		return true
	}
	return false
}
