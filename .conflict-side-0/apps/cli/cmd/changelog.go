// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/diffing"
)

var (
	changelogFormat         string
	changelogFailOnBreaking bool
)

var changelogCmd = &cobra.Command{
	Use:   "changelog <old-spec> <new-spec> [--format markdown|json] [--fail-on-breaking]",
	Short: "Generate human-readable API changelog and suggested SemVer bump",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldBytes, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read %s: %w", args[0], err)
		}
		newBytes, err := os.ReadFile(args[1])
		if err != nil {
			return fmt.Errorf("read %s: %w", args[1], err)
		}

		cl, err := diffing.GenerateChangelog(oldBytes, newBytes)
		if err != nil {
			return fmt.Errorf("generate changelog: %w", err)
		}

		if changelogFormat == "json" {
			out, err := cl.ToJSON()
			if err != nil {
				return fmt.Errorf("json format: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
		} else {
			fmt.Fprint(cmd.OutOrStdout(), cl.ToMarkdown())
		}

		if changelogFailOnBreaking && len(cl.Breaking) > 0 {
			return fmt.Errorf("detected %d breaking change(s)", len(cl.Breaking))
		}

		return nil
	},
}

func init() {
	changelogCmd.Flags().StringVar(&changelogFormat, "format", "markdown", "output format (markdown, json)")
	changelogCmd.Flags().BoolVar(&changelogFailOnBreaking, "fail-on-breaking", false, "exit with error if any breaking changes are detected")
}
