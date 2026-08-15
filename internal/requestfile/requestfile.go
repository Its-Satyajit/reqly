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

package requestfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// File couples a request definition with the variables it uses. It is the
// plain-text, Git-native on-disk format shared by the CLI, Desktop, and MCP.
//
// The same format is accepted in JSON and YAML; parsing is format-agnostic
// (JSON is detected first, YAML used as a fallback).
type File struct {
	Name      string            `json:"name,omitempty" yaml:"name,omitempty"`
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Request   request.Request   `json:"request" yaml:"request"`

	// PreRequest is a JavaScript snippet run before the request is sent. It
	// runs in a sandbox with a `reqly` global for reading/writing variables
	// and mutating the outgoing request.
	PreRequest string `json:"preRequest,omitempty" yaml:"preRequest,omitempty"`
	// PostRequest is a JavaScript snippet run after the response arrives. It
	// can inspect reqly.response, set variables, and register reqly.test()s.
	PostRequest string `json:"postRequest,omitempty" yaml:"postRequest,omitempty"`
}

// LoadFile reads and parses a request file from disk.
func LoadFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read request file %q: %w", path, err)
	}
	f, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse request file %q: %w", path, err)
	}
	return f, nil
}

// Parse parses request file contents in JSON or YAML format.
func Parse(data []byte) (*File, error) {
	var f File
	if err := json.Unmarshal(data, &f); err == nil {
		return validate(&f)
	}
	var y File
	if err := yaml.Unmarshal(data, &y); err != nil {
		return nil, fmt.Errorf("invalid request file: %w", err)
	}
	return validate(&y)
}

func validate(f *File) (*File, error) {
	if f.Request.URL == "" {
		return nil, fmt.Errorf("request file requires a request.url")
	}
	return f, nil
}

// Variable looks up a variable by key.
func (f *File) Variable(key string) (string, bool) {
	value, ok := f.Variables[key]
	return value, ok
}

// VariablesSet returns the file's variables as a variables.Set using the
// request scope, ready for interpolation during execution.
func (f *File) VariablesSet() *variables.Set {
	set := variables.NewSet()
	for key, value := range f.Variables {
		set.Set(variables.ScopeRequest, key, value)
	}
	return set
}

// VariableNames returns the sorted set of variable keys defined in the file.
func (f *File) VariableNames() []string {
	keys := make([]string, 0, len(f.Variables))
	for key := range f.Variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// LooksLikeFile reports whether path names an existing regular file or carries
// a request-file extension, which the CLI uses to decide whether an argument
// is a request file or a URL.
func LooksLikeFile(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return false
	}
	if info, err := os.Stat(path); err == nil {
		return !info.IsDir()
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".yaml", ".yml":
		return true
	}
	return false
}
