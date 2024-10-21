package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/run"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [taskID]",
	Short: "get the result of a task execution",
	Args:  cobra.MatchAll(cobra.ExactArgs(1)),
	RunE: func(_ *cobra.Command, args []string) error {
		return run.Get(args[0])
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
