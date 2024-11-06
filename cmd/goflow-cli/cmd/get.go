package cmd

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/service"
	pb "github.com/jamesTait-jt/goflow/cmd/goflow/goflow"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var getCmd = &cobra.Command{
	Use:   "get [taskID]",
	Short: "get the result of a task execution",
	Args:  cobra.MatchAll(cobra.ExactArgs(1)),
	RunE: func(_ *cobra.Command, args []string) error {
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
		goFlowService := service.NewGoFlowService(goFlowClient, 30*time.Second, logger)

		taskResult, err := goFlowService.Get(args[0])
		if err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("Task result: '%s'", taskResult))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
