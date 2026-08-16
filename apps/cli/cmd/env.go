// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/environments"
)

// envCmd manages Git-native environments stored as environments/<name>.yaml.
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environments and their variables",
	Long: `Manage Git-native environments stored as environments/<name>.yaml.

Each environment file holds variables: and secrets: maps. Secret values render
as [SECRET] in output and never print. The active environment is selected by
REQLY_ENV, the --env flag, a file's environment: field, or the workspace
descriptor's environment: field (highest wins).`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available environments",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		names, err := environments.List(".")
		if err != nil {
			return err
		}
		for _, name := range names {
			fmt.Fprintln(cmd.OutOrStdout(), name)
		}
		return nil
	},
}

var envShowCmd = &cobra.Command{
	Use:   "show [<name>]",
	Short: "Show an environment's variables with secrets masked",
	Long: `Print an environment's variables and secrets. Secrets render as
[SECRET]. Defaults to the active environment when <name> is omitted.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			name = resolveActiveName()
		}
		if name == "" {
			return errNoEnvironment
		}

		env, err := environments.Read(name, ".")
		if err != nil {
			return err
		}
		printEnv(cmd, env)
		return nil
	},
}

var envUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the workspace's active environment",
	Long: `Persist <name> as the active environment in the workspace descriptor
(reqly.yaml). Errors when the environment file is missing or when no workspace
descriptor exists.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if _, err := environments.Read(name, "."); err != nil {
			return err
		}
		if _, err := collections.LoadWorkspace("."); err != nil {
			return fmt.Errorf("no workspace descriptor found: %w", err)
		}
		if err := collections.SetWorkspaceEnvironment(".", name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "active environment set to %q\n", name)
		return nil
	},
}

var envValidateCmd = &cobra.Command{
	Use:   "validate [<name>]",
	Short: "Validate an environment file and its variable usage",
	Long: `Check an environment for common problems: variables whose names look
like secrets but are stored unencrypted in the variables map, keys defined in
both variables and secrets, and {{var}} references in workspace requests that
the environment does not provide. Defaults to the active environment.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			name = resolveActiveName()
		}
		if name == "" {
			return errNoEnvironment
		}

		env, err := environments.Read(name, ".")
		if err != nil {
			return err
		}
		issues, err := environments.Validate(env, workingDir())
		if err != nil {
			return err
		}
		if len(issues) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "environment %q is valid\n", name)
			return nil
		}
		for _, issue := range issues {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", issue.Severity, issue.Message)
		}
		return fmt.Errorf("environment %q has %d issue(s)", name, len(issues))
	},
}

var envDiffCmd = &cobra.Command{
	Use:   "diff <name-a> <name-b>",
	Short: "Compare two environments key by key",
	Long: `Compare two environments. Keys added, removed, or changed are printed
with their values; secret values always render as [SECRET].`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := environments.Read(args[0], ".")
		if err != nil {
			return err
		}
		b, err := environments.Read(args[1], ".")
		if err != nil {
			return err
		}
		diffs := environments.Diff(a, b)
		if len(diffs) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No structural changes found.")
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d difference(s) between %q and %q:\n", len(diffs), args[0], args[1])
		for _, d := range diffs {
			switch d.Status {
			case environments.StatusAdded:
				fmt.Fprintf(cmd.OutOrStdout(), "  + %s = %s\n", d.Name, d.To)
			case environments.StatusRemoved:
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s = %s\n", d.Name, d.From)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "  ~ %s: %s -> %s\n", d.Name, d.From, d.To)
			}
		}
		return nil
	},
}

// printEnv writes an environment's description, variables, and masked secrets
// to the command output. A key present in both variables and secrets prints
// once, as a secret, so `env show` never leaks the plain value.
func printEnv(cmd *cobra.Command, env *environments.Environment) {
	if env.Description != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "# %s\n", env.Description)
	}
	if len(env.Variables) == 0 && len(env.Secrets) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "(no variables)")
	}
	keys := make(map[string]bool, len(env.Variables)+len(env.Secrets))
	for key := range env.Variables {
		keys[key] = false
	}
	for key := range env.Secrets {
		keys[key] = true
	}
	names := make([]string, 0, len(keys))
	for key := range keys {
		names = append(names, key)
	}
	sort.Strings(names)
	for _, key := range names {
		if keys[key] {
			fmt.Fprintf(cmd.OutOrStdout(), "%s = %s\n", key, environments.MaskedSecret)
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s = %s\n", key, env.Variables[key])
	}
}

// workingDir is the current directory used as the scan root for env validate.
func workingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return filepath.Clean(wd)
}

// errNoEnvironment is returned when show/validate need an environment name but
// none is selected and none was passed.
var errNoEnvironment = fmt.Errorf("no environment selected; pass a name or set REQLY_ENV/--env")

// resolveActiveName returns the environment selected with the highest
// precedence from REQLY_ENV, the --env flag, or the workspace descriptor's
// environment: field, or "" when none is set.
func resolveActiveName() string {
	sel := environments.Selection{
		EnvFlag:   envSelection(os.Getenv("REQLY_ENV"), envFlag),
		ConfigEnv: collections.WorkspaceEnvironment("."),
	}
	return sel.Active()
}

func init() {
	envCmd.AddCommand(envListCmd, envShowCmd, envUseCmd, envValidateCmd, envDiffCmd)
	rootCmd.AddCommand(envCmd)
}
