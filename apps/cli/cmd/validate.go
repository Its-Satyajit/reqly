// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/validation"
)

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate OpenAPI specifications or Git-native project descriptors",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}

		info, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("stat %s: %w", target, err)
		}

		var res *validation.Result
		if info.IsDir() {
			res, err = validation.ValidateProject(target)
		} else {
			res, err = validation.ValidateOpenAPIFile(target)
		}

		if err != nil {
			return err
		}

		if !res.Valid {
			fmt.Fprintln(cmd.OutOrStdout(), "Validation failed:")
			for _, e := range res.Errors {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", e)
			}
			return fmt.Errorf("validation failed with %d error(s)", len(res.Errors))
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Validation passed cleanly.")
		return nil
	},
}
