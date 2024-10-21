package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/run"
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
	RunE: func(_ *cobra.Command, args []string) error {
		return run.Push(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
