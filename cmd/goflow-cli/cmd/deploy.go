package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/run"
	"github.com/spf13/cobra"
)

var (
	handlersPath string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy workerpool with Redis broker and compiled plugins",
	RunE: func(_ *cobra.Command, _ []string) error {
		return run.Deploy(handlersPath)
	},
}

func init() {
	deployCmd.Flags().StringVarP(&handlersPath, "path-to-handlers", "p", "", "The full path to the location of your custom handlers (required)")
	_ = deployCmd.MarkFlagRequired("path-to-handlers")

	rootCmd.AddCommand(deployCmd)
}
