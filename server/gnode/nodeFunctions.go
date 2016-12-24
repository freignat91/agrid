package gnode

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

type nodeFunctions struct {
	gnode *GNode
}

func (n *nodeFunctions) ping(mes *AntMes) error {
	logf.debug("execute ping from: %s\n", mes.Origin)
	answer := n.gnode.createAnswer(mes)
	answer.Args = []string{fmt.Sprintf("pong from %s (%s)", n.gnode.name, n.gnode.host)}
	n.gnode.senderManager.sendMessage(answer)
	return nil
}

func (n *nodeFunctions) pingFromTo(mes *AntMes) error {
	fmt.Printf("pingFromTo: %v\n", mes)
	if len(mes.Args) < 1 {
		return fmt.Errorf("Number of argument error, need the pingFromTo target")
	}
	fmt.Printf("args ok\n")
	target := mes.Args[0]
	logf.debug("execute pingFromTo from: %s tp %s\n", n.gnode.name, target)
	mesp := NewAntMes(target, true, "ping")
	mret, err := n.gnode.senderManager.sendMessageReturnAnswer(mesp, 3)
	if err != nil {
		return err
	}
	fmt.Printf("ping: %v\n", mret)
	ret := ""
	for _, node := range mret.Path {
		if ret == "" {
			ret = node
		} else {
			ret += fmt.Sprintf("%s -> %s", ret, node)
		}
	}
	ret += " -> " + target
	answer := n.gnode.createAnswer(mes)
	answer.Args = []string{ret}
	fmt.Printf("answer: %v\n", answer)
	n.gnode.senderManager.sendMessage(answer)
	return nil
}

func (n *nodeFunctions) setLogLevel(mes *AntMes) error {
	if len(mes.Args) < 1 {
		return fmt.Errorf("Number of argument error, need logLevel")
	}
	logf.setLevel(mes.Args[0])
	logf.printf("Set log level: " + logf.levelString())
	return nil
}

func (n *nodeFunctions) killNode(mes *AntMes) error {
	time.AfterFunc(time.Second*3, func() {
		os.Exit(0)
	})
	return nil
}

func (n *nodeFunctions) updateGrid(mes *AntMes) error {
	force := false
	if len(mes.Args) >= 1 && mes.Args[0] == "true" {
		force = true
	}
	n.gnode.startupManager.updateGrid(false, force)
	return nil
}

func (n *nodeFunctions) writeStatsInLog(mes *AntMes) error {
	logf.printf("IdMap size: %d", n.gnode.idMap.Len())
	n.gnode.receiverManager.stats()
	n.gnode.senderManager.stats()
	return nil
}

func (n *nodeFunctions) getConnections(mes *AntMes) error {
	ret := fmt.Sprintf("%s (%s): ", n.gnode.name, n.gnode.host)
	for name, _ := range n.gnode.targetMap {
		ret += (" " + name)
	}
	answer := n.gnode.createAnswer(mes)
	answer.Args = []string{ret}
	n.gnode.senderManager.sendMessage(answer)
	return nil
}

func (n *nodeFunctions) clear(mes *AntMes) error {
	n.gnode.idMap.CleanUp()
	n.gnode.fileManager.transferMap = make(map[string]*FileTransfer)
	logf.info("Node cleared")
	n.forceGC()
	return nil
}

func (g *nodeFunctions) forceGC() {
	debug.FreeOSMemory()
	runtime.GC()
}

func (n *nodeFunctions) createUser(mes *AntMes) error {
	if len(mes.Args) < 1 {
		return fmt.Errorf("Number of argument error, need userName")
	}
	userName := mes.Args[0]
	token := ""
	if len(mes.Args) >= 2 {
		token = mes.Args[1]
	}
	logf.info("Received create user %s\n", userName)
	if token == "" {
		token = n.gnode.getToken()
	}
	args := []string{userName, token}
	err := n.gnode.createUser(userName, token)
	if err != nil {
		return err
	}
	n.gnode.senderManager.sendMessage(&AntMes{
		Target:   "*",
		Origin:   n.gnode.name,
		Function: "createNodeUser",
		Args:     args,
	})
	answer := n.gnode.createAnswer(mes)
	answer.Args = []string{token}
	n.gnode.senderManager.sendMessage(answer)
	return nil
}

func (n *nodeFunctions) createNodeUser(mes *AntMes) error {
	if len(mes.Args) < 2 {
		return fmt.Errorf("Number of argument error, need userName, token")
	}
	userName := mes.Args[0]
	token := mes.Args[1]
	err := n.gnode.createUser(userName, token)
	if err != nil {
		return err
	}
	answer := n.gnode.createAnswer(mes)
	answer.Args = []string{token}
	n.gnode.senderManager.sendMessage(answer)
	return nil
}

func (n *nodeFunctions) removeUser(mes *AntMes) error {
	return fmt.Errorf("remove user not implemented yet")
}
