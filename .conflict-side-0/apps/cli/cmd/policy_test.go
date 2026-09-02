package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPolicyShowDefault(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"policy", "show"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("policy show: %v", err)
	}
	if !strings.Contains(buf.String(), "requireAudit") {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestPolicyValidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(path, []byte("maxWorkflowSteps: 5\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"policy", "validate", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("policy validate: %v", err)
	}
	if !strings.Contains(buf.String(), "valid") {
		t.Fatalf("unexpected: %s", buf.String())
	}
}

func TestPolicyEnforce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(path, []byte("allowedActions:\n  - request.send\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	// Need to ensure we use the file we created
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"policy", "enforce", "--action", "request.send", "--resource", "x", "--file", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("enforce allowed: %v", err)
	}
	if !strings.Contains(buf.String(), "Allowed") {
		t.Fatalf("unexpected: %s", buf.String())
	}
	// Denied
	buf.Reset()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"policy", "enforce", "--action", "theme.import", "--resource", "x", "--file", path})
	if err := rootCmd.Execute(); err == nil {
		t.Fatalf("expected deny")
	}
}
