package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/service"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy workerpool with Redis broker and compiled plugins",
	RunE: func(_ *cobra.Command, _ []string) error {
		return Deploy()
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func Deploy() error {
	conf, err := config.Get()
	if err != nil {
		return err
	}

	logger := log.NewConsoleLogger()

	deploymentService := service.NewDeploymentService(conf, logger)

	return deploymentService.Deploy()
}
