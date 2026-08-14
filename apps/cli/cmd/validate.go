package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the project and its specifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(core): dispatch to OpenAPI / schema validation.
		fmt.Fprintln(cmd.OutOrStdout(), "validate: not implemented yet")
		return nil
	},
}
