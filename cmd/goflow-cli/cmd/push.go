package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/service"
	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push a task to the workerpool",
	Args: func(cmd *cobra.Command, args []string) error {
		numRequiredArgs := 2

		if err := cobra.ExactArgs(numRequiredArgs)(cmd, args); err != nil {
			return err
		}

		if !json.Valid([]byte(args[1])) {
			return fmt.Errorf("payload must be a string in json format")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.Get()
		if err != nil {
			return err
		}

		serverAddr := fmt.Sprintf("%s:%d", conf.GoFlowServer.Address, grpcserver.GRPCPort)
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		defer conn.Close()

		goFlowClient := pb.NewGoFlowClient(conn)
		logger := log.NewConsoleLogger()

		goFlowService := service.NewGoFlowService(goFlowClient, time.Minute, logger)

		taskID, err := goFlowService.Push(args[0], args[1])
		if err != nil {
			return err
		}

		cmd.Printf("TaskID: '%s'\n", taskID)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
