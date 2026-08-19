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
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/collections"
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

func TestEnvironmentServiceUpdatePersistsChangesAndPreservesSecrets(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	if err := svc.Update("dev", "Updated description", map[string]string{"REGION": "eu-west-1"}); err != nil {
		t.Fatal(err)
	}

	env, err := svc.Read("dev")
	if err != nil {
		t.Fatal(err)
	}
	if env.Description != "Updated description" {
		t.Fatalf("description = %q, want Updated description", env.Description)
	}
	if env.Variables["REGION"] != "eu-west-1" {
		t.Fatalf("variables = %v", env.Variables)
	}
	// Secrets stay on disk and are never exposed through Read: only the name
	// shows up, proving Update preserved the secret while overwriting the
	// description and variables.
	if len(env.Secrets) != 1 || env.Secrets[0] != "API_KEY" {
		t.Fatalf("secrets = %v, want [API_KEY]", env.Secrets)
	}
}

func TestEnvironmentServiceUpdateMissingEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	if err := svc.Update("nope", "x", nil); err == nil {
		t.Fatal("expected error updating a missing environment, got nil")
	}
}

func TestEnvironmentServiceUpdateWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml

	svc := NewEnvironmentService(dir)
	if err := svc.Update("dev", "x", nil); err == nil {
		t.Fatal("expected error without workspace, got nil")
	}
}

func TestEnvironmentServiceUpdateSecretsSetsAndRemoves(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir) // dev.yaml has secret API_KEY=dev-secret

	svc := NewEnvironmentService(dir)
	// Change API_KEY, add DB_PASSWORD, remove LOGIN_TOKEN (not present: no-op).
	err := svc.UpdateSecrets("dev", map[string]string{"API_KEY": "new-key", "DB_PASSWORD": "hunter2"}, []string{"LOGIN_TOKEN"})
	if err != nil {
		t.Fatal(err)
	}

	// Secret values are never read back through the service, so verify the
	// file on disk directly.
	data, err := os.ReadFile(filepath.Join(dir, "environments", "dev.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "API_KEY: new-key") {
		t.Fatalf("API_KEY not updated:\n%s", content)
	}
	if !strings.Contains(content, "DB_PASSWORD: hunter2") {
		t.Fatalf("DB_PASSWORD not added:\n%s", content)
	}
	if strings.Contains(content, "dev-secret") {
		t.Fatalf("old API_KEY value leaked:\n%s", content)
	}
}

func TestEnvironmentServiceUpdateSecretsMissingEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	if err := svc.UpdateSecrets("nope", nil, nil); err == nil {
		t.Fatal("expected missing-environment error, got nil")
	}
}

func TestEnvironmentServiceUpdateSecretsWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml

	svc := NewEnvironmentService(dir)
	if err := svc.UpdateSecrets("dev", nil, nil); err == nil {
		t.Fatal("expected error without workspace, got nil")
	}
}

func TestEnvironmentServiceDeleteRemovesFileAndClearsActive(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir) // active = dev, environments: dev, prod

	svc := NewEnvironmentService(dir)
	if err := svc.Delete("dev"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "environments", "dev.yaml")); !os.IsNotExist(err) {
		t.Fatalf("dev.yaml still exists: %v", err)
	}
	active := collections.WorkspaceEnvironment(dir)
	if active != "" {
		t.Fatalf("active = %q, want cleared", active)
	}
	list, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Environments) != 1 || list.Environments[0].Name != "prod" {
		t.Fatalf("environments = %+v, want [prod]", list.Environments)
	}
}

func TestEnvironmentServiceDeleteNonActiveKeepsActive(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir) // active = dev

	svc := NewEnvironmentService(dir)
	if err := svc.Delete("prod"); err != nil {
		t.Fatal(err)
	}
	active := collections.WorkspaceEnvironment(dir)
	if active != "dev" {
		t.Fatalf("active = %q, want dev", active)
	}
}

func TestEnvironmentServiceDeleteMissingEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir)

	svc := NewEnvironmentService(dir)
	if err := svc.Delete("nope"); err == nil {
		t.Fatal("expected missing-environment error, got nil")
	}
}

func TestEnvironmentServiceDeleteWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml

	svc := NewEnvironmentService(dir)
	if err := svc.Delete("dev"); err == nil {
		t.Fatal("expected error without workspace, got nil")
	}
}

func TestEnvironmentServiceUpdateSecretsOnEnvWithoutSecrets(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "environments"), 0o755); err != nil {
		t.Fatal(err)
	}
	// No secrets: key present.
	if err := os.WriteFile(filepath.Join(dir, "environments", "dev.yaml"), []byte("description: no secrets\nvariables:\n  A: b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("environment: dev\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewEnvironmentService(dir)
	if err := svc.UpdateSecrets("dev", map[string]string{"NEW_KEY": "v"}, nil); err != nil {
		t.Fatalf("UpdateSecrets on env without secrets: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "environments", "dev.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "NEW_KEY: v") {
		t.Fatalf("new secret not written:\n%s", data)
	}
}

func TestEnvironmentServiceUpdateRejectsVariableSecretCollision(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir) // dev.yaml has secret API_KEY

	svc := NewEnvironmentService(dir)
	err := svc.Update("dev", "x", map[string]string{"API_KEY": "leaked-in-variables"})
	if err == nil || !strings.Contains(err.Error(), "both variables and secrets") {
		t.Fatalf("err = %v, want variable/secret collision error", err)
	}
}

func TestEnvironmentServiceUpdateSecretsRejectsVariableSecretCollision(t *testing.T) {
	dir := t.TempDir()
	writeEnvWorkspace(t, dir) // dev.yaml has variable REGION

	svc := NewEnvironmentService(dir)
	err := svc.UpdateSecrets("dev", map[string]string{"REGION": "secret-now"}, nil)
	if err == nil || !strings.Contains(err.Error(), "both variables and secrets") {
		t.Fatalf("err = %v, want variable/secret collision error", err)
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
