package client

import (
	"context"
	"fmt"

	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GoFlowGRPCClient struct {
	opts   goFlowGRPCClientOptions
	client pb.GoFlowClient
}

func NewGoFlowClient(connString string, opt ...GoFlowGRPCClientOption) (*GoFlowGRPCClient, error) {
	opts := defaultServerOptions

	for _, o := range opt {
		o.apply(&opts)
	}

	conn, err := grpc.NewClient(connString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GoFlowGRPCClient{
		opts:   opts,
		client: pb.NewGoFlowClient(conn),
	}, nil
}

func (g *GoFlowGRPCClient) Push(taskType, payload string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.opts.requestTimeout)
	defer cancel()

	r, err := g.client.PushTask(ctx, &pb.PushTaskRequest{TaskType: taskType, Payload: payload})
	if err != nil {
		return "", fmt.Errorf("failed to push task: %w", err)
	}

	return r.GetId(), nil
}

func (g *GoFlowGRPCClient) Get(taskID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.opts.requestTimeout)
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
