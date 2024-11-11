package cmd

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/grpc/client"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [taskID]",
	Short: "get the result of a task execution",
	Args:  cobra.MatchAll(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.Get()
		if err != nil {
			return err
		}

		logger := log.NewConsoleLogger()

		serverAddr := fmt.Sprintf("%s:%d", conf.GoFlowServer.Address, grpcserver.GRPCPort)
		goFlowService, err := client.NewGoFlowService(serverAddr, time.Minute, logger)
		if err != nil {
			return err
		}

		taskResult, err := goFlowService.Get(args[0])
		if err != nil {
			return err
		}

		cmd.Printf("Task result: '%s'\n", taskResult)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
