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

package core

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEnvWorkspace(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "environments"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"dev.yaml": `
description: Local development
variables:
  REGION: us-west-2
  LOG_LEVEL: debug
secrets:
  API_KEY: dev-secret
`,
		"prod.yaml": `
description: Production
variables:
  REGION: eu-central-1
`,
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(root, "environments", name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "reqly.yaml"), []byte("environment: dev\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentServiceListReturnsEnvironmentsAndActive(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	got, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}
	if got.Active != "dev" {
		t.Fatalf("active = %q, want dev", got.Active)
	}
	if len(got.Environments) != 2 {
		t.Fatalf("environments = %d, want 2", len(got.Environments))
	}
	dev := got.Environments[0]
	if dev.Name != "dev" || dev.Description != "Local development" {
		t.Fatalf("dev = %+v", dev)
	}
	if dev.Variables["REGION"] != "us-west-2" {
		t.Fatalf("dev variables = %v", dev.Variables)
	}
	if !contains(dev.Secrets, "API_KEY") {
		t.Fatalf("dev secrets = %v, want API_KEY name", dev.Secrets)
	}
	for _, env := range got.Environments {
		for _, v := range env.Variables {
			if v == "dev-secret" {
				t.Fatalf("environment %s leaked a secret value in variables", env.Name)
			}
		}
	}
}

func TestEnvironmentServiceListWithoutEnvironmentsIsEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("environment: dev\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewEnvironmentService(dir)
	got, err := svc.List()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Active != "dev" {
		t.Fatalf("active = %q, want dev (descriptor still read)", got.Active)
	}
	if len(got.Environments) != 0 {
		t.Fatalf("environments = %d, want 0", len(got.Environments))
	}
}

func TestEnvironmentServiceListWithoutWorkspaceIsEmpty(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml

	svc := NewEnvironmentService(dir)
	got, err := svc.List()
	if err != nil {
		t.Fatalf("expected no error without workspace, got %v", err)
	}
	if got.Active != "" || len(got.Environments) != 0 {
		t.Fatalf("got active=%q envs=%d, want empty", got.Active, len(got.Environments))
	}
}

func TestEnvironmentServiceReadReturnsEnvironmentWithoutSecretValues(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	env, err := svc.Read("dev")
	if err != nil {
		t.Fatal(err)
	}
	if env.Name != "dev" || env.Description != "Local development" {
		t.Fatalf("env = %+v", env)
	}
	if env.Variables["LOG_LEVEL"] != "debug" {
		t.Fatalf("variables = %v", env.Variables)
	}
	if !contains(env.Secrets, "API_KEY") {
		t.Fatalf("secrets = %v, want API_KEY", env.Secrets)
	}
}

func TestEnvironmentServiceReadMissingEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	if _, err := svc.Read("staging"); err == nil {
		t.Fatal("expected error for missing environment, got nil")
	}
}

func TestEnvironmentServiceSetActivePersistsToDescriptor(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	if err := svc.SetActive("prod"); err != nil {
		t.Fatal(err)
	}
	got, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}
	if got.Active != "prod" {
		t.Fatalf("active = %q, want prod (persisted to descriptor)", got.Active)
	}
}

func TestEnvironmentServiceSetActiveClears(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	if err := svc.SetActive(""); err != nil {
		t.Fatal(err)
	}
	got, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}
	if got.Active != "" {
		t.Fatalf("active = %q, want cleared", got.Active)
	}
}

func TestEnvironmentServiceSetActiveWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml

	svc := NewEnvironmentService(dir)
	if err := svc.SetActive("prod"); err == nil {
		t.Fatal("expected error without workspace, got nil")
	}
}

func TestEnvironmentServiceCreateWritesEnvironmentFile(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	err := svc.Create("staging", "Staging server", map[string]string{"REGION": "ap-south-1"})
	if err != nil {
		t.Fatal(err)
	}

	env, err := svc.Read("staging")
	if err != nil {
		t.Fatal(err)
	}
	if env.Name != "staging" || env.Description != "Staging server" {
		t.Fatalf("env = %+v", env)
	}
	if env.Variables["REGION"] != "ap-south-1" {
		t.Fatalf("variables = %v", env.Variables)
	}
	if len(env.Secrets) != 0 {
		t.Fatalf("secrets = %v, want none", env.Secrets)
	}
}

func TestEnvironmentServiceCreateDuplicateErrors(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	if err := svc.Create("dev", "Duplicate", nil); err == nil {
		t.Fatal("expected error creating a duplicate environment, got nil")
	}
}

func TestEnvironmentServiceCreateInvalidNameErrors(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	for _, name := range []string{"", "a/b", "..", "bad name", ".hidden"} {
		if err := svc.Create(name, "Invalid", nil); err == nil {
			t.Fatalf("expected error for name %q, got nil", name)
		}
	}
}

func TestEnvironmentServiceCreateWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml

	svc := NewEnvironmentService(dir)
	if err := svc.Create("staging", "No workspace", nil); err == nil {
		t.Fatal("expected error without workspace, got nil")
	}
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
