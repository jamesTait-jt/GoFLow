package server

import (
	"github.com/jamesTait-jt/goflow"
	"github.com/jamesTait-jt/goflow/task"
)

type GoFlowService struct {
	gf *goflow.GoFlow
}

func NewGoFlowService(gf *goflow.GoFlow) *GoFlowService {
	return &GoFlowService{gf: gf}
}

func (gf *GoFlowService) PushTask(taskType string, payload any) (string, error) {
	return gf.gf.Push(taskType, payload)
}

func (gf *GoFlowService) GetResult(taskID string) (task.Result, bool) {
	return gf.gf.GetResult(taskID)
}
