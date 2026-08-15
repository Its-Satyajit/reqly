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

package collections

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

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
