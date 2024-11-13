package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/grpc/client"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
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

		logger := log.NewConsoleLogger()

		serverAddr := fmt.Sprintf("%s:%d", conf.GoFlowServer.Address, grpcserver.GRPCPort)
		goFlowService, err := client.NewGoFlowClient(
			serverAddr,
			client.WithRequestTimeout(time.Minute),
			client.WithLogger(logger),
		)
		if err != nil {
			return err
		}

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
