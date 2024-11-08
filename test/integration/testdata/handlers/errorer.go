package main

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/task"
)

func handle(payload any) task.Result {
	err := fmt.Errorf("error for payload: %v", payload)
	return task.Result{ErrMsg: err.Error()}
}

func NewHandler() task.Handler {
	return handle
}
