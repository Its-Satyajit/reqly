package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)

// Manifest is plugins/<name>/manifest.json.
type Manifest struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Main         string   `json:"main"`
	Capabilities []string `json:"capabilities"`
}

// Plugin is a loaded plugin: manifest + Goja program.
type Plugin struct {
	Manifest Manifest
	Program  *goja.Program
	Dir      string
}

// Load validates manifest.json and compiles plugin.js via Goja in dir (plugins/<name>).
func Load(dir string) (*Plugin, error) {
	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("plugin: missing manifest %q: %w", manifestPath, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("plugin: invalid manifest %q: %w", manifestPath, err)
	}
	if m.Name == "" {
		return nil, fmt.Errorf("plugin: manifest missing name")
	}
	main := m.Main
	if main == "" {
		main = "plugin.js"
	}
	mainPath := filepath.Join(dir, main)
	code, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, fmt.Errorf("plugin: missing main %q: %w", mainPath, err)
	}
	prog, err := goja.Compile(main, string(code), false)
	if err != nil {
		return nil, fmt.Errorf("plugin: compile %q: %w", mainPath, err)
	}
	return &Plugin{Manifest: m, Program: prog, Dir: dir}, nil
}
