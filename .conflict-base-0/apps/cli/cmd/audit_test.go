package cmd

import (
	"bytes"
	"os"
	"testing"
)

func TestAuditListEmpty(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"audit", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("audit list: %v", err)
	}
	if buf.String() != "No audit entries\n" {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestAuditClear(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"audit", "clear"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("audit clear: %v", err)
	}
	if buf.String() != "Audit log cleared\n" {
		t.Fatalf("unexpected: %q", buf.String())
	}
}
