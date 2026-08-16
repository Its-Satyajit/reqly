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

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeWorkspace(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "environments"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "environments", "dev.yaml"), []byte(`
variables:
  REGION: us-west-2
secrets:
  API_KEY: dev-secret
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reqly.yaml"), []byte("environment: dev\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveAppEnvironmentFromWorkspaceDescriptor(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "")

	set, err := resolveAppEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := set.Resolve("REGION"); !ok || got != "us-west-2" {
		t.Fatalf("expected environment variable REGION=us-west-2, got %q (ok=%v)", got, ok)
	}
	if got, ok := set.Resolve("API_KEY"); !ok || got != "dev-secret" {
		t.Fatalf("expected environment secret API_KEY, got %q (ok=%v)", got, ok)
	}
}

func TestResolveAppEnvironmentFromREQLYEnv(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	// Second environment not selected by the descriptor.
	if err := os.WriteFile(filepath.Join(dir, "environments", "prod.yaml"), []byte("variables:\n  REGION: eu-central-1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "prod")

	set, err := resolveAppEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := set.Resolve("REGION"); !ok || got != "eu-central-1" {
		t.Fatalf("expected REQLY_ENV to win, got %q (ok=%v)", got, ok)
	}
}

func TestResolveAppEnvironmentWithoutWorkspaceIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "")

	set, err := resolveAppEnvironment()
	if err != nil {
		t.Fatalf("expected no error without workspace, got %v", err)
	}
	if set == nil {
		t.Fatal("expected a variable set (with process-env scope)")
	}
}
