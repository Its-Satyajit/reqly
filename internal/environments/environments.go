// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package environments

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Environment is a named set of variables plus optional secrets, stored as a
// Git-native YAML file under environments/<name>.yaml. The name comes from the
// filename, not the file contents.
type Environment struct {
	Name        string
	Description string
	Variables   map[string]string
	Secrets     map[string]string
}

// fileSchema is the on-disk YAML layout of an environment file.
type fileSchema struct {
	Description string            `yaml:"description,omitempty"`
	Variables   map[string]string `yaml:"variables,omitempty"`
	Secrets     map[string]string `yaml:"secrets,omitempty"`
}

// Load parses an environment file, deriving the environment name from the
// filename. Unknown keys are tolerated for forward compatibility.
func Load(path string) (*Environment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read environment file %q: %w", path, err)
	}
	var schema fileSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("invalid environment file %q: %w", path, err)
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if name == "" {
		return nil, fmt.Errorf("invalid environment file %q: empty name", path)
	}
	return &Environment{
		Name:        name,
		Description: schema.Description,
		Variables:   schema.Variables,
		Secrets:     schema.Secrets,
	}, nil
}

// Discover walks up from dir to the nearest directory containing an
// environments/ subdirectory, returning its absolute path. It returns an empty
// string when no environments/ directory exists anywhere up the tree; that is
// not an error (an env-less project is valid).
func Discover(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", dir, err)
	}
	for {
		candidate := filepath.Join(abs, "environments")
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", nil
		}
		abs = parent
	}
}

// Read loads the named environment relative to the directory discovered from
// dir. A selected-but-missing environment is a hard error.
func Read(name, dir string) (*Environment, error) {
	envDir, err := Discover(dir)
	if err != nil {
		return nil, err
	}
	if envDir == "" {
		return nil, fmt.Errorf("environment %q not found: no environments/ directory", name)
	}
	path := filepath.Join(envDir, name+".yaml")
	env, err := Load(path)
	if err != nil {
		return nil, fmt.Errorf("environment %q: %w", name, err)
	}
	return env, nil
}

// List returns the names of the environments in the nearest environments/
// directory discovered from dir, in lexical order. Only *.yaml files count;
// any other file (e.g. README.md) is ignored.
func List(dir string) ([]string, error) {
	envDir, err := Discover(dir)
	if err != nil {
		return nil, err
	}
	if envDir == "" {
		return nil, fmt.Errorf("no environments/ directory found")
	}
	entries, err := os.ReadDir(envDir)
	if err != nil {
		return nil, fmt.Errorf("read environments directory %q: %w", envDir, err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
	}
	sort.Strings(names)
	return names, nil
}

// Mask replaces every secret value appearing in text with [SECRET]. Values are
// matched longest-first so a secret that is a prefix of another never gets
// shadowed. Non-secret text is left untouched.
func (e *Environment) Mask(text string) string {
	return NewMasker(e.SecretValues()...).Mask(text)
}
