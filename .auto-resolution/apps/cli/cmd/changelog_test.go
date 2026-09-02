package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChangelogCommand_MarkdownAndJSON(t *testing.T) {
	tempDir := t.TempDir()
	oldFile := filepath.Join(tempDir, "old.json")
	newFile := filepath.Join(tempDir, "new.json")

	_ = os.WriteFile(oldFile, []byte(`{"openapi": "3.0.0", "paths": {"/users": {"get": {"responses": {"200": {"description": "OK"}}}}}}`), 0600)
	_ = os.WriteFile(newFile, []byte(`{"openapi": "3.0.0", "paths": {"/users": {"get": {"responses": {"200": {"description": "OK"}}}, "post": {"responses": {"201": {"description": "Created"}}}}}}`), 0600)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"changelog", oldFile, newFile})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("changelog command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "API Changelog") || !strings.Contains(out, "Additions") {
		t.Errorf("expected markdown changelog output, got: %s", out)
	}

	// Test JSON format
	buf.Reset()
	rootCmd.SetArgs([]string{"changelog", oldFile, newFile, "--format", "json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("changelog JSON command failed: %v", err)
	}
	jsonOut := buf.String()
	if !strings.Contains(jsonOut, `"suggested_semver": "minor"`) {
		t.Errorf("expected JSON changelog output, got: %s", jsonOut)
	}
}

func TestChangelogCommand_FailOnBreaking(t *testing.T) {
	tempDir := t.TempDir()
	oldFile := filepath.Join(tempDir, "old.json")
	newFile := filepath.Join(tempDir, "new.json")

	_ = os.WriteFile(oldFile, []byte(`{"openapi": "3.0.0", "paths": {"/users": {"get": {"responses": {"200": {"description": "OK"}}}, "delete": {"responses": {"204": {"description": "Deleted"}}}}}}`), 0600)
	_ = os.WriteFile(newFile, []byte(`{"openapi": "3.0.0", "paths": {"/users": {"get": {"responses": {"200": {"description": "OK"}}}}}}`), 0600)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"changelog", oldFile, newFile, "--fail-on-breaking"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected error with --fail-on-breaking when breaking changes exist, got nil")
	}
}
