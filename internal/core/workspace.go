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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/collections"
)

// WorkspaceRequest is a request file within a collection or folder, located
// by its workspace-relative Request Path (e.g. "users/auth/login").
type WorkspaceRequest struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// WorkspaceFolder is a nested container (recursively) within a collection.
type WorkspaceFolder struct {
	Name     string             `json:"name"`
	Path     string             `json:"path"`
	Folders  []WorkspaceFolder  `json:"folders"`
	Requests []WorkspaceRequest `json:"requests"`
}

// WorkspaceCollection is a top-level collection of folders and requests.
type WorkspaceCollection struct {
	Name     string             `json:"name"`
	Path     string             `json:"path"`
	Folders  []WorkspaceFolder  `json:"folders"`
	Requests []WorkspaceRequest `json:"requests"`
}

// WorkspaceTree is the bridge-friendly view of a workspace's collection
// hierarchy: collections → folders → requests, all name-sorted.
type WorkspaceTree struct {
	Name        string                `json:"name"`
	Path        string                `json:"path"`
	Collections []WorkspaceCollection `json:"collections"`
}

// WorkspaceService exposes the workspace's collection tree to front-ends on
// the same seam the CLI uses: internal/collections. It is UI-agnostic and
// read-only — request files are never written through it.
type WorkspaceService struct {
	root string
}

// NewWorkspaceService returns a WorkspaceService rooted at the given workspace
// root ("" means no workspace; Load then errors).
func NewWorkspaceService(root string) *WorkspaceService {
	return &WorkspaceService{root: root}
}

// Load returns the workspace's collection tree (collections, folders, and
// request files, all name-sorted) with workspace-relative Request Paths. A
// workspace without a collections/ directory yields an empty tree, not an
// error; a missing descriptor is an error.
func (s *WorkspaceService) Load() (*WorkspaceTree, error) {
	if s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to browse collections")
	}
	ws, err := collections.LoadWorkspace(s.root)
	if err != nil {
		return nil, err
	}
	return workspaceTreeDTO(ws), nil
}

// workspaceTreeDTO maps an internal workspace to its bridge-friendly tree.
// Request Paths are workspace-relative (collections/ prefix and file
// extension stripped), matching the identity used by FindRequest.
func workspaceTreeDTO(ws *collections.Workspace) *WorkspaceTree {
	name := ws.Config.Name
	if name == "" {
		name = filepath.Base(ws.Root)
	}
	tree := &WorkspaceTree{
		Name:        name,
		Path:        ws.Root,
		Collections: make([]WorkspaceCollection, 0, len(ws.Collections)),
	}
	for _, coll := range ws.Collections {
		tree.Collections = append(tree.Collections, WorkspaceCollection{
			Name:     coll.Name,
			Path:     containerPath(ws.Root, coll.Dir),
			Folders:  folderDTOs(ws.Root, coll.Folders),
			Requests: requestDTOs(ws.Root, coll.Requests),
		})
	}
	return tree
}

func folderDTOs(root string, folders []*collections.Folder) []WorkspaceFolder {
	out := make([]WorkspaceFolder, 0, len(folders))
	for _, f := range folders {
		out = append(out, WorkspaceFolder{
			Name:     f.Name,
			Path:     containerPath(root, f.Dir),
			Folders:  folderDTOs(root, f.Folders),
			Requests: requestDTOs(root, f.Requests),
		})
	}
	return out
}

func requestDTOs(root string, requests []*collections.RequestEntry) []WorkspaceRequest {
	out := make([]WorkspaceRequest, 0, len(requests))
	for _, r := range requests {
		out = append(out, WorkspaceRequest{Name: r.Name, Path: containerPath(root, r.Path)})
	}
	return out
}

// containerPath returns the workspace-relative path of a container directory
// or request file, with the leading "collections/" segment and any file
// extension stripped ("<root>/collections/users/auth/login.yaml" →
// "users/auth/login").
func containerPath(root, abs string) string {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return strings.TrimPrefix(abs, root)
	}
	rel = filepath.ToSlash(rel)
	rel = strings.TrimPrefix(rel, "collections/")
	if rel == "collections" {
		return ""
	}
	ext := filepath.Ext(rel)
	if ext != "" {
		rel = strings.TrimSuffix(rel, ext)
	}
	return rel
}
