package gnode

import (
	"fmt"
)

// Function dedicated for client usage
func CreateMessage(target string, returnAnswer bool, functionName string, args ...string) *AntMes {
	return CreateMessageWithArgName(target, returnAnswer, functionName, args)
}

// Function dedicated for client usage
func CreateMessageWithArgName(target string, returnAnswer bool, functionName string, args []string) *AntMes {
	return &AntMes{
		Target:       target,
		IsAnswer:     false,
		ReturnAnswer: returnAnswer,
		Path:         []string{},
		PathIndex:    -1,
		Function:     functionName,
		Args:         args,
	}
}

func (a *AntMes) toString() string {
	return fmt.Sprintf("mes=%s:%s->%s:%s", a.Id, a.Origin, a.Target, a.Function)
}
