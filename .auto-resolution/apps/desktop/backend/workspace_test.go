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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceStatusReflectsLaunchRoot(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	status, err := svc.WorkspaceStatus()
	if err != nil {
		t.Fatalf("WorkspaceStatus: %v", err)
	}
	if !status.Found || status.Path != wsDir {
		t.Errorf("status = (%v, %q), want (true, %q)", status.Found, status.Path, wsDir)
	}
}

func TestWorkspaceStatusWithoutWorkspace(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	svc := NewAppService()
	status, err := svc.WorkspaceStatus()
	if err != nil {
		t.Fatalf("WorkspaceStatus: %v", err)
	}
	if status.Found {
		t.Errorf("Found = true with no workspace, want false")
	}
}

// newServiceInWorkspaceAndConfig starts the service inside an unrelated
// workspace with an injected (empty) config dir, ready to open others.
func newServiceInWorkspaceAndConfig(t *testing.T) (*AppService, string) {
	t.Helper()
	configDir := t.TempDir()
	svc, _ := newServiceInWorkspace(t)
	svc.configDir = func() (string, error) { return configDir, nil }
	return svc, configDir
}

func workspaceWithCollection(t *testing.T, parent, name string) string {
	t.Helper()
	ws := filepath.Join(parent, name)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws, "reqly.yaml"), []byte("name: "+name+"\n"), 0o644); err != nil {
		t.Fatalf("write reqly.yaml: %v", err)
	}
	coll := filepath.Join(ws, "collections", name+"-coll")
	if err := os.MkdirAll(coll, 0o755); err != nil {
		t.Fatalf("mkdir collection: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coll, "reqly.yaml"), []byte("name: "+name+"-coll\n"), 0o644); err != nil {
		t.Fatalf("write collection descriptor: %v", err)
	}
	return ws
}

func TestWorkspaceOpenRebuildsServicesAndPersists(t *testing.T) {
	parent := t.TempDir()
	target := workspaceWithCollection(t, parent, "other-ws")

	svc, configDir := newServiceInWorkspaceAndConfig(t)

	tree, err := svc.WorkspaceOpen(target)
	if err != nil {
		t.Fatalf("WorkspaceOpen: %v", err)
	}
	if tree.Name != "other-ws" {
		t.Errorf("tree name = %q, want other-ws", tree.Name)
	}
	status, _ := svc.WorkspaceStatus()
	if status.Path != target {
		t.Errorf("root = %q, want %q", status.Path, target)
	}
	stored, err := os.ReadFile(filepath.Join(configDir, "reqly", "desktop.json"))
	if err != nil || !strings.Contains(string(stored), target) {
		t.Errorf("last-workspace not persisted to injected config dir (%v): %s", err, stored)
	}
}

func TestWorkspaceOpenRejectsNonWorkspace(t *testing.T) {
	svc, _ := newServiceInWorkspaceAndConfig(t)
	bare := t.TempDir()
	_, err := svc.WorkspaceOpen(bare)
	if err == nil {
		t.Fatal("open on bare directory succeeded")
	}
	if !strings.Contains(err.Error(), "create") {
		t.Errorf("error = %v, want create guidance", err)
	}
}

func TestWorkspaceCreateScaffoldsAndOpens(t *testing.T) {
	parent := t.TempDir()
	fresh := filepath.Join(parent, "brand-new")
	if err := os.MkdirAll(fresh, 0o755); err != nil {
		t.Fatal(err)
	}

	svc, _ := newServiceInWorkspaceAndConfig(t)
	tree, err := svc.WorkspaceCreate(fresh, "")
	if err != nil {
		t.Fatalf("WorkspaceCreate: %v", err)
	}
	if tree.Name != "brand-new" {
		t.Errorf("tree name = %q, want folder-derived brand-new", tree.Name)
	}
	if _, err := os.Stat(filepath.Join(fresh, "reqly.yaml")); err != nil {
		t.Errorf("descriptor missing after scaffold: %v", err)
	}
	if _, err := os.Stat(filepath.Join(fresh, "collections")); err != nil {
		t.Errorf("collections/ missing after scaffold: %v", err)
	}
}

func TestWorkspaceCreateDefaultsNameToFolder(t *testing.T) {
	parent := t.TempDir()
	fresh := filepath.Join(parent, "my-api")
	if err := os.MkdirAll(fresh, 0o755); err != nil {
		t.Fatal(err)
	}
	svc, _ := newServiceInWorkspaceAndConfig(t)
	if _, err := svc.WorkspaceCreate(fresh, ""); err != nil {
		t.Fatalf("WorkspaceCreate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(fresh, "reqly.yaml"))
	if !strings.Contains(string(data), "my-api") {
		t.Errorf("descriptor = %q, want folder-derived name", data)
	}
}

func TestWorkspaceCreateNeverOverwritesDescriptor(t *testing.T) {
	existing := workspaceWithCollection(t, t.TempDir(), "existing")
	before, _ := os.ReadFile(filepath.Join(existing, "reqly.yaml"))

	svc, _ := newServiceInWorkspaceAndConfig(t)
	_, err := svc.WorkspaceCreate(existing, "replacement")
	if err == nil {
		t.Fatal("create over an existing workspace succeeded")
	}
	after, _ := os.ReadFile(filepath.Join(existing, "reqly.yaml"))
	if string(before) != string(after) {
		t.Errorf("descriptor was modified:\nbefore=%q\nafter=%q", before, after)
	}
}

func TestRestoreLastReopensPersistedWorkspace(t *testing.T) {
	configDir := t.TempDir()
	parent := t.TempDir()
	stored := workspaceWithCollection(t, parent, "stored-ws")

	if err := os.MkdirAll(filepath.Join(configDir, "reqly"), 0o755); err != nil {
		t.Fatal(err)
	}
	persist := filepath.Join(configDir, "reqly", "desktop.json")
	if err := os.WriteFile(persist, []byte(`{"lastWorkspace":"`+stored+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	elsewhere := t.TempDir()
	t.Chdir(elsewhere)
	svc := NewAppService()
	svc.configDir = func() (string, error) { return configDir, nil }

	status, err := svc.WorkspaceRestoreLast()
	if err != nil {
		t.Fatalf("WorkspaceRestoreLast: %v", err)
	}
	if !status.Found || status.Path != stored {
		t.Errorf("restored root = (%v, %q), want persisted %q", status.Found, status.Path, stored)
	}
}

func TestRestoreLastIgnoresInvalidStoredPath(t *testing.T) {
	configDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(configDir, "reqly"), 0o755); err != nil {
		t.Fatal(err)
	}
	gone := filepath.Join(t.TempDir(), "deleted")
	persist := filepath.Join(configDir, "reqly", "desktop.json")
	if err := os.WriteFile(persist, []byte(`{"lastWorkspace":"`+gone+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	svc, wsDir := newServiceInWorkspace(t)
	svc.configDir = func() (string, error) { return configDir, nil }
	status, err := svc.WorkspaceRestoreLast()
	if err != nil {
		t.Fatalf("WorkspaceRestoreLast: %v", err)
	}
	if !status.Found || status.Path != wsDir {
		t.Errorf("root = (%v, %q), want untouched CWD workspace %q", status.Found, status.Path, wsDir)
	}
}
