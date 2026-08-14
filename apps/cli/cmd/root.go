package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "api-client",
	Short: "A local-first API development environment",
	Long: `A local-first, Git-native API development environment.

Requests, tests, schemas, mocks, environments, and documentation live together
as version-controlled project files. The CLI shares the same Go core as the
desktop application.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(
		runCmd,
		testCmd,
		collectionCmd,
		mockCmd,
		validateCmd,
		diffCmd,
		docsCmd,
	)
}
