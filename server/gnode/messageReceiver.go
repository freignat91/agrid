package gnode

import (
	"reflect"
	"time"
)

type MessageReceiver struct {
	usage           int
	gnode           *GNode
	receiverManager *ReceiverManager
	id              int
}

func (r *MessageReceiver) start() {
	go func() {
		//log.Printf("Executor start: %d\n", e.id)
		for {
			//log.Printf("Executor %d wait for chan\n", e.id)
			mes := <-r.receiverManager.ioChan
			//log.Printf("Executor %d chan: %v\n", e.id, mes)
			if mes != nil {
				r.usage++
				reached, stop := r.targetReached(mes)
				if reached {
					if mes.IsAnswer {
						r.receiveAnswer(mes)
						return
					}
					r.executeMessage(mes)
				}
				if !stop {
					r.gnode.senderManager.sendMessage(mes)
				}
			}
		}
	}()
}

func (r *MessageReceiver) targetReached(mes *AntMes) (bool, bool) {
	if mes.Target == "*" {
		return true, false
	}
	if mes.Target == "" {
		return true, true
	}
	if r.gnode.name == mes.Target {
		return true, true
	}
	return false, false
}

func (r *MessageReceiver) executeMessage(mes *AntMes) {
	//Internal functions format: function(mes *AntMes) error
	if function, ok := r.receiverManager.functionMap[mes.Function]; ok {
		f := reflect.ValueOf(function)
		//logf.info("Execute function: %s\n", mes.Function)
		ret := f.Call([]reflect.Value{reflect.ValueOf(mes)})
		//logf.info("function: %s return: %v\n", mes.Function, ret)
		if ret[0].Interface() != nil {
			err := ret[0].Interface().(error)
			answer := r.gnode.createAnswer(mes)
			answer.ErrorMes = err.Error()
			r.gnode.senderManager.sendMessage(answer)
		}
	}
}

func (r *MessageReceiver) receiveAnswer(mes *AntMes) {
	logf.debugMes(mes, "Receive answer: %v\n", mes)
	if mes.IsPathWriter {
		r.updateTrace(mes)
	}
	if mes.Target == r.gnode.name {
		logf.debugMes(mes, "answer reached its target: %v\n", mes.Id)
		if mes.FromClient != "" {
			if r.gnode.clientMap.exists(mes.FromClient) {
				client := r.gnode.clientMap.get(mes.FromClient).(gnodeClient)
				if err := client.stream.Send(mes); err != nil {
					logf.error("Send back answer to client error: %v\n", err)
					return
				}
				logf.debugMes(mes, "answer id %s sent back to client %s\n", mes.Id, client.name)
			} else {
				logf.debugMes(mes, "answer id %s sent back, client %s not found\n", mes.Id, mes.FromClient)
			}
		}
		if mes.AnswerWait {
			logf.debugMes(mes, "answer originId=%s saved in receiveMap\n", mes.OriginId)
			r.receiverManager.answerMap[mes.OriginId] = mes
			r.receiverManager.getChan <- mes.OriginId
		}
	}
}

// add a trace giving the direction (a local target) to reach a target (the Origin)
func (r *MessageReceiver) updateTrace(mes *AntMes) {
	logf.debugMes(mes, "Updating trace with mes %v\n", mes)
	target := mes.Origin
	if len(mes.Path) < 2 {
		//logf.warn("Path too short to update trace\n")
		return
	}
	localTarget, ok := r.gnode.targetMap[mes.Path[1]]
	if !ok {
		logf.warn("Local target %s doesn't exist locally %s\n", mes.Path[1])
		return
	}
	if trace, ok := r.gnode.traceMap[target]; ok {
		logf.debugMes(mes, "Confirm trace for target %s using local target %s : %d\n", target, localTarget.name, trace.persistence)
		trace.persistence--
		if trace.persistence <= 0 {
			delete(r.gnode.traceMap, target)
		}
		return
	}
	logf.debugMes(mes, "create trace for target %s using local target %s\n", target, localTarget.name)
	r.gnode.traceMap[target] = &gnodeTrace{
		creationTime: time.Now(),
		persistence:  config.tracePersistence,
		target:       localTarget,
	}

}
