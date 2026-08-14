package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff API definitions, requests, or responses",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(core): dispatch to the API diff engine.
		fmt.Fprintln(cmd.OutOrStdout(), "diff: not implemented yet")
		return nil
	},
}
