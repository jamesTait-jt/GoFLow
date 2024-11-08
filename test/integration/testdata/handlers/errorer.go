package main

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/task"
)

func handle(payload any) task.Result {
	err := fmt.Errorf("error for payload: %v", payload)
	return task.Result{Error: err}
}

func NewHandler() task.Handler {
	return handle
}
