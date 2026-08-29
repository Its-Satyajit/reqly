// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/rbac"
)

var rbacCmd = &cobra.Command{
	Use:   "rbac",
	Short: "Local RBAC (roles and permissions, 0600)",
	Long:  `Show roles or check if a user can perform an action via .reqly/rbac.yaml (0600).`,
}

var rbacListCmd = &cobra.Command{
	Use:   "list",
	Short: "List roles and their permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			path = rbac.DefaultPath(".")
		}
		r, err := rbac.Load(path)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		for _, name := range rbac.ListRoles(r) {
			role := r.Roles[name]
			fmt.Fprintf(out, "%s: %v\n", role.Name, role.Permissions)
		}
		return nil
	},
}

var rbacCheckCmd = &cobra.Command{
	Use:   "check --user <user> --action <action> --resource <resource>",
	Short: "Check if a user can perform an action",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, _ := cmd.Flags().GetString("user")
		action, _ := cmd.Flags().GetString("action")
		resource, _ := cmd.Flags().GetString("resource")
		if user == "" || action == "" || resource == "" {
			return fmt.Errorf("--user, --action and --resource are required")
		}
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			path = rbac.DefaultPath(".")
		}
		r, err := rbac.Load(path)
		if err != nil {
			return err
		}
		if err := rbac.Enforce(r, user, action, resource); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Allowed")
		return nil
	},
}

func init() {
	rbacListCmd.Flags().String("file", "", "rbac file path (default .reqly/rbac.yaml)")
	rbacCheckCmd.Flags().String("user", "", "user to check")
	rbacCheckCmd.Flags().String("action", "", "action to check")
	rbacCheckCmd.Flags().String("resource", "", "resource to check")
	rbacCheckCmd.Flags().String("file", "", "rbac file path")
	rbacCmd.AddCommand(rbacListCmd)
	rbacCmd.AddCommand(rbacCheckCmd)
	rootCmd.AddCommand(rbacCmd)
}
