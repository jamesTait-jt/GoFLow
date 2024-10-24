package cmd

import (
	"os"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/spf13/cobra"
)

var confPath string

var rootCmd = &cobra.Command{
	Use:   "goflow",
	Short: "Goflow CLI tool to deploy workerpool and plugins using Docker",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := config.Load(confPath)
		if err != nil {
			return err
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&confPath, "config", "./.goflow.yaml", "config file (default is ./.goflow.yaml)")
}
