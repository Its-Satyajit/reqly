package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/plugin"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin management",
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		root := findWorkspaceRoot(".")
		if root == "" {
			root = "."
		}
		dir := filepath.Join(root, "plugins")
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(cmd.OutOrStdout(), "no plugins")
				return nil
			}
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				p, err := plugin.Load(filepath.Join(dir, e.Name()))
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "%s: invalid (%v)\n", e.Name(), err)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%v)\n", p.Manifest.Name, p.Manifest.Version, p.Manifest.Capabilities)
			}
		}
		return nil
	},
}

var pluginValidateCmd = &cobra.Command{
	Use:   "validate <name>",
	Short: "Validate a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		root := findWorkspaceRoot(".")
		if root == "" {
			root = "."
		}
		dir := filepath.Join(root, "plugins", name)
		if _, err := plugin.Load(dir); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s: valid\n", name)
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(pluginListCmd, pluginValidateCmd)
}
