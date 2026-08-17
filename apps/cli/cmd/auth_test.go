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
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// writeCachedToken seeds a token store in the workspace root with one cached
// token for a synthetic config.
func writeCachedToken(t *testing.T, root, token string, expiry time.Time) {
	t.Helper()
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := json.Marshal(map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"endpoint":     "https://auth.example.com/token",
		"expiry":       expiry,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set("some-workspace:some-config", string(blob)); err != nil {
		t.Fatal(err)
	}
}

// chdirWorkspace moves the process into root for the duration of the test.
func chdirWorkspace(t *testing.T, root string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAuthStatusShowsCachedToken(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(1*time.Hour))
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{
		"https://auth.example.com/token",
		"very",   // masked prefix visible
		"oken",   // masked suffix visible
		"cached", // state
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	if strings.Contains(output, "very-secret-access-token") {
		t.Fatalf("status leaked the full token:\n%s", output)
	}
	if !strings.Contains(output, "****************") {
		t.Fatalf("expected masked stars, got:\n%s", output)
	}
}

func TestAuthStatusEmpty(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "no cached tokens") {
		t.Fatalf("expected 'no cached tokens', got:\n%s", out.String())
	}
}

func TestAuthStatusExpired(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(-1*time.Hour))
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "expired") {
		t.Fatalf("expected 'expired' state, got:\n%s", out.String())
	}
}

func TestAuthLogoutClearsTokens(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(1*time.Hour))
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "logout"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "cleared 1 cached token(s)") {
		t.Fatalf("expected 'cleared 1 cached token(s)', got:\n%s", out.String())
	}

	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := store.Keys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 0 {
		t.Fatalf("store still has keys after logout: %v", keys)
	}
}

func TestAuthLogoutEmpty(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "logout"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "cleared 0 cached token(s)") {
		t.Fatalf("expected 'cleared 0 cached token(s)', got:\n%s", out.String())
	}
}
