// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/scim"
)

var scimCmd = &cobra.Command{
	Use:   "scim",
	Short: "SCIM provisioning (local in-memory, zero telemetry)",
	Long: `Manage SCIM users and groups locally (in-memory store for M73).

This is a local, per-invocation store with zero telemetry: users created in
one command do not persist to the next invocation. This is by design. Listing
after creating in a separate process will show "No users".`,
}

var scimUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage SCIM users",
}

var scimUserCreateCmd = &cobra.Command{
	Use:   "create --username <name> --email <email>",
	Short: "Create a SCIM user",
	Long:  `Create a SCIM user in the local in-memory store. Note: store is per-invocation and does not persist across separate reqly invocations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")
		store := scim.NewStore()
		u, err := store.CreateUser(scim.User{UserName: username, Email: email})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created user %q (%s)\n", u.UserName, u.ID)
		fmt.Fprintln(cmd.ErrOrStderr(), "note: SCIM store is in-memory per invocation (M73); users do not persist to next command")
		return nil
	},
}

var scimUserListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SCIM users",
	Long:  `List SCIM users from the local in-memory store. Note: store is per-invocation (M73) and does not persist across separate reqly invocations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store := scim.NewStore()
		users := store.ListUsers()
		out := cmd.OutOrStdout()
		if len(users) == 0 {
			fmt.Fprintln(out, "No users (in-memory store is per-invocation; users from prior commands are not persisted)")
			return nil
		}
		for _, u := range users {
			fmt.Fprintf(out, "%s\t%s\t%v\n", u.ID, u.UserName, u.Active)
		}
		return nil
	},
}

func init() {
	scimUserCreateCmd.Flags().String("username", "", "userName")
	scimUserCreateCmd.Flags().String("email", "", "email")
	_ = scimUserCreateCmd.MarkFlagRequired("username")
	scimUserCmd.AddCommand(scimUserCreateCmd)
	scimUserCmd.AddCommand(scimUserListCmd)
	scimCmd.AddCommand(scimUserCmd)
	rootCmd.AddCommand(scimCmd)
}
