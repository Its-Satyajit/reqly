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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// WorkspaceStatusResponse tells the frontend whether a workspace is open and
// where it lives, so the UI can branch between normal operation and the
// empty-state bootstrap flow at startup.
type WorkspaceStatusResponse struct {
	Found bool   `json:"found"`
	Path  string `json:"path,omitempty"`
}

// pickDir opens a native directory chooser; injectable so tests never touch
// a real dialog. Nil means "use the Wails dialog".
func (s *AppService) pickDirectory() (string, error) {
	if s.pickDir != nil {
		return s.pickDir()
	}
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("picker unavailable: application is not running")
	}
	return app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		CanChooseDirectories: true,
		CanChooseFiles:       false,
		CanCreateDirectories: true,
	}).PromptForSingleSelection()
}

// configDirPath returns the directory that holds Reqly's desktop preferences;
// injectable for tests. Nil means os.UserConfigDir.
func (s *AppService) configDirPath() (string, error) {
	if s.configDir != nil {
		return s.configDir()
	}
	return os.UserConfigDir()
}

// lastWorkspaceFile is the persisted-preferences file inside configDirPath().
var lastWorkspaceFile = filepath.Join("reqly", "desktop.json")

// WorkspaceStatus reports whether a workspace is currently open.
func (s *AppService) WorkspaceStatus() (*WorkspaceStatusResponse, error) {
	return &WorkspaceStatusResponse{Found: s.root != "", Path: s.root}, nil
}

// WorkspacePickFolder shows the native directory chooser and returns the raw
// selection ("" when cancelled). The frontend decides what to do with it —
// open an existing workspace, or offer to scaffold one.
func (s *AppService) WorkspacePickFolder() (string, error) {
	dir, err := s.pickDirectory()
	if err != nil {
		return "", fmt.Errorf("folder picker failed: %w", err)
	}
	return dir, nil
}

// WorkspaceOpen switches every service to the workspace rooted at dir and
// returns its collection tree. Non-workspace directories are rejected with
// create guidance rather than silently scaffolded.
func (s *AppService) WorkspaceOpen(dir string) (*core.WorkspaceTree, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", dir, err)
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%q is not a folder", abs)
	}
	if !collections.IsWorkspace(abs) {
		return nil, fmt.Errorf("%q is not a Reqly workspace yet: create one to start using it here", filepath.Base(abs))
	}
	// Ensure standard workspace folders exist so features (collections, environments, tests, .reqly) operate smoothly
	for _, folder := range []string{"collections", "environments", "tests", ".reqly"} {
		_ = os.MkdirAll(filepath.Join(abs, folder), 0o755)
	}
	s.rebuildServices(abs)
	if err := s.persistLastWorkspace(abs); err != nil {
		return nil, fmt.Errorf("save last workspace: %w", err)
	}
	return s.workspace.Load()
}

// WorkspaceCreate scaffolds a minimal Git-native workspace (descriptor plus
// standard directories: collections/, environments/, tests/, .reqly/) at dir —
// deriving the name from the folder when empty — then opens it. An existing descriptor is never touched.
func (s *AppService) WorkspaceCreate(dir string, name string) (*core.WorkspaceTree, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", dir, err)
	}
	if collections.IsWorkspace(abs) {
		return nil, fmt.Errorf("%q already contains a Reqly workspace: open it instead", filepath.Base(abs))
	}
	for _, folder := range []string{"collections", "environments", "tests", ".reqly"} {
		if err := os.MkdirAll(filepath.Join(abs, folder), 0o755); err != nil {
			return nil, fmt.Errorf("create %s folder: %w", folder, err)
		}
	}
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(abs)
	}
	descriptor := fmt.Sprintf("name: %s\n", name)
	if err := os.WriteFile(filepath.Join(abs, "reqly.yaml"), []byte(descriptor), 0o644); err != nil {
		return nil, fmt.Errorf("write reqly.yaml: %w", err)
	}
	return s.WorkspaceOpen(abs)
}

// WorkspaceRestoreLast reopens the persisted last workspace when it is still
// valid, so a normally-launched app lands back where the user left off. A
// missing or invalid stored path is a silent no-op returning current status.
func (s *AppService) WorkspaceRestoreLast() (*WorkspaceStatusResponse, error) {
	stored, err := s.readLastWorkspace()
	if err != nil || stored == "" || !collections.IsWorkspace(stored) || stored == s.root {
		return s.WorkspaceStatus()
	}
	tree, err := s.WorkspaceOpen(stored)
	if err != nil {
		return s.WorkspaceStatus()
	}
	_ = tree
	return &WorkspaceStatusResponse{Found: true, Path: stored}, nil
}

// rebuildServices re-wires every root-scoped service after a workspace
// switch. This must construct exactly what NewAppService does — the services
// are thin wrappers around the resolved root (ADR 0025 keeps business logic
// in internal/core).
func (s *AppService) rebuildServices(root string) {
	opened := secrets.OpenForWorkspace(root, "keychain")
	s.requests = core.NewRunServiceWithTokenStore(root, opened.Store)
	s.authBackend = opened.Backend
	s.environments = core.NewEnvironmentServiceWithStore(root, opened.Store)
	s.workspace = core.NewWorkspaceService(root)
	s.runs = core.NewCollectionRunService(root)
	s.root = root
	if opened.Store != nil {
		s.auth = core.NewAuthService(opened.Store, root)
	} else {
		s.auth = nil
	}
}

// persistLastWorkspace records dir as the workspace to restore on next launch.
func (s *AppService) persistLastWorkspace(dir string) error {
	configDir, err := s.configDirPath()
	if err != nil {
		return err
	}
	target := filepath.Join(configDir, lastWorkspaceFile)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{"lastWorkspace": dir})
	if err != nil {
		return err
	}
	return os.WriteFile(target, payload, 0o600)
}

// readLastWorkspace returns the persisted path, or "" when absent/unreadable.
func (s *AppService) readLastWorkspace() (string, error) {
	configDir, err := s.configDirPath()
	if err != nil {
		return "", err
	}
	payload, err := os.ReadFile(filepath.Join(configDir, lastWorkspaceFile))
	if err != nil {
		return "", err
	}
	var prefs struct {
		LastWorkspace string `json:"lastWorkspace"`
	}
	if err := json.Unmarshal(payload, &prefs); err != nil {
		return "", err
	}
	return strings.TrimSpace(prefs.LastWorkspace), nil
}
