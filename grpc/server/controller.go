package server

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"github.com/jamesTait-jt/goflow/task"

	"github.com/jamesTait-jt/goflow/pkg/log"
)

type goFlowService interface {
	PushTask(taskType string, payload any) (string, error)
	GetResult(taskID string) (task.Result, bool, error)
}

// GoFlowServiceController implements the GoFlow gRPC service.
// It provides the methods to handle PushTask and GetResult requests
// from clients. The controller interacts with the provided `goFlowService`
// to process tasks and results.
type GoFlowServiceController struct {
	svc    goFlowService
	logger log.Logger
	pb.UnimplementedGoFlowServer
}

// NewGoFlowServiceController creates a new GoFlowServiceController with
// the provided service (`goFlowService`) and logger. It returns a
// pointer to the newly created controller.
func NewGoFlowServiceController(svc goFlowService, logger log.Logger) *GoFlowServiceController {
	return &GoFlowServiceController{svc: svc, logger: logger}
}

// PushTask handles the gRPC PushTask request. It processes the incoming task
// request by calling the PushTask method on the service and returns the task's
// ID in the response. If there is an error while pushing the task, the method
// returns the error to the caller.
func (c *GoFlowServiceController) PushTask(_ context.Context, in *pb.PushTaskRequest) (*pb.PushTaskReply, error) {
	c.logger.Info(fmt.Sprintf("Received push task: [%s] [%s]", in.GetTaskType(), in.GetPayload()))

	id, err := c.svc.PushTask(in.GetTaskType(), in.GetPayload())
	if err != nil {
		return nil, err
	}

	return &pb.PushTaskReply{Id: id}, nil
}

// GetResult handles the gRPC GetResult request. It retrieves the result for
// a given task by calling the GetResult method on the service. The result is
// returned in the response. If the task is not complete or doesn't exist,
// it returns an error message.
//
// If the result contains a non-nil payload, it is marshalled to a string
// for inclusion in the response. Otherwise, the method includes the error message.
func (c *GoFlowServiceController) GetResult(_ context.Context, in *pb.GetResultRequest) (*pb.GetResultReply, error) {
	c.logger.Info(fmt.Sprintf("Received get result: [%s]", in.GetTaskID()))

	result, ok, err := c.svc.GetResult(in.GetTaskID())
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("task not complete or didnt exist")
	}

	if result.Payload == nil {
		return &pb.GetResultReply{
			ErrMsg: result.ErrMsg,
		}, nil
	}

	var parsedPayload string

	switch p := result.Payload.(type) {
	case string:
		parsedPayload = p
	default:
		marshalledPayload, err := json.Marshal(p)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result payload: %v", result)
		}

		parsedPayload = string(marshalledPayload)
	}

	return &pb.GetResultReply{
		Result: parsedPayload,
		ErrMsg: result.ErrMsg,
	}, nil
}
