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

	"github.com/Its-Satyajit/reqly/internal/git"
)

// GitStatus reports branch and per-file status for the workspace repository.
func (s *AppService) GitStatus() (*git.StatusResult, error) {
	if s == nil || s.root == "" {
		return &git.StatusResult{RepoFound: false}, nil
	}
	return git.Status(s.root)
}

// GitStage stages the given workspace-relative paths.
func (s *AppService) GitStage(paths []string) error {
	if s == nil || s.root == "" {
		return fmt.Errorf("no workspace found: open a reqly workspace to manage git")
	}
	return git.Stage(s.root, paths)
}

// GitCommit records staged changes with the given message.
func (s *AppService) GitCommit(message string) error {
	if s == nil || s.root == "" {
		return fmt.Errorf("no workspace found: open a reqly workspace to commit")
	}
	return git.Commit(s.root, message)
}

// GitUnstage removes the given workspace-relative paths from the index.
func (s *AppService) GitUnstage(paths []string) error {
	if s == nil || s.root == "" {
		return fmt.Errorf("no workspace found: open a reqly workspace to manage git")
	}
	return git.Unstage(s.root, paths)
}
