package cmd

import (
	"os"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/spf13/cobra"
)

var confPath string

var rootCmd = &cobra.Command{
	Use:   "goflow",
	Short: "Goflow CLI tool to deploy workerpool and plugins using Docker",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		err := config.Load(confPath)
		if err != nil {
			return err
		}

		return nil
	},
}

// Execute runs the root command for the Goflow CLI application.
// It initializes and parses any command-line flags or arguments,
// and attempts to execute the specified command. If an error occurs
// during execution, Execute exits the application with a non-zero status code.
//
// This function should be called from the main package to start the CLI tool.
// It utilizes the Cobra library for command-line parsing, and loads
// configuration settings from a specified file path before executing commands.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&confPath, "config", "./.goflow.yaml", "config file (default is ./.goflow.yaml)")
}
