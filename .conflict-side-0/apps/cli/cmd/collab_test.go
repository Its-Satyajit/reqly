package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCollabAddAndList(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"collab", "add", "--user", "alice", "--role", "admin"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("collab add: %v", err)
	}
	if !strings.Contains(buf.String(), "alice") {
		t.Fatalf("unexpected add: %s", buf.String())
	}
	buf.Reset()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"collab", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("collab list: %v", err)
	}
	if !strings.Contains(buf.String(), "alice") || !strings.Contains(buf.String(), "admin") {
		t.Fatalf("unexpected list: %s", buf.String())
	}
	buf.Reset()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"collab", "remove", "--user", "alice"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("collab remove: %v", err)
	}
	buf.Reset()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"collab", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("collab list after remove: %v", err)
	}
	if !strings.Contains(buf.String(), "No collaborators") {
		t.Fatalf("unexpected after remove: %s", buf.String())
	}
}
