// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package environments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestSelectionActivePrecedence(t *testing.T) {
	tests := []struct {
		name    string
		sel     Selection
		want    string
		wantAny string
	}{
		{
			name: "flag beats file and config",
			sel:  Selection{EnvFlag: "prod", FileEnv: "dev", ConfigEnv: "staging"},
			want: "prod",
		},
		{
			name: "file beats config",
			sel:  Selection{FileEnv: "dev", ConfigEnv: "staging"},
			want: "dev",
		},
		{
			name: "config alone",
			sel:  Selection{ConfigEnv: "staging"},
			want: "staging",
		},
		{
			name: "empty higher sources fall through",
			sel:  Selection{EnvFlag: "", FileEnv: "dev", ConfigEnv: ""},
			want: "dev",
		},
		{
			name: "no selection",
			sel:  Selection{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sel.Active(); got != tt.want {
				t.Fatalf("Active: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveSetPopulatesEnvironmentScope(t *testing.T) {
	root := t.TempDir()
	envDir := filepath.Join(root, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, "dev.yaml"), []byte(`
variables:
  API_URL: https://api.dev.example.com
secrets:
  API_KEY: dev-secret
`), 0o644); err != nil {
		t.Fatal(err)
	}

	set, masker, err := ResolveSet(root, Selection{EnvFlag: "dev"})
	if err != nil {
		t.Fatalf("ResolveSet: %v", err)
	}
	if got, _ := set.Resolve("API_URL"); got != "https://api.dev.example.com" {
		t.Fatalf("API_URL: got %q", got)
	}
	if got, _ := set.Resolve("API_KEY"); got != "dev-secret" {
		t.Fatalf("API_KEY: got %q", got)
	}
	if masker.Mask("dev-secret") != MaskedSecret {
		t.Fatal("masker should redact the environment's secret value")
	}
	if got := masker.Mask("https://api.dev.example.com"); got != "https://api.dev.example.com" {
		t.Fatalf("masker must not redact non-secret text: %q", got)
	}
}

func TestResolveSetNoSelectionReturnsEmptyEnvironmentScope(t *testing.T) {
	root := t.TempDir()
	set, masker, err := ResolveSet(root, Selection{})
	if err != nil {
		t.Fatalf("ResolveSet: %v", err)
	}
	if _, ok := set.Get(variables.ScopeEnvironment, "anything"); ok {
		t.Fatal("expected empty environment scope")
	}
	if masker == nil {
		t.Fatal("masker must never be nil")
	}
	if got := masker.Mask("plain text"); got != "plain text" {
		t.Fatalf("masker with no secrets must pass text through: %q", got)
	}
}

func TestResolveSetSelectedButMissingErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "environments"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ResolveSet(root, Selection{EnvFlag: "staging"}); err == nil {
		t.Fatal("expected error for selected-but-missing environment")
	}
}

func TestResolveSetAlwaysPopulatesProcessEnvScope(t *testing.T) {
	root := t.TempDir()
	t.Setenv("REQLY_TEST_PROCESS_VAR", "from-os")
	set, _, err := ResolveSet(root, Selection{})
	if err != nil {
		t.Fatalf("ResolveSet: %v", err)
	}
	if got, ok := set.Get(variables.ScopeProcessEnv, "REQLY_TEST_PROCESS_VAR"); !ok || got != "from-os" {
		t.Fatalf("process-env scope: got %q, %v", got, ok)
	}
}

func TestResolveSetLoadsDotEnvIntoProcessScope(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("DOTENV_VAR=from-dotenv\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	set, _, err := ResolveSet(root, Selection{})
	if err != nil {
		t.Fatalf("ResolveSet: %v", err)
	}
	if got, ok := set.Get(variables.ScopeProcessEnv, "DOTENV_VAR"); !ok || got != "from-dotenv" {
		t.Fatalf("dotenv in process scope: got %q, %v", got, ok)
	}
}
