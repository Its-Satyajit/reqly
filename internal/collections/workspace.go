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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// Workspace is the top-level container for an API project, mirrored to a
// directory of Git-native files.
//
// Layout:
//
//	<root>/
//	├── reqly.yaml              # workspace descriptor
//	└── collections/
//	    ├── <collection>/
//	    │   ├── reqly.yaml      # collection descriptor
//	    │   ├── <request>.yaml  # request file (requestfile format)
//	    │   └── <folder>/
//	    │       ├── reqly.yaml  # folder descriptor
//	    │       └── <request>.yaml
type Workspace struct {
	// Root is the workspace directory.
	Root string
	// Config is the workspace-level descriptor.
	Config Config
	// Collections in discovery order.
	Collections []*Collection
}

// Collection groups related requests under the workspace.
type Collection struct {
	// Name is the collection directory name.
	Name string
	// Dir is the collection directory.
	Dir string
	// Config is the collection descriptor.
	Config Config
	// Folders in discovery order.
	Folders []*Folder
	// Requests in discovery order.
	Requests []*RequestEntry
}

// Folder is a nested container within a collection or another folder.
type Folder struct {
	// Name is the folder directory name.
	Name string
	// Dir is the folder directory.
	Dir string
	// Config is the folder descriptor.
	Config Config
	// Folders in discovery order.
	Folders []*Folder
	// Requests in discovery order.
	Requests []*RequestEntry
}

// RequestEntry is a request file located in a collection or folder.
type RequestEntry struct {
	// Name is the file name without extension.
	Name string
	// Path is the absolute path to the request file.
	Path string
	// File is the parsed request file.
	File *requestfile.File
}

// LoadWorkspace reads a workspace rooted at dir. dir must contain a
// reqly.yaml descriptor; collections live in dir/collections.
func LoadWorkspace(dir string) (*Workspace, error) {
	cfg, ok, err := loadConfig(dir)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not a workspace: %s has no %s", dir, configFileName)
	}

	w := &Workspace{Root: dir, Config: cfg}

	collectionsDir := filepath.Join(dir, "collections")
	entries, err := os.ReadDir(collectionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return w, nil
		}
		return nil, fmt.Errorf("read collections dir: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		collDir := filepath.Join(collectionsDir, e.Name())
		coll, err := loadCollection(collDir)
		if err != nil {
			return nil, err
		}
		w.Collections = append(w.Collections, coll)
	}

	sort.Slice(w.Collections, func(i, j int) bool { return w.Collections[i].Name < w.Collections[j].Name })
	return w, nil
}

// loadCollection reads a collection directory: its descriptor, request files,
// and nested folders.
func loadCollection(dir string) (*Collection, error) {
	cfg, ok, err := loadConfig(dir)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("collection %s has no %s", dir, configFileName)
	}

	coll := &Collection{
		Name:   filepath.Base(dir),
		Dir:    dir,
		Config: cfg,
	}

	if err := loadContainer(&coll.Folders, &coll.Requests, dir, 1); err != nil {
		return nil, err
	}
	return coll, nil
}

// loadContainer fills folders and requests for a collection or folder dir.
// maxDepth bounds folder nesting to avoid cycles or runaway recursion.
func loadContainer(folders *[]*Folder, requests *[]*RequestEntry, dir string, depth int) error {
	const maxDepth = 16
	if depth > maxDepth {
		return fmt.Errorf("folder nesting too deep under %s (max %d)", dir, maxDepth)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}

	var subdirs []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			subdirs = append(subdirs, e)
			continue
		}
		name := e.Name()
		if name == configFileName || !isRequestFile(name) {
			continue
		}
		path := filepath.Join(dir, name)
		f, err := requestfile.LoadFile(path)
		if err != nil {
			return err
		}
		*requests = append(*requests, &RequestEntry{
			Name: strings.TrimSuffix(name, filepath.Ext(name)),
			Path: path,
			File: f,
		})
	}

	for _, d := range subdirs {
		folderDir := filepath.Join(dir, d.Name())
		folderCfg, ok, err := loadConfig(folderDir)
		if err != nil {
			return err
		}
		if !ok {
			// A subdirectory without a descriptor is not part of the workspace.
			continue
		}
		folder := &Folder{
			Name:   d.Name(),
			Dir:    folderDir,
			Config: folderCfg,
		}
		if err := loadContainer(&folder.Folders, &folder.Requests, folderDir, depth+1); err != nil {
			return err
		}
		*folders = append(*folders, folder)
	}

	sort.Slice(*folders, func(i, j int) bool { return (*folders)[i].Name < (*folders)[j].Name })
	sort.Slice(*requests, func(i, j int) bool { return (*requests)[i].Name < (*requests)[j].Name })
	return nil
}

// isRequestFile reports whether name looks like a request file (JSON or YAML).
func isRequestFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json", ".yaml", ".yml":
		return true
	}
	return false
}

// RequestPath identifies a request within the workspace by relative path,
// e.g. "users/list-users.yaml" or "users/auth/login.json". Folders are
// separated by "/".
type RequestPath string

// FindRequest locates a request by its workspace-relative path. It returns the
// owning collection, the folder ancestor chain (outermost first), the request
// entry, and true when found.
func (w *Workspace) FindRequest(path RequestPath) (*Collection, []*Folder, *RequestEntry, bool) {
	parts := splitPath(string(path))
	if len(parts) == 0 {
		return nil, nil, nil, false
	}

	collName := parts[0]
	var coll *Collection
	for _, c := range w.Collections {
		if c.Name == collName {
			coll = c
			break
		}
	}
	if coll == nil {
		return nil, nil, nil, false
	}

	rest := parts[1:]
	if len(rest) == 0 {
		return nil, nil, nil, false
	}

	// The last part is the request file; everything between is folders.
	var chain []*Folder
	current := coll.Folders
	for i := 0; i < len(rest)-1; i++ {
		name := rest[i]
		var next *Folder
		for _, f := range current {
			if f.Name == name {
				next = f
				break
			}
		}
		if next == nil {
			return nil, nil, nil, false
		}
		chain = append(chain, next)
		current = next.Folders
	}

	fileName := rest[len(rest)-1]
	if idx := strings.LastIndexByte(fileName, '.'); idx >= 0 {
		fileName = fileName[:idx]
	}

	if len(chain) == 0 {
		for _, r := range coll.Requests {
			if r.Name == fileName {
				return coll, nil, r, true
			}
		}
		return nil, nil, nil, false
	}

	leaf := chain[len(chain)-1]
	for _, r := range leaf.Requests {
		if r.Name == fileName {
			return coll, chain, r, true
		}
	}
	return nil, nil, nil, false
}

// splitPath splits a relative path on "/", collapsing empty segments.
func splitPath(path string) []string {
	var parts []string
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
