package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRBACList(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"rbac", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rbac list: %v", err)
	}
	if !strings.Contains(buf.String(), "admin") {
		t.Fatalf("unexpected: %s", buf.String())
	}
}

func TestRBACCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rbac.yaml")
	if err := os.WriteFile(path, []byte("roles:\n  admin:\n    name: admin\n    permissions: [\"*\"]\n  viewer:\n    name: viewer\n    permissions: [\"request.send\"]\nuserRoles:\n  bob: viewer\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"rbac", "check", "--user", "bob", "--action", "request.send", "--resource", "x", "--file", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("check allowed: %v", err)
	}
	if !strings.Contains(buf.String(), "Allowed") {
		t.Fatalf("unexpected: %s", buf.String())
	}
	buf.Reset()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"rbac", "check", "--user", "bob", "--action", "workflow.run", "--resource", "x", "--file", path})
	if err := rootCmd.Execute(); err == nil {
		t.Fatalf("expected deny")
	}
}
