package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <request>",
	Short: "Execute a single request",
	Long:  "Execute a single request by name or ID from the current project.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(core): dispatch to the shared request engine (internal/request).
		fmt.Fprintf(cmd.OutOrStdout(), "run %s: not implemented yet\n", args[0])
		return nil
	},
}
