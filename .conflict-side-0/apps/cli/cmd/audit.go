// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/audit"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Local audit trail (append-only, 0600)",
	Long:  `List or clear the local audit log at .reqly/audit.log (JSONL, 0600).`,
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List audit entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := audit.NewStore(".")
		if err != nil {
			return err
		}
		entries, err := store.List()
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if len(entries) == 0 {
			fmt.Fprintln(out, "No audit entries")
			return nil
		}
		for _, e := range entries {
			fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\n", e.Timestamp.Format("2006-01-02 15:04:05"), e.Actor, e.Action, e.Resource, e.Details)
		}
		return nil
	},
}

var auditClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the audit log",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := audit.NewStore(".")
		if err != nil {
			return err
		}
		if err := store.Clear(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Audit log cleared")
		return nil
	},
}

func init() {
	auditCmd.AddCommand(auditListCmd)
	auditCmd.AddCommand(auditClearCmd)
	rootCmd.AddCommand(auditCmd)
}
