// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/policy"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Local organization policies (0600, Git-native)",
	Long:  `Show or validate the local policy file at .reqly/policy.yaml (0600).`,
}

var policyShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			path = policy.DefaultPath(".")
		}
		p, err := policy.Load(path)
		if err != nil {
			return err
		}
		data, err := yaml.Marshal(p)
		if err != nil {
			return err
		}
		_, err = cmd.OutOrStdout().Write(data)
		return err
	},
}

var policyValidateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate a policy file",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := policy.DefaultPath(".")
		if len(args) > 0 {
			path = args[0]
		}
		p, err := policy.Load(path)
		if err != nil {
			return err
		}
		if err := policy.Validate(p); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Policy is valid")
		return nil
	},
}

var policyEnforceCmd = &cobra.Command{
	Use:   "enforce --action <action> --resource <resource>",
	Short: "Check if an action is allowed by the policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		action, _ := cmd.Flags().GetString("action")
		resource, _ := cmd.Flags().GetString("resource")
		if action == "" || resource == "" {
			return fmt.Errorf("--action and --resource are required")
		}
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			path = policy.DefaultPath(".")
		}
		p, err := policy.Load(path)
		if err != nil {
			return err
		}
		if err := policy.Enforce(p, action, resource); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Allowed")
		return nil
	},
}

func init() {
	policyShowCmd.Flags().String("file", "", "policy file path (default .reqly/policy.yaml)")
	policyValidateCmd.Flags().String("file", "", "policy file path")
	policyEnforceCmd.Flags().String("action", "", "action to check (e.g. request.send)")
	policyEnforceCmd.Flags().String("resource", "", "resource to check")
	policyEnforceCmd.Flags().String("file", "", "policy file path")
	policyCmd.AddCommand(policyShowCmd)
	policyCmd.AddCommand(policyValidateCmd)
	policyCmd.AddCommand(policyEnforceCmd)
	rootCmd.AddCommand(policyCmd)
}
