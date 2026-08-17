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
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
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

// launchBrowser opens url in the system default browser. It is a package
// variable so tests can substitute a fake driver.
var launchBrowser = func(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}

// authCmd reports and manages locally cached OAuth tokens. Client
// Credentials acquisition is automatic on first request; Authorization Code
// tokens are acquired interactively with `auth login`.
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Inspect and manage locally cached OAuth tokens",
	Long: `Inspect and manage locally cached OAuth 2.0 tokens.

Client Credentials tokens are acquired automatically when a request uses an
OAuth2 auth scheme and are cached per workspace in
<workspace>/.reqly/tokens.json (0600). Authorization Code tokens are acquired
interactively: run "reqly auth login <config>" once to authorize in the
system browser, then requests reuse the cached token (and its refresh token)
until it expires or you run "reqly auth logout".`,
}

// authStatusCmd reports cached tokens for the nearest workspace.
var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cached OAuth tokens for this workspace",
	Long: `List cached OAuth 2.0 tokens for the workspace nearest to the current
directory, one per line: token endpoint, grant type, expiry, masked token,
whether a refresh token is cached, and whether the token is still valid.
Token values never print in full.`,
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
			grant := tok.GrantType
			if grant == "" {
				grant = "-"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\t%s\t%s\n",
				endpoint, grant, formatExpiry(tok.Expiry), maskToken(tok.AccessToken),
				yesNo(tok.RefreshToken != ""), state)
		}
		return nil
	},
}

// authLoginTimeoutSeconds bounds how long `auth login` waits for the browser
// callback before failing.
var authLoginTimeoutSeconds = 300

// authLoginCmd performs the Authorization Code + PKCE flow on demand: it
// opens the system browser at the provider's authorization page, waits for
// the loopback callback, exchanges the code, and caches the token.
var authLoginCmd = &cobra.Command{
	Use:   "login <config>",
	Short: "Authorize OAuth 2.0 in the system browser",
	Long: `Perform the OAuth 2.0 Authorization Code + PKCE flow and cache the token.

<config> is a YAML or JSON file with the OAuth config keys:

  authorization_url: https://idp.example.com/authorize   # required
  token_url:         https://idp.example.com/token       # required
  client_id:         my-client                           # required
  client_secret:     s3cr3t                              # required
  redirect_uri:      http://127.0.0.1:8080/callback      # optional (loopback default)
  scope:             read write                          # optional

The system browser opens at the authorization page; the loopback callback
server on an ephemeral 127.0.0.1 port receives the redirect, verifies the
state, and exchanges the code for a token, which is cached in
<workspace>/.reqly/tokens.json. The client secret and tokens never print.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadAuthConfig(args[0])
		if err != nil {
			return err
		}
		// The login command always performs the authorization_code grant;
		// normalize the config so the cache key matches a request descriptor
		// with grant_type: authorization_code.
		cfg["grant_type"] = "authorization_code"
		root := findWorkspaceRoot(".")
		if root == "" {
			return fmt.Errorf("no workspace found: run reqly auth login inside a workspace (reqly.yaml)")
		}
		store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
		if err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		ctx, cancel := context.WithTimeout(ctx, time.Duration(authLoginTimeoutSeconds)*time.Second)
		defer cancel()

		src := &auth.AuthorizationCodeSource{
			Open: func(_ context.Context, authorizationURL string) error {
				fmt.Fprintf(cmd.OutOrStdout(), "opening %s in your browser…\n", authorizationURL)
				return launchBrowser(authorizationURL)
			},
		}
		cached := auth.NewCachedTokenSource(src, store, auth.TokenCacheKey(root, cfg))
		tok, err := cached.Token(ctx, cfg, variables.NewSet())
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		endpoint := cfg["token_url"]
		if endpoint == "" {
			endpoint = "(unknown)"
		}
		fmt.Fprintln(cmd.OutOrStdout(), "login complete — token cached")
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n",
			endpoint, formatExpiry(tok.Expiry), maskToken(tok.AccessToken),
			yesNo(tok.RefreshToken != ""))
		return nil
	},
}

// authLogoutCmd clears cached tokens for the nearest workspace.
var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear cached OAuth tokens for this workspace",
	Long: `Remove all cached OAuth 2.0 tokens for the workspace nearest to the
current directory, access and refresh tokens included. The next request that
needs a token re-acquires one.`,
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

// loadAuthConfig reads a flat OAuth auth config from a YAML or JSON file.
func loadAuthConfig(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read auth config: %w", err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse auth config %s: %w", path, err)
	}
	cfg := make(map[string]string, len(raw))
	for k, v := range raw {
		cfg[k] = fmt.Sprint(v)
	}
	return cfg, nil
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

// yesNo renders a boolean as yes/no for tabular status output.
func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func init() {
	authLoginCmd.Flags().IntVar(&authLoginTimeoutSeconds, "timeout", 300, "seconds to wait for the browser callback")
	authCmd.AddCommand(authLoginCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
