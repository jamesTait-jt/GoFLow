// Package client provides a gRPC client for interfacing with a GoFlow server.
// It enables pushing tasks to the server and retrieving task results through
// gRPC method calls, encapsulated within GoFlowGRPCClient.
package client

import (
	"context"
	"fmt"

	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GoFlowGRPCClient is a client for interacting with the GoFlow gRPC server.
// It provides methods to push tasks to the server and retrieve results.
// GoFlowGRPCClient manages connection and configuration details through options.
type GoFlowGRPCClient struct {
	opts   goFlowGRPCClientOptions
	client pb.GoFlowClient
}

// NewGoFlowClient creates a new GoFlowGRPCClient connected to the GoFlow gRPC server
// specified by connString. Optional configuration can be provided to customize client settings.
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

// Push submits a task to the GoFlow server. It takes a task type and payload
// as strings, and returns the task's ID if successfully pushed, or an error otherwise.
//
// The Push method uses a context with a timeout defined in the client's options.
// If the request fails, the returned error wraps the underlying gRPC error.
func (g *GoFlowGRPCClient) Push(taskType, payload string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.opts.requestTimeout)
	defer cancel()

	r, err := g.client.PushTask(ctx, &pb.PushTaskRequest{TaskType: taskType, Payload: payload})
	if err != nil {
		return "", fmt.Errorf("failed to push task: %w", err)
	}

	return r.GetId(), nil
}

// Get retrieves the result of a task from the GoFlow server using the provided task ID.
// If successful, Get returns the task result as a string. If the task encountered an error,
// it returns the error message instead. Returns an error if the request fails.
//
// Get uses a context with a timeout from the client's options, and wraps errors
// from the underlying gRPC call.
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
