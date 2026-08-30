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

package collections

import (
	"encoding/json"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"

	"github.com/Its-Satyajit/reqly/internal/request"
)

// configFileName is the descriptor file that marks a directory as a workspace,
// collection, or folder. It must be valid JSON or YAML.
const configFileName = "reqly.yaml"

// Config is the shared descriptor format for a workspace, collection, or
// folder. Every property is optional and inherited down the hierarchy
// (Workspace → Collection → Folder → Request) unless overridden.
type Config struct {
	// Name is the display name of this container.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// BaseURL is the base URL. Empty inherits the parent's. A relative value
	// (no scheme) is joined onto the parent's resolved base URL.
	BaseURL string `json:"baseURL,omitempty" yaml:"baseURL,omitempty"`
	// Headers are merged with the parent's; the child wins on key conflicts.
	Headers []request.Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	// Auth replaces the parent's auth entirely when its Type is non-empty.
	Auth request.Auth `json:"auth,omitempty" yaml:"auth,omitempty"`
	// Variables feed the scope chain (workspace=global, collection, folder).
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	// Environment is the active environment name for the workspace. It is
	// overridden by a file's environment: field, the --env flag, and
	// REQLY_ENV at runtime.
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// loadConfig reads a Config from dir/reqly.yaml (JSON or YAML). It returns a
// zero Config and false when the file is absent.
func loadConfig(dir string) (Config, bool, error) {
	path := joinConfigPath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, false, nil
		}
		return Config{}, false, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err == nil {
		return cfg, true, nil
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, false, fmt.Errorf("parse config %q: %w", path, err)
	}
	return cfg, true, nil
}

// joinConfigPath returns the path to a container's config file.
func joinConfigPath(dir string) string {
	return dir + string(os.PathSeparator) + configFileName
}

// WorkspaceEnvironment returns the environment: field of the workspace
// descriptor nearest to dir, or "" when no workspace is present.
func WorkspaceEnvironment(dir string) string {
	ws, err := LoadWorkspace(dir)
	if err != nil || ws == nil {
		return ""
	}
	return ws.Config.Environment
}

// IsWorkspace reports whether dir contains a workspace descriptor (reqly.yaml).
func IsWorkspace(dir string) bool {
	_, ok, err := loadConfig(dir)
	return err == nil && ok
}

// SetWorkspaceEnvironment persists name as the active environment in the
// workspace descriptor at dir, preserving all other fields. It creates the
// descriptor when it does not exist.
func SetWorkspaceEnvironment(dir string, name string) error {
	cfg, _, err := loadConfig(dir)
	if err != nil {
		return err
	}
	cfg.Environment = name
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	path := joinConfigPath(dir)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}
	return nil
}
