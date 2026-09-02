// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/Its-Satyajit/reqly/internal/theme"
)

// ThemeList returns built-in themes for the desktop theme picker.
func (s *AppService) ThemeList() []theme.Theme {
	return theme.BuiltInThemes()
}

// ThemeExport returns the YAML for a theme by id.
func (s *AppService) ThemeExport(id string) (string, error) {
	for _, t := range theme.BuiltInThemes() {
		if t.ID == id {
			data, err := theme.MarshalYAML(t)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("theme %q not found", id)
}

// ThemeImport parses and validates a theme YAML/JSON string and returns its CSS.
func (s *AppService) ThemeImport(yamlStr string) (string, error) {
	th, err := theme.Parse([]byte(yamlStr))
	if err != nil {
		return "", err
	}
	return theme.ToCSS(th)
}
