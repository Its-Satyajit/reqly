// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/collab"
)

var collabCmd = &cobra.Command{
	Use:   "collab",
	Short: "Shared workspaces (Git-native, local)",
	Long:  `Manage collaborators for a shared workspace via .reqly/collab.yaml (0600).`,
}

var collabListCmd = &cobra.Command{
	Use:   "list",
	Short: "List collaborators",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			path = collab.DefaultPath(".")
		}
		ws, err := collab.Load(path)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if len(ws.Collaborators) == 0 {
			fmt.Fprintln(out, "No collaborators")
			return nil
		}
		for _, c := range ws.Collaborators {
			fmt.Fprintf(out, "%s\t%s\t%s\n", c.User, c.Role, c.AddedAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	},
}

var collabAddCmd = &cobra.Command{
	Use:   "add --user <user> --role <viewer|editor|admin>",
	Short: "Add a collaborator",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, _ := cmd.Flags().GetString("user")
		role, _ := cmd.Flags().GetString("role")
		if user == "" || role == "" {
			return fmt.Errorf("--user and --role are required")
		}
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			path = collab.DefaultPath(".")
		}
		ws, err := collab.Load(path)
		if err != nil {
			return err
		}
		if ws.Path == "" {
			ws.Path = "."
		}
		if err := collab.AddCollaborator(&ws, user, role); err != nil {
			return err
		}
		if err := collab.Save(path, ws); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Added %q as %s\n", user, role)
		return nil
	},
}

var collabRemoveCmd = &cobra.Command{
	Use:   "remove --user <user>",
	Short: "Remove a collaborator",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, _ := cmd.Flags().GetString("user")
		if user == "" {
			return fmt.Errorf("--user is required")
		}
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			path = collab.DefaultPath(".")
		}
		ws, err := collab.Load(path)
		if err != nil {
			return err
		}
		if err := collab.RemoveCollaborator(&ws, user); err != nil {
			return err
		}
		if err := collab.Save(path, ws); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %q\n", user)
		return nil
	},
}

func init() {
	collabListCmd.Flags().String("file", "", "collab file path (default .reqly/collab.yaml)")
	collabAddCmd.Flags().String("user", "", "user to add")
	collabAddCmd.Flags().String("role", "", "role (viewer|editor|admin)")
	collabAddCmd.Flags().String("file", "", "collab file path")
	collabRemoveCmd.Flags().String("user", "", "user to remove")
	collabRemoveCmd.Flags().String("file", "", "collab file path")
	collabCmd.AddCommand(collabListCmd)
	collabCmd.AddCommand(collabAddCmd)
	collabCmd.AddCommand(collabRemoveCmd)
	rootCmd.AddCommand(collabCmd)
}
