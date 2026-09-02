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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Its-Satyajit/reqly/internal/validation"
)

// SpecRead returns the raw content of a spec file at a workspace-relative or absolute path.
func (s *AppService) SpecRead(path string) (string, error) {
	abs, err := s.resolveTestPath(path)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("read spec %q: %w", path, err)
	}
	return string(b), nil
}

// SpecSave writes content to a spec file atomically with 0644 permissions.
func (s *AppService) SpecSave(path string, content string) error {
	abs, err := s.resolveTestPath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("create spec dir: %w", err)
	}
	tmp := abs + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write spec %q: %w", path, err)
	}
	if err := os.Rename(tmp, abs); err != nil {
		return fmt.Errorf("save spec %q: %w", path, err)
	}
	return nil
}

// ValidateProject validates a workspace project at path (workspace-relative or absolute).
func (s *AppService) ValidateProject(path string) (*validation.Result, error) {
	target := path
	if target == "" {
		if s.root != "" {
			target = s.root
		} else {
			target = "."
		}
	} else {
		abs, err := s.resolveTestPath(path)
		if err == nil {
			target = abs
		}
	}
	return validation.ValidateProject(target)
}
