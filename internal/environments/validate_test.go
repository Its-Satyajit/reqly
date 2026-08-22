// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package environments

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSecretExposureWarnings(t *testing.T) {
	env := &Environment{
		Variables: map[string]string{
			"API_URL":     "https://api.example.com",
			"API_KEY":     "plain-key", // name looks secret -> warning
			"DB_PASSWORD": "hunter2",   // name looks secret -> warning
		},
		Secrets: map[string]string{
			"TOKEN": "tok-1",
		},
	}
	issues, err := Validate(env, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) == 0 {
		t.Fatal("expected secret-exposure warnings")
	}
	for _, want := range []string{"API_KEY", "DB_PASSWORD"} {
		if !containsIssue(issues, want) {
			t.Fatalf("expected warning mentioning %s, got %v", want, issues)
		}
	}
}

func TestValidateDuplicateKeyWarning(t *testing.T) {
	env := &Environment{
		Variables: map[string]string{"SHARED": "v"},
		Secrets:   map[string]string{"SHARED": "s"},
	}
	issues, err := Validate(env, "")
	if err != nil {
		t.Fatal(err)
	}
	if !containsIssue(issues, "SHARED") {
		t.Fatalf("expected duplicate-key warning, got %v", issues)
	}
}

func TestValidateCleanEnvironmentHasNoIssues(t *testing.T) {
	env := &Environment{
		Variables: map[string]string{"API_URL": "https://api.example.com"},
		Secrets:   map[string]string{"API_KEY": "k"},
	}
	issues, err := Validate(env, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
}

func TestValidateOutsideWorkspaceSkipsScanWithoutError(t *testing.T) {
	env := &Environment{
		Variables: map[string]string{"API_URL": "https://api.example.com"},
	}
	if _, err := Validate(env, t.TempDir()); err != nil {
		t.Fatalf("standalone env outside a workspace should validate: %v", err)
	}
}

func TestValidateUndefinedVariablesAcrossWorkspace(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "collections", "users"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reqly.yaml"), []byte("name: ws\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "collections", "users", "reqly.yaml"), []byte("name: users\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "collections", "users", "list.yaml"), []byte(`
request:
  method: GET
  url: https://api.example.com/users
  headers:
    - key: Authorization
      value: "Bearer {{API_KEY}}"
    - key: X-Region
      value: "{{REGION}}"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	env := &Environment{
		Secrets: map[string]string{"API_KEY": "k"},
		// REGION is missing from the environment.
	}
	issues, err := Validate(env, root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "REGION") && strings.Contains(issue.Message, "undefined") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected undefined-variable warning for REGION, got %v", issues)
	}
	if containsIssue(issues, "API_KEY") && strings.Contains(joinIssues(issues), "undefined") {
		t.Fatalf("API_KEY is defined by the environment and should not be flagged: %v", issues)
	}
}

func TestValidateUndefinedVariablesInTestFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reqly.yaml"), []byte("name: ws\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "check.yaml"), []byte(`
request:
  method: GET
  url: https://api.example.com/{{ACCOUNT_ID}}/status
tests:
  - type: status
    eq: 200
`), 0o644); err != nil {
		t.Fatal(err)
	}

	env := &Environment{}
	issues, err := Validate(env, root)
	if err != nil {
		t.Fatal(err)
	}
	if !containsIssue(issues, "ACCOUNT_ID") {
		t.Fatalf("expected undefined-variable warning for ACCOUNT_ID from test file, got %v", issues)
	}
}

func TestValidateDoesNotFlagRequestFileVariables(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reqly.yaml"), []byte("name: ws\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "collections", "reqly.yaml"), []byte("name: c\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "collections", "list.yaml"), []byte(`
variables:
  REGION: us-east-1
request:
  method: GET
  url: https://api.example.com/{{REGION}}/users
`), 0o644); err != nil {
		t.Fatal(err)
	}

	env := &Environment{}
	issues, err := Validate(env, root)
	if err != nil {
		t.Fatal(err)
	}
	if containsIssue(issues, "REGION") {
		t.Fatalf("REGION is defined by the request file and should not be flagged: %v", issues)
	}
}

func TestValidateDoesNotFlagGlobalVariables(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reqly.yaml"), []byte("name: ws\nvariables:\n  REGION: us-east-1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "collections", "reqly.yaml"), []byte("name: c\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "collections", "list.yaml"), []byte(`
request:
  method: GET
  url: https://api.example.com/{{REGION}}/users
`), 0o644); err != nil {
		t.Fatal(err)
	}

	env := &Environment{}
	issues, err := Validate(env, root)
	if err != nil {
		t.Fatal(err)
	}
	if containsIssue(issues, "REGION") {
		t.Fatalf("REGION is defined at global scope and should not be flagged: %v", issues)
	}
}

func containsIssue(issues []Issue, needle string) bool {
	for _, issue := range issues {
		if strings.Contains(issue.Message, needle) {
			return true
		}
	}
	return false
}

func joinIssues(issues []Issue) string {
	messages := make([]string, 0, len(issues))
	for _, issue := range issues {
		messages = append(messages, issue.Message)
	}
	return strings.Join(messages, "\n")
}
