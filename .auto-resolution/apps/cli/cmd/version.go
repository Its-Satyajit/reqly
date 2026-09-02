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

var (
	versionVerbose bool
	versionCommit  bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Reqly version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if versionCommit {
			fmt.Fprintln(cmd.OutOrStdout(), version.Commit)
			return nil
		}
		if versionVerbose {
			fmt.Fprintf(cmd.OutOrStdout(), "version: %s\ncommit: %s\n", version.Version, version.Commit)
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), version.Version)
		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionVerbose, "verbose", false, "print version and commit")
	versionCmd.Flags().BoolVar(&versionCommit, "commit", false, "print commit hash only")
}
