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

package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeCollectionWorkspace builds a temp-dir workspace with one collection
// (users) holding two requests and a nested folder (auth) holding one
// request, mirroring the internal/collections fixtures.
func writeCollectionWorkspace(t *testing.T, root string) {
	t.Helper()
	files := map[string]string{
		"reqly.yaml":                                 `name: demo`,
		"collections/users/reqly.yaml":               `name: users`,
		"collections/users/list-users.yaml":          "name: List Users\nrequest: {method: GET, url: users}",
		"collections/users/get-user.yaml":            "request: {method: GET, url: users/1}",
		"collections/users/auth/reqly.yaml":          `name: auth`,
		"collections/users/auth/login.yaml":          "request: {method: POST, url: auth/login}",
		"collections/users/not-a-container/":         "",
		"collections/users/not-a-container/req.yaml": "request: {method: GET, url: hidden}",
	}
	for name, contents := range files {
		path := filepath.Join(root, name)
		if strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWorkspaceServiceLoadReturnsTree(t *testing.T) {
	dir := t.TempDir()
	writeCollectionWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	tree, err := svc.Load()
	if err != nil {
		t.Fatal(err)
	}

	if tree.Name != "demo" {
		t.Fatalf("name = %q, want demo", tree.Name)
	}
	if tree.Path != dir {
		t.Fatalf("path = %q, want %q", tree.Path, dir)
	}
	if len(tree.Collections) != 1 {
		t.Fatalf("collections = %d, want 1", len(tree.Collections))
	}

	coll := tree.Collections[0]
	if coll.Name != "users" || coll.Path != "users" {
		t.Fatalf("collection = %+v", coll)
	}
	if len(coll.Requests) != 2 {
		t.Fatalf("requests = %d, want 2", len(coll.Requests))
	}
	if coll.Requests[0].Name != "get-user" || coll.Requests[0].Path != "users/get-user" {
		t.Fatalf("requests[0] = %+v", coll.Requests[0])
	}
	if coll.Requests[1].Name != "list-users" || coll.Requests[1].Path != "users/list-users" {
		t.Fatalf("requests[1] = %+v", coll.Requests[1])
	}

	if len(coll.Folders) != 1 {
		t.Fatalf("folders = %d, want 1", len(coll.Folders))
	}
	folder := coll.Folders[0]
	if folder.Name != "auth" || folder.Path != "users/auth" {
		t.Fatalf("folder = %+v", folder)
	}
	if len(folder.Requests) != 1 || folder.Requests[0].Path != "users/auth/login" {
		t.Fatalf("folder requests = %+v", folder.Requests)
	}
	if len(folder.Folders) != 0 {
		t.Fatalf("nested folder folders = %d, want 0", len(folder.Folders))
	}
	for _, req := range coll.Requests {
		if strings.Contains(req.Path, "not-a-container") {
			t.Fatalf("descriptor-less dir leaked into tree: %+v", tree)
		}
	}
}

func TestWorkspaceServiceLoadEmptyWorkspaceIsEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("name: empty\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewWorkspaceService(dir)
	tree, err := svc.Load()
	if err != nil {
		t.Fatalf("expected no error for empty workspace, got %v", err)
	}
	if len(tree.Collections) != 0 {
		t.Fatalf("collections = %d, want 0", len(tree.Collections))
	}
}

func TestWorkspaceServiceLoadWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml

	svc := NewWorkspaceService(dir)
	if _, err := svc.Load(); err == nil {
		t.Fatal("expected error without a workspace, got nil")
	}
}

func TestWorkspaceServiceLoadNameFallsBackToDirectoryName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-api")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("baseURL: https://example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewWorkspaceService(dir)
	tree, err := svc.Load()
	if err != nil {
		t.Fatal(err)
	}
	if tree.Name != "my-api" {
		t.Fatalf("name = %q, want my-api (basename fallback)", tree.Name)
	}
}

func TestWorkspaceServiceLoadSkipsDescriptorlessSubdirectories(t *testing.T) {
	dir := t.TempDir()
	writeCollectionWorkspace(t, dir)

	// not-a-container has no reqly.yaml, so it is not a folder: the tree must
	// not surface it.
	svc := NewWorkspaceService(dir)
	tree, err := svc.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Collections) != 1 {
		t.Fatalf("collections = %d, want 1 (descriptor-less dir skipped)", len(tree.Collections))
	}
	for _, req := range tree.Collections[0].Requests {
		if strings.Contains(req.Path, "not-a-container") {
			t.Fatalf("descriptor-less dir leaked into tree: %+v", tree)
		}
	}
}
