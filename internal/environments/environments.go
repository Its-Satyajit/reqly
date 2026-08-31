// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package environments

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
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
	Secrets     any               `yaml:"secrets,omitempty"`
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

	secrets := make(map[string]string)
	switch v := schema.Secrets.(type) {
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				secrets[s] = ""
			}
		}
	case map[string]any:
		for k, val := range v {
			if str, ok := val.(string); ok {
				secrets[k] = str
			} else {
				secrets[k] = ""
			}
		}
	}

	return &Environment{
		Name:        name,
		Description: schema.Description,
		Variables:   schema.Variables,
		Secrets:     secrets,
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
	if !validName(name) {
		return nil, fmt.Errorf("environment %q: name must be a plain filename component (letters, digits, '-', '_', '.', no path separators)", name)
	}
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

// Save writes env to environments/<name>.yaml in the environments/ directory
// discovered from dir. It errors when no environments/ directory exists, when
// the name is empty, or when the name is not a safe single filename component.
func Save(env *Environment, dir string) error {
	if env.Name == "" {
		return fmt.Errorf("save environment: empty name")
	}
	if !validName(env.Name) {
		return fmt.Errorf("save environment %q: name must be a plain filename component (letters, digits, '-', '_', '.', no path separators)", env.Name)
	}
	envDir, err := Discover(dir)
	if err != nil {
		return err
	}
	if envDir == "" {
		if dir == "" {
			return fmt.Errorf("save environment %q: no environments/ directory found", env.Name)
		}
		// In a workspace without an environments/ folder yet, auto-create it
		envDir = filepath.Join(dir, "environments")
		if err := os.MkdirAll(envDir, 0o755); err != nil {
			return fmt.Errorf("create environments directory %q: %w", envDir, err)
		}
	}
	path := filepath.Join(envDir, env.Name+".yaml")
	var secretNames []string
	if len(env.Secrets) > 0 {
		secretNames = make([]string, 0, len(env.Secrets))
		for k := range env.Secrets {
			secretNames = append(secretNames, k)
		}
		sort.Strings(secretNames)
	}
	schema := fileSchema{
		Description: env.Description,
		Variables:   env.Variables,
		Secrets:     secretNames,
	}
	data, err := yaml.Marshal(&schema)
	if err != nil {
		return fmt.Errorf("encode environment %q: %w", env.Name, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write environment file %q: %w", path, err)
	}
	return nil
}

// validName reports whether name is a safe single filename component for an
// environment: non-empty, not dot-only, and free of path separators and other
// characters that would break the YAML filename or the descriptor.
func validName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.ContainsAny(name, `/\`+" \t\n") {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return false
	}
	return true
}
