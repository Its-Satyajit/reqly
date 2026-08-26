// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Reqly version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), version.Version)
		return nil
	},
}
