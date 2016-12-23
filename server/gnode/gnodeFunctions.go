package gnode

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

var functionMap map[string]interface{}

func initFunctionMap() {
	functionMap = make(map[string]interface{})
	functionMap["ping"] = ping
	functionMap["pingFromTo"] = pingFromTo
	functionMap["serviceInfo"] = serviceInfo
	functionMap["setLogLevel"] = setLogLevel
	functionMap["getConnections"] = getConnections
	functionMap["killNode"] = killNode
	functionMap["updateGrid"] = updateGrid
	functionMap["writeStatsInLog"] = writeStatsInLog
	functionMap["clear"] = clear
	functionMap["forceGC"] = forceGC
	functionMap["getNodeName"] = getNodeName
	functionMap["createUser"] = createUser
	functionMap["createNodeUser"] = createNodeUser
	functionMap["removeUser"] = removeUser

}

func ping(g *GNode, name string) string {
	logf.debug("execute ping from: %s\n", name)
	return fmt.Sprintf("pong from %s (%s)", g.name, g.host)
}

func pingFromTo(g *GNode, target string) string {
	logf.debug("execute pingFromTo from: %s tp %s\n", g.name, target)
	mes := NewAntMes(target, true, "ping", "pingFromTo")
	mret, err := g.senderManager.sendMessageReturnAnswer(mes, 3)
	if err != nil {
		return err.Error()
	}
	ret := ""
	for _, node := range mret.Path {
		ret += fmt.Sprintf("%s -> %s", ret, node)
	}
	return ret + " -> " + target
}

func getNodeName(g *GNode, dec int) string {
	index := g.nodeIndex + dec
	if index >= len(g.nodeNameList) {
		index = index - len(g.nodeNameList)
	}
	return g.nodeNameList[index]
}

func getConnections(g *GNode, name string) string {
	ret := fmt.Sprintf("%s (%s): ", g.name, g.host)
	for name, _ := range g.targetMap {
		ret += (" " + name)
	}
	return ret
}

func serviceInfo(g *GNode) int {
	return g.nbNode
}

func setLogLevel(g *GNode, level string) {
	logf.setLevel(level)
	logf.printf("Set log level: " + logf.levelString())
}

func killNode(g *GNode) string {
	time.AfterFunc(time.Second*3, func() {
		os.Exit(0)
	})
	return g.host
}

func updateGrid(g *GNode, force bool) {
	g.startupManager.updateGrid(false, force)
}

func writeStatsInLog(g *GNode) {
	logf.printf("IdMap size: %d", g.idMap.Len())
	g.receiverManager.stats()
	g.senderManager.stats()
}

func clear(g *GNode) {
	g.idMap.CleanUp()
	g.fileManager.transferMap = make(map[string]*FileTransfer)
	logf.info("Node cleared")
	forceGC(g, true)
	//stats := &runtime.MemStats{}
	//runtime.ReadMemStats(stats)
}

func forceGC(g *GNode, verbose bool) {
	if verbose {
		//logf.info("GC forced")
	}
	debug.FreeOSMemory()
	runtime.GC()
}

func createUser(g *GNode, userName string) string {
	logf.info("Received create user %s\n", userName)
	token := g.getToken()
	args := []string{userName, token}
	g.senderManager.sendMessage(&AntMes{
		Target:   "*",
		Origin:   g.name,
		Function: "createNodeUser",
		Args:     args,
	})
	return token
}

func createNodeUser(g *GNode, userName string, token string) string {
	err := g.createUser(userName, token)
	if err != nil {
		return err.Error()
	}
	return "done"
}

func removeUser(g *GNode, userName string) string {
	logf.info("Remove user %s\n", userName)
	return ""
}
