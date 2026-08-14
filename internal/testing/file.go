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

package testing

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Its-Satyajit/reqly/internal/request"
)

// TestFile couples a request definition with the assertions that run against
// its response. It is the on-disk format shared by the CLI, Desktop, and MCP.
type TestFile struct {
	Name    string          `json:"name"`
	Request request.Request `json:"request"`
	Tests   []Test          `json:"tests"`
}

// Suite converts a TestFile into a Suite for evaluation.
func (f TestFile) Suite() Suite {
	return Suite{Name: f.Name, Tests: f.Tests}
}

// LoadTestFile reads and parses a test file from disk.
func LoadTestFile(path string) (*TestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read test file %q: %w", path, err)
	}
	return ParseTestFile(data)
}

// ParseTestFile parses test file contents.
func ParseTestFile(data []byte) (*TestFile, error) {
	var tf TestFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("parse test file: %w", err)
	}
	if tf.Request.URL == "" {
		return nil, fmt.Errorf("test file requires a request.url")
	}
	if len(tf.Tests) == 0 {
		return nil, fmt.Errorf("test file requires at least one test")
	}
	return &tf, nil
}
