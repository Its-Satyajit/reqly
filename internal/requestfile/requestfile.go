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

package requestfile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"

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
	// Environment selects the environment to apply to this request. It is
	// overridden by the --env flag and REQLY_ENV at runtime.
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`

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
		return nil, fmt.Errorf("request file requires 'request.url'; ensure fields are nested under the 'request:' key")
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

// Save writes a request file back to disk, preserving the original format
// (JSON for .json files, YAML otherwise) and writing atomically via a temp
// file in the same directory followed by a rename. The request must have a
// non-empty URL. The destination file's permissions are preserved.
func Save(path string, f *File) error {
	data, err := marshal(path, f)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("write request file %q: %w", path, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := os.Chmod(tmpName, mode); err != nil {
		tmp.Close()
		return fmt.Errorf("write request file %q: %w", path, err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write request file %q: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write request file %q: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("write request file %q: %w", path, err)
	}
	return nil
}

// marshal serializes a File in its on-disk format: JSON for .json paths,
// YAML otherwise. The URL is validated before serializing.
func marshal(path string, f *File) ([]byte, error) {
	if _, err := validate(f); err != nil {
		return nil, err
	}
	if isJSONPath(path) {
		return json.MarshalIndent(f, "", "  ")
	}
	return yaml.Marshal(f)
}

// isJSONPath reports whether a path carries a .json extension.
func isJSONPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".json")
}

// Fingerprint returns a stable content hash of a request file's bytes, used
// to detect on-disk changes between the time a request was opened and saved.
func Fingerprint(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
