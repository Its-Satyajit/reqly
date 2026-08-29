package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestThemeList(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"theme", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("theme list: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "atlas-light") || !strings.Contains(out, "atlas-dark") {
		t.Fatalf("unexpected list output: %s", out)
	}
}

func TestThemeExport(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"theme", "export", "atlas-light"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("theme export: %v", err)
	}
	if !strings.Contains(buf.String(), "atlas-light") {
		t.Fatalf("unexpected export: %s", buf.String())
	}
}

func TestThemeImport(t *testing.T) {
	yamlData := []byte("id: my-custom\nlabel: My Custom\nappearance: light\ntokens:\n  primary: \"#123456\"\n")
	dir := t.TempDir()
	f := filepath.Join(dir, "theme.yaml")
	if err := os.WriteFile(f, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"theme", "import", f})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("theme import: %v", err)
	}
	if !strings.Contains(buf.String(), "my-custom") {
		t.Fatalf("unexpected import output: %s", buf.String())
	}
}
