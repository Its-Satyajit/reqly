// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/diffing"
	"github.com/Its-Satyajit/reqly/internal/openapi"
)

var diffFailOnBreaking bool

var diffCmd = &cobra.Command{
	Use:   "diff <file1> <file2> [--fail-on-breaking]",
	Short: "Diff API definitions, requests, or responses",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		bytes1, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read %s: %w", args[0], err)
		}
		bytes2, err := os.ReadFile(args[1])
		if err != nil {
			return fmt.Errorf("read %s: %w", args[1], err)
		}

		doc1, err1 := openapi.Load(bytes1)
		doc2, err2 := openapi.Load(bytes2)

		var res *diffing.DiffResult
		if err1 == nil && err2 == nil {
			// OpenAPI semantic diffing
			rawRes, err := diffing.OpenAPI(doc1, doc2)
			if err != nil {
				return err
			}
			res = diffing.WithSeverity(rawRes)
		} else {
			// Generic JSON/YAML structural diffing
			res, err = diffing.JSON(bytes1, bytes2)
			if err != nil {
				return err
			}
		}

		if !res.HasChanges {
			fmt.Fprintln(cmd.OutOrStdout(), "No structural changes found.")
			return nil
		}

		breakingCount := 0
		fmt.Fprintf(cmd.OutOrStdout(), "Found %d change(s):\n", len(res.Changes))
		for _, c := range res.Changes {
			severityLabel := ""
			if c.Severity != "" {
				severityLabel = fmt.Sprintf(" [%s]", c.Severity)
				if c.Severity == diffing.SeverityBreaking {
					breakingCount++
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  [%s]%s %v: %v -> %v\n", c.Type, severityLabel, strings.Join(c.Path, "."), c.From, c.To)
		}

		if diffFailOnBreaking && breakingCount > 0 {
			return fmt.Errorf("detected %d breaking change(s)", breakingCount)
		}

		return nil
	},
}

func init() {
	diffCmd.Flags().BoolVar(&diffFailOnBreaking, "fail-on-breaking", false, "exit with error if any breaking changes are detected")
}
