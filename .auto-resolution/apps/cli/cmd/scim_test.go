package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSCIMUserCreate(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"scim", "user", "create", "--username", "alice", "--email", "alice@example.com"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("scim user create: %v", err)
	}
	if !strings.Contains(buf.String(), "alice") {
		t.Fatalf("unexpected: %s", buf.String())
	}
}

func TestSCIMUserList(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"scim", "user", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("scim user list: %v", err)
	}
	// In-memory store is empty per invocation, so expect "No users"
	if !strings.Contains(buf.String(), "No users") {
		t.Fatalf("unexpected: %s", buf.String())
	}
}
