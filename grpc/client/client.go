package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: Add functional options
type GoFlowService struct {
	client         pb.GoFlowClient
	requestTimeout time.Duration
	logger         log.Logger
}

func NewGoFlowService(connString string, timeout time.Duration, logger log.Logger) (*GoFlowService, error) {
	conn, err := grpc.NewClient(connString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GoFlowService{
		client:         pb.NewGoFlowClient(conn),
		requestTimeout: timeout,
		logger:         logger,
	}, nil
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

	switch r.GetErrMsg() {
	case "":
		return r.GetResult(), nil
	default:
		return r.GetErrMsg(), nil
	}
}
