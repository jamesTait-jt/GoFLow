package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/run"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
)

var (
	handlersPath string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy workerpool with Redis broker and compiled plugins",
	RunE: func(_ *cobra.Command, _ []string) error {
		conf, err := config.Get()
		if err != nil {
			return err
		}

		logger := log.NewConsoleLogger()

		return run.Deploy(conf, logger)
	},
}

func init() {
	deployCmd.Flags().StringVarP(&handlersPath, "path-to-handlers", "p", "", "The full path to the location of your custom handlers (required)")

	rootCmd.AddCommand(deployCmd)
}
