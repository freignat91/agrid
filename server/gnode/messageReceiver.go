package gnode

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
				r.executeMessage(mes)
				mes = nil
			}
		}
	}()
}

func (r *MessageReceiver) executeMessage(mes *AntMes) {
	//logf.printf("execute message: %s\n", mes.toString())
	r.usage++
	reached, stop := r.targetReached(mes)
	if reached {
		if mes.IsAnswer {
			r.receiveAnswer(mes)
			return
		}
		//Special file function
		if mes.Function == "storeBlock" {
			if err := r.gnode.fileManager.storeBlock(mes); err != nil {
				logf.error("Store block error: %v\n", err)
			}
			return
		} else if mes.Function == "storeBlocAck" {
			if err := r.gnode.fileManager.storeBlockAck(mes); err != nil {
				logf.error("Store block ack error: %v\n", err)
			}
			return
		} else if mes.Function == "getFileBlocks" {
			if err := r.gnode.fileManager.sendBlock(mes); err != nil {
				logf.error("Send block error: %v\n", err)
			}
			return
		} else if mes.Function == "sendBackBlock" {
			if err := r.gnode.fileManager.receivedBackBlock(mes); err != nil {
				logf.error("ReveiveBackBlock error: %v\n", err)
			}
			return
		}
		logf.debugMes(mes, "Executor %d:Received mes: %+v\n", r.id, mes)
		if err := r.executeFunction(mes); err != nil {
			serr := fmt.Sprintf("Error executing function %s with message: %v error: %v\n", mes.Function, mes, err)
			logf.error(serr)
			if mes.FromClient != "" {
				ret := make([]reflect.Value, 1, 1)
				ret[0] = reflect.ValueOf(serr)
				r.sendAnswer(mes, ret)
			}
		}
		if stop {
			return
		}
	}
	if !stop {
		r.gnode.senderManager.sendMessage(mes)
	}
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

func (r *MessageReceiver) executeFunction(mes *AntMes) error {
	//logf.info("Executing function %s with args %v\n", mes.Function, mes.Args)
	fc, ok := functionMap[mes.Function]
	if !ok {
		return fmt.Errorf("The function %s does not exist\n", mes.Function)
	}

	f := reflect.ValueOf(fc)
	if len(mes.Args) != f.Type().NumIn()-1 {
		return fmt.Errorf("The number of params for function %s is not adapted function=%d received=%d\n", mes.Function, f.Type().NumIn()-1, len(mes.Args))
	}
	in, err := r.convertFunctionArgs(f, mes.Function, mes.Args)
	if err != nil {
		return fmt.Errorf("The function execution error: %v\n", err)
	}
	ret := f.Call(in)
	logf.debugMes(mes, "Function %s executed, return %v\n", mes.Function, ret)
	if mes.ReturnAnswer {
		r.sendAnswer(mes, ret)
	}
	return nil
}

func (r *MessageReceiver) sendAnswer(mes *AntMes, ret []reflect.Value) error {
	args := []string{}
	for _, val := range ret {
		arg, err := r.marshal(val)
		if err != nil {
			return err
		}
		args = append(args, arg)
	}
	retMes := &AntMes{
		Id:           fmt.Sprintf("answer-%s-%s", r.gnode.name, mes.Id),
		Origin:       r.gnode.name,
		Target:       mes.Origin,
		FromClient:   mes.FromClient,
		Path:         mes.Path,
		PathIndex:    int32(len(mes.Path) - 1),
		IsAnswer:     true,
		ReturnAnswer: false,
		OriginId:     mes.Id,
		Function:     mes.Function,
		Debug:        mes.Debug,
		IsPathWriter: mes.IsPathWriter,
		AnswerWait:   mes.AnswerWait,
		Args:         args,
	}
	if retMes.Target == r.gnode.name {
		r.receiveAnswer(retMes)
	} else {
		r.gnode.senderManager.sendMessage(retMes)
	}
	return nil
}

func (r *MessageReceiver) receiveAnswer(mes *AntMes) {
	logf.debugMes(mes, "Receive answer: %v\n", mes)
	if mes.IsPathWriter {
		r.updateTrace(mes)
	}
	if mes.Target == r.gnode.name {
		logf.debugMes(mes, "answer reached its target: %v\n", mes.Id)
		if mes.FromClient != "" {
			if client, ok := r.gnode.clientMap[mes.FromClient]; ok {
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
			logf.info("answer originId=%s saved in receiveMap\n", mes.OriginId)
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

func (r *MessageReceiver) convertFunctionArgs(f reflect.Value, name string, args []string) ([]reflect.Value, error) {
	in := make([]reflect.Value, f.Type().NumIn())
	in[0] = reflect.ValueOf(r.gnode)
	if args != nil {
		k := 0
		for _, arg := range args {
			atype := f.Type().In(k + 1)
			val, err := r.unmarshal(arg, atype.String())
			if err != nil {
				return nil, fmt.Errorf("Error umarshaling arg %d on function %s: %v", k, name, err)
			}
			in[k+1] = val
			k++
		}
	}
	return in, nil
}

func (r *MessageReceiver) unmarshal(arg string, atype string) (reflect.Value, error) {
	if atype == "int" {
		val, err := strconv.Atoi(arg)
		if err != nil {
			return reflect.ValueOf(arg), fmt.Errorf("Argument: %s is not an int", arg)
		}
		return reflect.ValueOf(val), nil
	} else if atype == "bool" {
		if strings.ToLower(arg) == "true" {
			return reflect.ValueOf(true), nil
		}
		return reflect.ValueOf(false), nil

	} else if atype == "time.Time" {
		val, err := time.Parse(time.RFC3339Nano, arg)
		if err != nil {
			return reflect.ValueOf(arg), fmt.Errorf("Argument: %s is not a time.Time", arg)
		}
		return reflect.ValueOf(val), nil
	}
	return reflect.ValueOf(arg), nil
}

func (r *MessageReceiver) marshal(val reflect.Value) (string, error) {
	atype := val.Type().String()
	if atype == "int" {
		return fmt.Sprintf("%d", val.Int()), nil
	}
	if atype == "bool" {
		return fmt.Sprintf("%t", val.Bool()), nil
	}
	return val.String(), nil
}

func (r *MessageReceiver) sendback(mes *AntMes, ret []reflect.Value) error {
	args := []string{}
	for _, val := range ret {
		arg, err := r.marshal(val)
		if err != nil {

		}
		args = append(args, arg)
	}
	//return g.sendMessage(mes.Origin, mes.Name, args...)
	return nil
}
