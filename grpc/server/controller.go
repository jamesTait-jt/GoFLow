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

type GoFlowServiceController struct {
	svc    goFlowService
	logger log.Logger
	pb.UnimplementedGoFlowServer
}

func NewGoFlowServiceController(svc goFlowService, logger log.Logger) *GoFlowServiceController {
	return &GoFlowServiceController{svc: svc, logger: logger}
}

func (c *GoFlowServiceController) PushTask(_ context.Context, in *pb.PushTaskRequest) (*pb.PushTaskReply, error) {
	c.logger.Info(fmt.Sprintf("Received push task: [%s] [%s]", in.GetTaskType(), in.GetPayload()))

	id, err := c.svc.PushTask(in.GetTaskType(), in.GetPayload())
	if err != nil {
		return nil, err
	}

	return &pb.PushTaskReply{Id: id}, nil
}

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
