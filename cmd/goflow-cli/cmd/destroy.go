package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/run"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy workerpool and redis containers",
	RunE: func(_ *cobra.Command, _ []string) error {
		conf, err := config.Get()
		if err != nil {
			return err
		}

		logger := log.NewConsoleLogger()

		return run.Destroy(conf, logger)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
