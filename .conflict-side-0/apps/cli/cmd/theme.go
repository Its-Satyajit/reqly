// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/theme"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage shareable UI themes (import/export/list)",
	Long:  `Git-native theme sharing: list built-ins, export a theme, import a custom theme file.`,
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List built-in and available themes",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		for _, t := range theme.BuiltInThemes() {
			fmt.Fprintf(out, "%s\t%s\t%s\n", t.ID, t.Label, t.Appearance)
		}
		return nil
	},
}

var themeExportCmd = &cobra.Command{
	Use:   "export <theme-id>",
	Short: "Export a theme as YAML",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		outPath, _ := cmd.Flags().GetString("out")
		var found *theme.Theme
		for _, t := range theme.BuiltInThemes() {
			if t.ID == id {
				tt := t
				found = &tt
				break
			}
		}
		if found == nil {
			return fmt.Errorf("theme %q not found", id)
		}
		data, err := yaml.Marshal(found)
		if err != nil {
			return err
		}
		if outPath != "" {
			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				return fmt.Errorf("write: %w", err)
			}
			return nil
		}
		_, err = cmd.OutOrStdout().Write(data)
		return err
	},
}

var themeImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import and validate a custom theme file (YAML or JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		th, err := theme.Parse(data)
		if err != nil {
			return err
		}
		css, err := theme.ToCSS(th)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Imported theme %q (%s) — CSS: %s\n", th.ID, th.Label, css)
		return nil
	},
}

func init() {
	themeExportCmd.Flags().String("out", "", "write to file instead of stdout")
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeExportCmd)
	themeCmd.AddCommand(themeImportCmd)
	rootCmd.AddCommand(themeCmd)
}
