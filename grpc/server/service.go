package server

import (
	"github.com/jamesTait-jt/goflow"
	"github.com/jamesTait-jt/goflow/task"
)

// GoFlowService provides an object for interacting with a GoFlow instance.
// It exposes methods to push tasks to the GoFlow system and retrieve task results.
//
// The service acts as a wrapper around the GoFlow instance, forwarding requests
// to the underlying GoFlow system for task submission and result retrieval.
type GoFlowService struct {
	gf *goflow.GoFlow
}

// NewGoFlowService creates and returns a new GoFlowService instance that wraps
// the provided GoFlow instance. The service can then be used to submit tasks and
// fetch results from the GoFlow system.
func NewGoFlowService(gf *goflow.GoFlow) *GoFlowService {
	return &GoFlowService{gf: gf}
}

// PushTask pushes a task of the specified type with the provided payload
// to the GoFlow system. It returns the task ID upon success, or an error if the
// task could not be pushed.
func (gf *GoFlowService) PushTask(taskType string, payload any) (string, error) {
	return gf.gf.Push(taskType, payload)
}

// GetResult retrieves the result of a task identified by taskID from the GoFlow system.
// It returns the task result, a boolean indicating if the result was found, and
// an error if any occurred while fetching the result.
func (gf *GoFlowService) GetResult(taskID string) (task.Result, bool, error) {
	return gf.gf.GetResult(taskID)
}
