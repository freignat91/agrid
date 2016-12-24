package agridapi

import (
	"fmt"
	"sort"
	"time"
)

// NodeClear force a node to clear its memory
func (api *AgridAPI) NodeClear(node string) error {
	if node == "" {
		node = "*"
	}
	client, err := api.getClient()
	if err != nil {
		return err
	}
	if _, err := client.createSendMessage(node, false, "clear"); err != nil {
		return err
	}
	return nil
}

// NodeKill force a node to kill it-self
func (api *AgridAPI) NodeKill(node string) error {
	client, err := api.getClient()
	if err != nil {
		return err
	}
	if _, err := client.createSendMessage(node, true, "killNode"); err != nil {
		return err
	}
	return nil
}

// NodePing ping a node
func (api *AgridAPI) NodePing(node string, debugTrace bool) (string, error) {
	client, err := api.getClient()
	if err != nil {
		return "", err
	}

	mes := client.createMessage(node, true, "ping", "client")
	mes.Debug = debugTrace
	mret, errs := client.sendMessage(mes, true)
	if errs != nil {
		return "", errs
	}
	ret := ""
	for _, node := range mret.Path {
		if ret == "" {
			ret = node
		} else {
			ret += fmt.Sprintf("%s -> %s", ret, node)
		}
	}
	return fmt.Sprintf("%s -> %s", ret, node), nil
}

// NodePingFrom ping a node from another node
func (api *AgridAPI) NodePingFromTo(node1 string, node2 string, debugTrace bool) (string, error) {
	client, err := api.getClient()
	if err != nil {
		return "", err
	}

	mes := client.createMessage(node1, true, "pingFromTo", node2)
	mes.Debug = api.isDebug()
	fmt.Printf("mes: %v\n", mes)
	ret, errs := client.sendMessage(mes, true)
	if errs != nil {
		return "", errs
	}
	return ret.Args[0], nil
}

// NodeSetLogLevel set a node log level: "error", "warn", "info", "debug"
func (api *AgridAPI) NodeSetLogLevel(node string, logLevel string) error {
	client, err := api.getClient()
	if err != nil {
		return err
	}
	if err := client.createSendMessageNoAnswer(node, "setLogLevel", logLevel); err != nil {
		return err
	}
	return nil
}

func (api *AgridAPI) NodeLs() ([]string, error) {
	rep := []string{}
	client, err := api.getClient()
	if err != nil {
		return rep, err
	}
	_, errp := client.createSendMessage("*", false, "getConnections")
	if errp != nil {
		return rep, errp
	}
	nb := 0
	t0 := time.Now()
	for {
		mes, ok := client.getNextAnswer(100)
		if ok {
			nb++
			rep = append(rep, mes.Args[0])
		}
		if time.Now().Sub(t0) > time.Second*5 {
			break
		}
		if nb == client.nbNode {
			break
		}
	}
	sort.Strings(rep)
	return rep, nil
}

func (api *AgridAPI) NodeUpdateGrid(node string, force bool) error {
	client, err := api.getClient()
	if err != nil {
		return err
	}
	if err := client.createSendMessageNoAnswer(node, "updateGrid", fmt.Sprintf("%t", force)); err != nil {
		return err
	}
	return nil
}
