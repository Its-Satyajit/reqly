package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run project tests",
	Long:  "Execute the test suites defined in the current project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(core): dispatch to the shared testing engine (internal/testing).
		fmt.Fprintln(cmd.OutOrStdout(), "test: not implemented yet")
		return nil
	},
}
