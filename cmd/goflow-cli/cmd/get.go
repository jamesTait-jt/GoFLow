package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/run"
	"github.com/spf13/cobra"
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

		return run.Get(args[0], conf.GoFlowServer.Address)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
