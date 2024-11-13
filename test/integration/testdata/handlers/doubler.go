package main

import (
	"encoding/json"
	"fmt"

	"github.com/jamesTait-jt/goflow/task"
)

type input struct {
	N int
}

func handle(payload any) task.Result {
	instr, ok := payload.(string)
	if !ok {
		return task.Result{
			ErrMsg: fmt.Sprintf("wrong input type: %t", payload),
		}
	}

	var in input
	err := json.Unmarshal([]byte(instr), &in)
	if err != nil {
		return task.Result{
			ErrMsg: fmt.Sprintf("badly formed input [%v]: %v", instr, err),
		}
	}

	return task.Result{Payload: in.N * 2}
}

func NewHandler() task.Handler {
	return handle
}
