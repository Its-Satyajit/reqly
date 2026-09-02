package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	manifest := `{"name":"test","version":"1.0.0","main":"plugin.js","capabilities":["tag"]}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.js"), []byte(`module.exports = { tag: () => "hi" }`), 0644); err != nil {
		t.Fatalf("write plugin: %v", err)
	}
	p, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Manifest.Name != "test" {
		t.Fatalf("name %q", p.Manifest.Name)
	}
}

func TestLoadMissingManifest(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); err == nil {
		t.Fatalf("want error")
	}
}
