package gnode

import (
	"fmt"
	"strings"
	"time"
)

type MessageSender struct {
	usage         int
	gnode         *GNode
	senderManager *SenderManager
	id            int
}

func (s *MessageSender) start() {
	go func() {
		for {
			mes := <-s.senderManager.ioChan
			if mes != nil {
				s.usage++
				if err := s.sendMessage(mes); err != nil {
					logf.error("sendMessage error id=%s: %v", mes.Id, err)
				}
				mes = nil
			}
		}
	}()
}

// route message
func (s *MessageSender) sendMessage(mes *AntMes) error {
	//logf.info("sendMessageEff: %s\n", mes.toString())
	if mes.Id == "" {
		mes.Id = s.gnode.getNewId(true)
		mes.Origin = s.gnode.name
	}
	if mes.Target == s.gnode.name {
		s.gnode.receiverManager.receiveMessage(mes)
		return nil
	}
	if !s.gnode.connectReady {
		logf.error("Connections not ready abort send\n")
		return nil
	}
	if mes.IsAnswer {
		logf.debugMes(mes, "Send answer: %v\n", mes)
		if err := s.sendBackMesFollowingPath(mes); err != nil {
			logf.error("Answer id=%s error: %v\n", mes.Id, err)
			return err
		}
		return nil
	}
	mes.Path = append(mes.Path, s.gnode.name)
	mes.PathIndex = int32(len(mes.Path) - 1)
	//logf.debug(mes, "Send message id=%s: %+v\n", mes.Id, mes)
	if mes.Target == "*" {
		return s.broadcastMes(mes)
	}
	if target, ok := s.gnode.targetMap[mes.Target]; ok {
		logf.debugMes(mes, "Send message direc to target id=%s\n", mes.Id)
		return s.sendToTarget(target, mes)
	}
	if trace, ok := s.gnode.traceMap[mes.Target]; ok {
		trace.nbUsed++
		logf.debugMes(mes, "Use trace on target %s : %d\n", trace.target.name, trace.nbUsed)
		//logf.info("Use trace on target %s : %d\n", trace.target.name, trace.nbUsed)
		return s.sendToTarget(trace.target, mes)
	}
	return s.broadcastMes(mes)
}

// Send message using all targets
func (s *MessageSender) broadcastMes(mes *AntMes) error {
	logf.debugMes(mes, "broadcast message id=%s\n", mes.Id)
	mes.IsPathWriter = true
	for _, target := range s.gnode.targetMap {
		s.sendToTarget(target, mes)
	}
	return nil
}

// find the target corresponding to a message path and send the mes to it.
func (s *MessageSender) sendBackMesFollowingPath(mes *AntMes) error {
	logf.debugMes(mes, "sendBackAnswerFolowingPath id=%s indexr=%d, path%v\n", mes.Id, mes.GetIsAnswer, mes.Path)
	if mes.PathIndex < 0 || int(mes.PathIndex) >= len(mes.Path) {
		return fmt.Errorf("PathIndex error: %v", mes)
	}
	nextTarget := mes.Path[mes.PathIndex]
	logf.debugMes(mes, "Next target=%s\n", nextTarget)
	mes.PathIndex = mes.PathIndex - 1
	for _, target := range s.gnode.targetMap {
		if target.name == nextTarget {
			if mes.IsPathWriter {
				s.updateTrace(mes)
			}
			s.sendToTarget(target, mes)
			return nil
		}
	}
	return fmt.Errorf("Local target %s not found on %s", nextTarget, s.gnode.name)
}

// send the message using one specific target
func (s *MessageSender) sendToTarget(target *gnodeTarget, mes *AntMes) error {
	logf.debugMes(mes, "Send message %s to target %s (%s)\n", mes.Id, target.name, target.host)
	err := target.sendMessage(mes)
	if err != nil {
		if strings.Index(err.Error(), "the connection is unavailable") >= 0 {
			logf.info("Connection is broken with %s (%s)\n", target.name, target.host)
			s.gnode.closeTarget(target)
			if s.gnode.reduceMode {
				s.gnode.startupManager.sendUpdateGrid()
			}
			return err
		}
		return fmt.Errorf("Message id=%s route gnode %s ExecuteFunction error: %v\n", mes.Id, target.name, err)
	}
	return nil
}

// add a trace giving the direction (a local target) to reach a target (the Origin)
func (s *MessageSender) updateTrace(mes *AntMes) {
	logf.debugMes(mes, "Updating trace with mes %v\n", mes)
	target := mes.Origin
	localTargetName := ""
	for i, node := range mes.Path {
		if node == s.gnode.name {
			if i+1 < len(mes.Path) {
				localTargetName = mes.Path[i+1]
			}
		}
	}
	if localTargetName == "" {
		//logf.warn("updatePath, target=%s: impossible to find the localTarget: path=%v\n", target, mes.Path)
		return
	}
	localTarget, ok := s.gnode.targetMap[localTargetName]
	if !ok {
		logf.warn("Local target %s doesn't exist locally %s\n", mes.Path[1])
		return
	}
	if trace, ok := s.gnode.traceMap[target]; ok {
		logf.debugMes(mes, "Confirm trace for target %s using local target %s : %d\n", target, localTarget.name, trace.persistence)
		trace.persistence--
		if trace.persistence <= 0 {
			delete(s.gnode.traceMap, target)
		}
		return
	}
	logf.debugMes(mes, "create trace for target %s using local target %s\n", target, localTarget.name)
	s.gnode.traceMap[target] = &gnodeTrace{
		creationTime: time.Now(),
		persistence:  config.tracePersistence,
		target:       localTarget,
	}
}

func (t *gnodeTarget) sendMessage(mes *AntMes) error {
	if _, err := t.client.ExecuteFunction(ctx, mes); err != nil {
		logf.error("Send message error executeFunction error: %s to %s: %v\n", mes.Id, t.name, err)
		return err
	}
	return nil
}
