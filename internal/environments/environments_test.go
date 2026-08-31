// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package environments

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadParsesEnvironmentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dev.yaml")
	if err := os.WriteFile(path, []byte(`
description: Development environment
variables:
  API_URL: https://api.dev.example.com
  REGION: us-east-1
secrets:
  API_KEY: dev-key-123
  DB_PASSWORD: hunter2
`), 0o644); err != nil {
		t.Fatal(err)
	}

	env, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if env.Name != "dev" {
		t.Fatalf("name: got %q, want %q", env.Name, "dev")
	}
	if env.Description != "Development environment" {
		t.Fatalf("description: got %q", env.Description)
	}
	wantVars := map[string]string{"API_URL": "https://api.dev.example.com", "REGION": "us-east-1"}
	if !reflect.DeepEqual(env.Variables, wantVars) {
		t.Fatalf("variables: got %v, want %v", env.Variables, wantVars)
	}
	wantSecrets := map[string]string{"API_KEY": "dev-key-123", "DB_PASSWORD": "hunter2"}
	if !reflect.DeepEqual(env.Secrets, wantSecrets) {
		t.Fatalf("secrets: got %v, want %v", env.Secrets, wantSecrets)
	}
}

func TestLoadRejectsMalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.yaml")
	if err := os.WriteFile(path, []byte("variables: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestLoadRejectsUnknownStructure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prod.yaml")
	if err := os.WriteFile(path, []byte("baseURL: https://prod.example.com\nname: something-else\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Unknown keys are tolerated (forward compatibility); the name is always
	// derived from the filename.
	if env.Name != "prod" {
		t.Fatalf("name: got %q, want %q", env.Name, "prod")
	}
	if len(env.Variables) != 0 || len(env.Secrets) != 0 {
		t.Fatalf("unexpected variables/secrets: %v %v", env.Variables, env.Secrets)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(filepath.Join(dir, "nope.yaml")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDiscoverFindsNearestEnvironmentsDir(t *testing.T) {
	root := t.TempDir()
	envDir := filepath.Join(root, "api", "v1", "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A deeper project dir inside the layout should find the same nearest dir.
	start := filepath.Join(root, "api", "v1", "requests")

	got, err := Discover(start)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	want, _ := filepath.Abs(envDir)
	if got != want {
		t.Fatalf("Discover: got %q, want %q", got, want)
	}
}

func TestDiscoverReturnsEmptyWhenNoEnvironmentsDir(t *testing.T) {
	start := t.TempDir()
	got, err := Discover(start)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != "" {
		t.Fatalf("Discover: got %q, want empty", got)
	}
}

func TestReadLoadsNamedEnvironment(t *testing.T) {
	root := t.TempDir()
	envDir := filepath.Join(root, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, "prod.yaml"), []byte("secrets:\n  KEY: prod-9\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env, err := Read("prod", root)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if env.Name != "prod" || env.Secrets["KEY"] != "prod-9" {
		t.Fatalf("Read: unexpected env %+v", env)
	}
}

func TestReadErrorsWhenEnvironmentMissing(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "environments"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Read("staging", root); err == nil {
		t.Fatal("expected error for selected-but-missing environment")
	}
}

func TestReadErrorsWhenNoEnvironmentsDir(t *testing.T) {
	start := t.TempDir()
	if _, err := Read("dev", start); err == nil {
		t.Fatal("expected error when no environments dir exists")
	}
}

func TestListReturnsEnvironmentNames(t *testing.T) {
	root := t.TempDir()
	envDir := filepath.Join(root, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"dev.yaml", "prod.yaml", "README.md"} {
		if err := os.WriteFile(filepath.Join(envDir, name), []byte("variables:\n  A: 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	names, err := List(root)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 2 || !contains(names, "dev") || !contains(names, "prod") {
		t.Fatalf("List: got %v", names)
	}
}

func TestListErrorsWhenNoEnvironmentsDir(t *testing.T) {
	start := t.TempDir()
	if _, err := List(start); err == nil {
		t.Fatal("expected error when no environments dir exists")
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func TestMaskReplacesSecretValues(t *testing.T) {
	env := &Environment{
		Secrets: map[string]string{"API_KEY": "topsecret", "DB_PASSWORD": "hunter2"},
	}
	got := env.Mask("headers: Authorization: Bearer topsecret and hunter2 db")
	if got != "headers: Authorization: Bearer [SECRET] and [SECRET] db" {
		t.Fatalf("Mask: got %q", got)
	}
}

func TestMaskLeavesNonSecretTextUntouched(t *testing.T) {
	env := &Environment{Secrets: map[string]string{"API_KEY": "s3cr3t"}}
	got := env.Mask("plain text with an unrelated token in it")
	if got != "plain text with an unrelated token in it" {
		t.Fatalf("Mask: got %q", got)
	}
}

func TestSaveWritesEnvironmentFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	env := &Environment{
		Name:        "staging",
		Description: "Staging server",
		Variables:   map[string]string{"REGION": "ap-south-1"},
		Secrets:     map[string]string{"API_KEY": "staging-key"},
	}
	if err := Save(env, dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(filepath.Join(envDir, "staging.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "staging" || got.Description != "Staging server" {
		t.Fatalf("got = %+v", got)
	}
	if !reflect.DeepEqual(got.Variables, env.Variables) {
		t.Fatalf("variables = %v, want %v", got.Variables, env.Variables)
	}
	wantSecrets := map[string]string{"API_KEY": ""}
	if !reflect.DeepEqual(got.Secrets, wantSecrets) {
		t.Fatalf("secrets = %v, want %v", got.Secrets, wantSecrets)
	}
}

func TestSaveWithoutEnvironmentsDirCreatesDirectory(t *testing.T) {
	dir := t.TempDir() // no environments/ subdir initially
	env := &Environment{Name: "dev", Description: "New env", Variables: map[string]string{"A": "1"}}
	if err := Save(env, dir); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(filepath.Join(dir, "environments", "dev.yaml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Name != "dev" || got.Variables["A"] != "1" {
		t.Fatalf("got: %+v", got)
	}
}

func TestLoadParsesSecretsList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "staging.yaml")
	if err := os.WriteFile(path, []byte(`
description: Staging environment
variables:
  REGION: ap-south-1
secrets:
  - API_KEY
  - DB_PASSWORD
`), 0o644); err != nil {
		t.Fatal(err)
	}

	env, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if env.Name != "staging" {
		t.Fatalf("name: got %q, want staging", env.Name)
	}
	wantSecrets := map[string]string{"API_KEY": "", "DB_PASSWORD": ""}
	if !reflect.DeepEqual(env.Secrets, wantSecrets) {
		t.Fatalf("secrets: got %v, want %v", env.Secrets, wantSecrets)
	}
}

func TestSaveWritesSecretsAsListWithoutValues(t *testing.T) {
	dir := t.TempDir()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	env := &Environment{
		Name:        "prod",
		Description: "Production",
		Variables:   map[string]string{"REGION": "us-east-1"},
		Secrets:     map[string]string{"API_KEY": "supersecretvalue", "DB_PASSWORD": "secretpassword"},
	}
	if err := Save(env, dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(envDir, "prod.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	// Check that secret values are NEVER written into the YAML file
	if strings.Contains(content, "supersecretvalue") || strings.Contains(content, "secretpassword") {
		t.Fatalf("secret values leaked into YAML file on disk:\n%s", content)
	}

	got, err := Load(filepath.Join(envDir, "prod.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Secrets["API_KEY"]; !ok {
		t.Fatalf("expected API_KEY in secrets map, got: %v", got.Secrets)
	}
	if _, ok := got.Secrets["DB_PASSWORD"]; !ok {
		t.Fatalf("expected DB_PASSWORD in secrets map, got: %v", got.Secrets)
	}
}
