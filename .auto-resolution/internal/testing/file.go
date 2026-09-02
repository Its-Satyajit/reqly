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

package testing

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// TestFile couples a request definition with the assertions that run against
// its response. It is the on-disk format shared by the CLI, Desktop, and MCP,
// accepted in JSON or YAML.
type TestFile struct {
	Name      string            `json:"name" yaml:"name"`
	Request   request.Request   `json:"request" yaml:"request"`
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Tests     []Test            `json:"tests" yaml:"tests"`
	// Environment selects the environment to apply to this test. It is
	// overridden by the --env flag and REQLY_ENV at runtime.
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// Suite converts a TestFile into a Suite for evaluation.
func (f TestFile) Suite() Suite {
	return Suite{Name: f.Name, Tests: f.Tests}
}

// VariablesSet returns the file's variables as a variables.Set using the
// request scope, ready for interpolation during execution.
func (f TestFile) VariablesSet() *variables.Set {
	set := variables.NewSet()
	for key, value := range f.Variables {
		set.Set(variables.ScopeRequest, key, value)
	}
	return set
}

// LoadTestFile reads and parses a test file from disk.
func LoadTestFile(path string) (*TestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read test file %q: %w", path, err)
	}
	return ParseTestFile(data)
}

// ParseTestFile parses test file contents in JSON or YAML format.
func ParseTestFile(data []byte) (*TestFile, error) {
	var tf TestFile
	if err := json.Unmarshal(data, &tf); err == nil {
		return validateTestFile(&tf)
	}
	var y TestFile
	if err := yaml.Unmarshal(data, &y); err != nil {
		return nil, fmt.Errorf("parse test file: %w", err)
	}
	return validateTestFile(&y)
}

func validateTestFile(tf *TestFile) (*TestFile, error) {
	if tf.Request.URL == "" {
		return nil, fmt.Errorf("test file requires a request.url")
	}
	if len(tf.Tests) == 0 {
		return nil, fmt.Errorf("test file requires at least one test")
	}
	return tf, nil
}
