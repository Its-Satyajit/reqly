// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/diffing"
)

var diffCmd = &cobra.Command{
	Use:   "diff <file1> <file2>",
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

		res, err := diffing.JSON(bytes1, bytes2)
		if err != nil {
			return err
		}

		if !res.HasChanges {
			fmt.Fprintln(cmd.OutOrStdout(), "No structural changes found.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Found %d change(s):\n", len(res.Changes))
		for _, c := range res.Changes {
			fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %v: %v -> %v\n", c.Type, c.Path, c.From, c.To)
		}
		return nil
	},
}
