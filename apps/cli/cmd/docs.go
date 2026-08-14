package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate API documentation",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(core): dispatch to the documentation generator.
		fmt.Fprintln(cmd.OutOrStdout(), "docs: not implemented yet")
		return nil
	},
}
