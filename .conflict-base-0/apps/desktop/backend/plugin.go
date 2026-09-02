// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/plugin"
)

// PluginView is the bridge view for a plugin.
type PluginView struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	Valid        bool     `json:"valid"`
	Error        string   `json:"error,omitempty"`
	Dir          string   `json:"dir"`
}

// PluginList returns plugins found under <root>/plugins/*.
func (s *AppService) PluginList() ([]PluginView, error) {
	root := s.root
	if root == "" {
		root = "."
	}
	pluginsDir := filepath.Join(root, "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []PluginView
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(pluginsDir, e.Name())
		pl, err := plugin.Load(dir)
		if err != nil {
			out = append(out, PluginView{Name: e.Name(), Valid: false, Error: err.Error(), Dir: dir})
			continue
		}
		out = append(out, PluginView{Name: pl.Manifest.Name, Version: pl.Manifest.Version, Capabilities: pl.Manifest.Capabilities, Valid: true, Dir: dir})
	}
	return out, nil
}

// PluginValidate validates a single plugin by name.
func (s *AppService) PluginValidate(name string) (*PluginView, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	root := s.root
	if root == "" {
		root = "."
	}
	dir := filepath.Join(root, "plugins", name)
	pl, err := plugin.Load(dir)
	if err != nil {
		return &PluginView{Name: name, Valid: false, Error: err.Error(), Dir: dir}, nil
	}
	return &PluginView{Name: pl.Manifest.Name, Version: pl.Manifest.Version, Capabilities: pl.Manifest.Capabilities, Valid: true, Dir: dir}, nil
}
