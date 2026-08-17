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
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// tokenStoreFor resolves the token store for the workspace nearest to startDir.
// Returns nil when no workspace descriptor exists up the tree.
func tokenStoreFor(startDir string) (*secrets.FileStore, error) {
	root := findWorkspaceRoot(startDir)
	if root == "" {
		return nil, nil
	}
	return secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
}

// authCmd reports and manages locally cached OAuth tokens. Acquisition is
// automatic on first request; there is no interactive login.
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Inspect locally cached OAuth tokens",
	Long: `Inspect and clear locally cached OAuth 2.0 tokens.

Tokens are acquired automatically when a request uses an OAuth2 auth scheme and
cached per workspace in <workspace>/.reqly/tokens.json (0600). There is no
interactive login: the next request re-acquires a token whenever the cached one
is missing or expired.`,
}

// authStatusCmd reports cached tokens for the nearest workspace.
var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cached OAuth tokens for this workspace",
	Long: `List cached OAuth 2.0 tokens for the workspace nearest to the current
directory, one per line: token endpoint, expiry, a masked token, and whether
the token is still valid. Token values never print in full.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := tokenStoreFor(".")
		if err != nil {
			return err
		}
		if store == nil {
			fmt.Fprintln(cmd.OutOrStdout(), "no workspace found: no cached tokens")
			return nil
		}

		keys, err := store.Keys()
		if err != nil {
			return fmt.Errorf("read token store: %w", err)
		}
		if len(keys) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "no cached tokens")
			return nil
		}

		sort.Strings(keys)
		now := time.Now()
		for _, key := range keys {
			raw, err := store.Get(key)
			if err != nil {
				return fmt.Errorf("read token %s: %w", key, err)
			}
			tok, err := auth.ParseCachedToken(raw)
			if err != nil {
				return fmt.Errorf("decode token %s: %w", key, err)
			}

			state := "cached"
			if !tok.Expiry.IsZero() && now.After(tok.Expiry) {
				state = "expired"
			}

			endpoint := tok.Endpoint
			if endpoint == "" {
				endpoint = "(unknown)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n",
				endpoint, formatExpiry(tok.Expiry), maskToken(tok.AccessToken), state)
		}
		return nil
	},
}

// authLogoutCmd clears cached tokens for the nearest workspace.
var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear cached OAuth tokens for this workspace",
	Long: `Remove all cached OAuth 2.0 tokens for the workspace nearest to the
current directory. The next request that needs a token re-acquires one from the
token endpoint.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := tokenStoreFor(".")
		if err != nil {
			return err
		}
		if store == nil {
			fmt.Fprintln(cmd.OutOrStdout(), "no workspace found: nothing to clear")
			return nil
		}

		keys, err := store.Keys()
		if err != nil {
			return fmt.Errorf("read token store: %w", err)
		}
		for _, key := range keys {
			if err := store.Delete(key); err != nil {
				return fmt.Errorf("clear token %s: %w", key, err)
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "cleared %d cached token(s)\n", len(keys))
		return nil
	},
}

// formatExpiry renders a token expiry, or "-" when unknown.
func formatExpiry(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.RFC3339)
}

// maskToken renders a token with only its first and last four characters
// visible, so status output never prints a full credential.
func maskToken(token string) string {
	if token == "" {
		return "(empty)"
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}

func init() {
	authCmd.AddCommand(authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
