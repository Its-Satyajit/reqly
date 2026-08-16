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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEnv(t *testing.T, dir, name, contents string) {
	t.Helper()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, name+".yaml"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func executeCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), err
}

func TestEnvListCommand(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "variables:\n  A: 1\n")
	writeEnv(t, dir, "prod", "variables:\n  A: 2\n")
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "list")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"dev", "prod"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output:\n%s", want, output)
		}
	}
}

func TestEnvListNoEnvironmentsErrors(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if _, err := executeCmd(t, "env", "list"); err == nil {
		t.Fatal("expected error when no environments/ dir exists")
	}
}

func TestEnvShowMasksSecrets(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", `
description: Development
variables:
  API_URL: https://api.dev.example.com
secrets:
  API_KEY: dev-super-secret
`)
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "show", "dev")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "https://api.dev.example.com") {
		t.Fatalf("expected variable visible:\n%s", output)
	}
	if strings.Contains(output, "dev-super-secret") {
		t.Fatal("secret value leaked in output")
	}
	if !strings.Contains(output, "[SECRET]") {
		t.Fatalf("expected [SECRET] masking:\n%s", output)
	}
}

func TestEnvShowDefaultsToActiveEnvironment(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "prod", "variables:\n  REGION: us-east-1\n")
	if err := os.MkdirAll(filepath.Join(dir, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("environment: prod\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "show")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "us-east-1") {
		t.Fatalf("expected active env vars:\n%s", output)
	}
}

func TestEnvShowMissingEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if _, err := executeCmd(t, "env", "show", "nope"); err == nil {
		t.Fatal("expected error for missing environment")
	}
}

func TestEnvUsePersistsToWorkspaceDescriptor(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "prod", "variables:\n  A: 2\n")
	if err := os.MkdirAll(filepath.Join(dir, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("name: workspace\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	if _, err := executeCmd(t, "env", "use", "prod"); err != nil {
		t.Fatal(err)
	}

	descriptor, err := os.ReadFile(filepath.Join(dir, "reqly.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(descriptor), "prod") {
		t.Fatalf("expected environment: prod in descriptor:\n%s", descriptor)
	}
}

func TestEnvUseMissingEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("name: workspace\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	if _, err := executeCmd(t, "env", "use", "nope"); err == nil {
		t.Fatal("expected error for missing environment")
	}
}

func TestEnvUseWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "variables:\n  A: 1\n")
	t.Chdir(dir)

	if _, err := executeCmd(t, "env", "use", "dev"); err == nil {
		t.Fatal("expected error when no workspace descriptor exists")
	}
}

func TestEnvValidateReportsSecretExposure(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "variables:\n  API_KEY: plain\n")
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "validate", "dev")
	if err == nil {
		t.Fatal("expected validate to fail on secret exposure")
	}
	if !strings.Contains(output, "API_KEY") {
		t.Fatalf("expected API_KEY warning:\n%s", output)
	}
}

func TestEnvValidateCleanPasses(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "variables:\n  API_URL: https://example.com\nsecrets:\n  API_KEY: k\n")
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "validate", "dev")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "is valid") {
		t.Fatalf("expected valid message:\n%s", output)
	}
}

func TestEnvValidateScansWorkspaceRequests(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "secrets:\n  API_KEY: k\n")
	if err := os.MkdirAll(filepath.Join(dir, "collections", "users"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("name: ws\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "collections", "users", "reqly.yaml"), []byte("name: users\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "collections", "users", "list.yaml"), []byte(`
request:
  method: GET
  url: https://example.com
  headers:
    - key: Authorization
      value: "Bearer {{API_KEY}}"
    - key: X-Region
      value: "{{REGION}}"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "validate", "dev")
	if err == nil {
		t.Fatal("expected validate to fail on undefined REGION")
	}
	if !strings.Contains(output, "REGION") {
		t.Fatalf("expected undefined REGION warning:\n%s", output)
	}
	if strings.Contains(output, "API_KEY") {
		t.Fatalf("API_KEY is defined and should not be flagged:\n%s", output)
	}
}

func TestEnvValidateDefaultsToActiveEnvironment(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "prod", "variables:\n  API_KEY: plain\n")
	if err := os.MkdirAll(filepath.Join(dir, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("environment: prod\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "validate")
	if err == nil {
		t.Fatal("expected validate to fail")
	}
	if !strings.Contains(output, "API_KEY") {
		t.Fatalf("expected API_KEY warning:\n%s", output)
	}
}

func TestEnvDiffShowsChanges(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "variables:\n  KEEP: same\n  CHANGE: from\nsecrets:\n  SECRET_CHANGE: old-secret\n")
	writeEnv(t, dir, "prod", "variables:\n  KEEP: same\n  CHANGE: to\n  NEW: added\nsecrets:\n  SECRET_CHANGE: new-secret\n")
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "diff", "dev", "prod")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"CHANGE", "NEW", "SECRET_CHANGE"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in diff output:\n%s", want, output)
		}
	}
	if strings.Contains(output, "old-secret") || strings.Contains(output, "new-secret") {
		t.Fatal("secret values leaked in diff output")
	}
	if strings.Contains(output, "KEEP") {
		t.Fatalf("unchanged key should not appear:\n%s", output)
	}
}

func TestEnvDiffIdenticalEnvironments(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "variables:\n  A: 1\nsecrets:\n  S: x\n")
	writeEnv(t, dir, "prod", "variables:\n  A: 1\nsecrets:\n  S: x\n")
	t.Chdir(dir)

	output, err := executeCmd(t, "env", "diff", "dev", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "No structural changes found.") {
		t.Fatalf("expected identical message:\n%s", output)
	}
}

func TestEnvDiffMissingEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "variables:\n  A: 1\n")
	t.Chdir(dir)

	if _, err := executeCmd(t, "env", "diff", "dev", "nope"); err == nil {
		t.Fatal("expected error for missing environment")
	}
}
