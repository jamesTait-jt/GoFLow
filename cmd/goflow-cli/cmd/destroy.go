package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/run"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy workerpool and redis containers",
	RunE: func(_ *cobra.Command, _ []string) error {
		return run.Destroy()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
