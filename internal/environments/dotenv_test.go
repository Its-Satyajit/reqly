// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package environments

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseDotEnv(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "simple",
			input: "API_URL=https://api.example.com\nAPI_KEY=abc123\n",
			want:  map[string]string{"API_URL": "https://api.example.com", "API_KEY": "abc123"},
		},
		{
			name:  "comments and blank lines",
			input: "# comment\n\nAPI_URL=https://api.example.com\n# another\nREGION=us-east-1\n",
			want:  map[string]string{"API_URL": "https://api.example.com", "REGION": "us-east-1"},
		},
		{
			name:  "double-quoted values",
			input: `MSG="hello world"` + "\n" + `URL="https://x.com?a=1&b=2"` + "\n",
			want:  map[string]string{"MSG": "hello world", "URL": "https://x.com?a=1&b=2"},
		},
		{
			name:  "single-quoted values",
			input: "PASS='p@ss word'\n",
			want:  map[string]string{"PASS": "p@ss word"},
		},
		{
			name:  "inline comment",
			input: "KEY=value # comment\n",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "export prefix",
			input: "export API_URL=https://api.example.com\n",
			want:  map[string]string{"API_URL": "https://api.example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDotEnv(tt.input)
			if err != nil {
				t.Fatalf("ParseDotEnv: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseDotEnv: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDotEnvIgnoresInvalidLines(t *testing.T) {
	got, err := ParseDotEnv("THIS IS NOT A PAIR\n=novalue\nKEY=value\n")
	if err != nil {
		t.Fatalf("ParseDotEnv: %v", err)
	}
	if len(got) != 1 || got["KEY"] != "value" {
		t.Fatalf("ParseDotEnv: got %v", got)
	}
}

func TestLoadDotEnvDiscoversNearestAndOSWins(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "sub", "dir")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("SHARED=from-dotenv\nONLY_DOTENV=dotenv-val\nOVERRIDDEN=dotenv-val\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("OVERRIDDEN", "os-wins")
	t.Setenv("ONLY_OS", "os-val")

	vars, err := LoadDotEnv(nested)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if vars["SHARED"] != "from-dotenv" {
		t.Fatalf("SHARED: got %q", vars["SHARED"])
	}
	if vars["OVERRIDDEN"] != "os-wins" {
		t.Fatalf("OVERRIDDEN: got %q, want OS wins over .env", vars["OVERRIDDEN"])
	}
	if vars["ONLY_OS"] != "os-val" {
		t.Fatalf("ONLY_OS: got %q", vars["ONLY_OS"])
	}
	if _, ok := vars["ONLY_DOTENV"]; !ok {
		t.Fatal("ONLY_DOTENV missing")
	}
}

func TestLoadDotEnvWithoutFileReturnsOSOnly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("REQLY_TEST_OS_VAR", "present")
	vars, err := LoadDotEnv(dir)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if vars["REQLY_TEST_OS_VAR"] != "present" {
		t.Fatalf("expected OS var present, got %q", vars["REQLY_TEST_OS_VAR"])
	}
}
