package main

import (
	"github.com/jamesTait-jt/goflow/task"
)

func handle(payload any) task.Result {
	n := payload.(int)

	return task.Result{Payload: n * 2}
}

func NewHandler() task.Handler {
	return handle
}
