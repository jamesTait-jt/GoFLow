package main

import (
	"github.com/jamesTait-jt/goflow/task"
)

func handle(payload any) task.Result {
	return task.Result{Payload: payload}
}

func NewHandler() task.Handler {
	return handle
}
