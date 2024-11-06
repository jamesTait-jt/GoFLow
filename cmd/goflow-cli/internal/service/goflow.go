package service

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jamesTait-jt/goflow/cmd/goflow/goflow"
	"github.com/jamesTait-jt/goflow/pkg/log"
)

type GoFlowService struct {
	client         pb.GoFlowClient
	requestTimeout time.Duration
	logger         log.Logger
}

func NewGoFlowService(client pb.GoFlowClient, timeout time.Duration, logger log.Logger) *GoFlowService {
	return &GoFlowService{
		client:         client,
		requestTimeout: timeout,
		logger:         logger,
	}
}

func (g *GoFlowService) Push(taskType, payload string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.requestTimeout)
	defer cancel()

	r, err := g.client.PushTask(ctx, &pb.PushTaskRequest{TaskType: taskType, Payload: payload})
	if err != nil {
		return "", fmt.Errorf("failed to push task: %w", err)
	}

	return r.GetId(), nil
}

func (g *GoFlowService) Get(taskID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.requestTimeout)
	defer cancel()

	r, err := g.client.GetResult(ctx, &pb.GetResultRequest{TaskID: taskID})
	if err != nil {
		return "", fmt.Errorf("could not get result for taskID '%s': %w", taskID, err)
	}

	return r.GetResult(), nil
}
