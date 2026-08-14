package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Run the local mock server",
	Long:  "Start a local mock server generated from the project's OpenAPI definitions or mocks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(core): dispatch to the shared mock server (internal/mocking).
		fmt.Fprintln(cmd.OutOrStdout(), "mock: not implemented yet")
		return nil
	},
}
