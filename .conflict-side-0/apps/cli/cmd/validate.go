// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

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
	Long: `Validate an OpenAPI spec or a Git-native workspace.

  reqly validate [path]              # auto-detect: file → OpenAPI, dir → project
  reqly validate openapi <path>      # alias for OpenAPI file
  reqly validate project [path]      # alias for workspace dir
  reqly openapi validate <spec>      # same as 'validate openapi'`,
	Args: cobra.MaximumNArgs(1),
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

var validateOpenAPICmd = &cobra.Command{
	Use:   "openapi <path>",
	Short: "Validate an OpenAPI specification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validation.ValidateOpenAPIFile(args[0])
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

var validateProjectCmd = &cobra.Command{
	Use:   "project [path]",
	Short: "Validate a Git-native workspace",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}
		res, err := validation.ValidateProject(target)
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

func init() {
	validateCmd.AddCommand(validateOpenAPICmd, validateProjectCmd)
}
