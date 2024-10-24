package run

import (
	"context"
	"log"
	"time"

	pb "github.com/jamesTait-jt/goflow/cmd/goflow/goflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Push(taskType, payload, serverAddr string) error {
	conn, err := grpc.NewClient(
		serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	goFlowClient := pb.NewGoFlowClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	r, err := goFlowClient.PushTask(ctx, &pb.PushTaskRequest{TaskType: taskType, Payload: payload})
	if err != nil {
		log.Fatalf("could not push: %v", err)
	}
	log.Printf("Task ID: %s", r.GetId())

	return nil
}
