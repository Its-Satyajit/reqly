// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Theme defines a shareable UI theme. Tokens are CSS variable values
// without the -- prefix, e.g. Tokens["primary"] = "#c93517".
type Theme struct {
	ID         string            `json:"id" yaml:"id"`
	Label      string            `json:"label" yaml:"label"`
	Appearance string            `json:"appearance" yaml:"appearance"` // light | dark
	Tokens     map[string]string `json:"tokens,omitempty" yaml:"tokens,omitempty"`
}

var idRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var hexRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// BuiltInThemes returns the built-in themes that ship with Reqly.
func BuiltInThemes() []Theme {
	return []Theme{
		{ID: "atlas-light", Label: "Atlas Light", Appearance: "light", Tokens: map[string]string{"primary": "#c93517", "background": "#fbfbfa"}},
		{ID: "atlas-dark", Label: "Atlas Dark", Appearance: "dark", Tokens: map[string]string{"primary": "#ff6f52", "background": "#0d1015"}},
		{ID: "windows-11-light", Label: "Windows 11 Light", Appearance: "light", Tokens: map[string]string{"primary": "#0067c0", "background": "#f3f3f3"}},
		{ID: "windows-11-dark", Label: "Windows 11 Dark", Appearance: "dark", Tokens: map[string]string{"primary": "#4cc2ff", "background": "#202020"}},
		{ID: "windows-11", Label: "Windows 11", Appearance: "dark", Tokens: map[string]string{"primary": "#4cc2ff", "background": "#202020"}},
		{ID: "macos-tahoe-light", Label: "macOS Tahoe Light", Appearance: "light", Tokens: map[string]string{"primary": "#007aff", "background": "#f5f5f7"}},
		{ID: "macos-tahoe-dark", Label: "macOS Tahoe Dark", Appearance: "dark", Tokens: map[string]string{"primary": "#0a84ff", "background": "#1c1c1e"}},
		{ID: "macos-tahoe", Label: "macOS Tahoe", Appearance: "dark", Tokens: map[string]string{"primary": "#0a84ff", "background": "#1c1c1e"}},
		{ID: "linux-kde-light", Label: "Linux KDE Light", Appearance: "light", Tokens: map[string]string{"primary": "#3daee9", "background": "#eff0f1"}},
		{ID: "linux-kde-dark", Label: "Linux KDE Dark", Appearance: "dark", Tokens: map[string]string{"primary": "#3daee9", "background": "#31363b"}},
		{ID: "linux-kde", Label: "Linux KDE", Appearance: "dark", Tokens: map[string]string{"primary": "#3daee9", "background": "#31363b"}},
		{ID: "linux-gnome-light", Label: "Linux GNOME Light", Appearance: "light", Tokens: map[string]string{"primary": "#3584e4", "background": "#fafafb"}},
		{ID: "linux-gnome-dark", Label: "Linux GNOME Dark", Appearance: "dark", Tokens: map[string]string{"primary": "#3584e4", "background": "#222226"}},
		{ID: "linux-gnome", Label: "Linux GNOME", Appearance: "dark", Tokens: map[string]string{"primary": "#3584e4", "background": "#222226"}},
	}
}

// Validate checks a theme for required fields and token format.
func Validate(t Theme) error {
	if t.ID == "" {
		return fmt.Errorf("id is required")
	}
	if !idRe.MatchString(t.ID) {
		return fmt.Errorf("id %q must be kebab-case [a-z0-9-]", t.ID)
	}
	if strings.TrimSpace(t.Label) == "" {
		return fmt.Errorf("label is required")
	}
	if t.Appearance != "light" && t.Appearance != "dark" {
		return fmt.Errorf("appearance must be 'light' or 'dark'")
	}
	for k, v := range t.Tokens {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("token key cannot be empty")
		}
		if v != "" && !hexRe.MatchString(v) && !strings.HasPrefix(v, "hsl") && !strings.HasPrefix(v, "rgb") {
			// Allow hex or hsl/rgb for flexibility, but reject obvious garbage.
			if len(v) < 3 {
				return fmt.Errorf("token %q value %q is not a valid color", k, v)
			}
		}
	}
	return nil
}

// Parse parses a theme from YAML or JSON bytes (YAML is a superset of JSON).
func Parse(data []byte) (Theme, error) {
	var t Theme
	if err := yaml.Unmarshal(data, &t); err != nil {
		return Theme{}, fmt.Errorf("parse theme: %w", err)
	}
	if err := Validate(t); err != nil {
		return Theme{}, err
	}
	return t, nil
}

// MarshalYAML marshals a theme to YAML.
func MarshalYAML(t Theme) ([]byte, error) {
	if err := Validate(t); err != nil {
		return nil, err
	}
	return yaml.Marshal(t)
}

// MarshalJSON marshals a theme to JSON.
func MarshalJSON(t Theme) ([]byte, error) {
	if err := Validate(t); err != nil {
		return nil, err
	}
	return json.Marshal(t)
}

// ToCSS generates a CSS block for the theme, e.g. `[data-theme="my-theme"] { --primary: #... }`.
func ToCSS(t Theme) (string, error) {
	if err := Validate(t); err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`[data-theme="%s"] {`, t.ID))
	for k, v := range t.Tokens {
		// Normalize key to --kebab-case without -- prefix handling.
		key := strings.TrimPrefix(k, "--")
		b.WriteString(fmt.Sprintf(" --%s: %s;", key, v))
	}
	b.WriteString(" }")
	return b.String(), nil
}

// IsBuiltIn reports whether id is a built-in theme.
func IsBuiltIn(id string) bool {
	for _, t := range BuiltInThemes() {
		if t.ID == id {
			return true
		}
	}
	return false
}
