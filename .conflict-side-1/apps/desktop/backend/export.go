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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/exporter"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/version"
)

// Export formats supported by the dialog — identical fidelity to
// `reqly export postman | openapi | har | workspace`.
const (
	exportFormatPostman   = "postman"
	exportFormatOpenAPI   = "openapi"
	exportFormatHar       = "har"
	exportFormatWorkspace = "workspace"
)

// ExportRequest describes one export. Format selects the target; Collection
// narrows OpenAPI exports to a single collection (empty → whole workspace).
// OutName overrides the generated file or directory base name.
type ExportRequest struct {
	Format     string `json:"format"`
	Collection string `json:"collection,omitempty"`
	OutName    string `json:"outName,omitempty"`
}

// ExportResult reports where the export landed and how much it carried.
type ExportResult struct {
	Format       string `json:"format"`
	Path         string `json:"path"`
	RequestCount int    `json:"requestCount,omitempty"`
	EntryCount   int    `json:"entryCount,omitempty"`
}

// Export writes a shareable artifact derived from the open workspace into
// `<root>/.reqly/exports/` — outside Git, next to history.db. Postman and
// OpenAPI flatten every request exactly like the CLI; HAR synthesizes from
// history with secrets masked; workspace copies descriptors + request files
// via SaveWorkspace.
func (s *AppService) Export(req ExportRequest) (*ExportResult, error) {
	if s == nil || s.workspace == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to export")
	}
	format := strings.ToLower(strings.TrimSpace(req.Format))
	outDir := filepath.Join(s.root, ".reqly", "exports")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("create exports dir: %w", err)
	}

	switch format {
	case exportFormatPostman:
		ws, err := collections.LoadWorkspace(s.root)
		if err != nil {
			return nil, fmt.Errorf("load workspace: %w", err)
		}
		requests, count, err := flattenExportRequests(ws, "")
		if err != nil {
			return nil, err
		}
		name := exportTitle("", ws)
		data, err := exporter.ExportToPostmanJSON(name, requests)
		if err != nil {
			return nil, fmt.Errorf("export postman: %w", err)
		}
		path := filepath.Join(outDir, req.outName(defaultExportName(name), ".json"))
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return nil, err
		}
		return &ExportResult{Format: format, Path: path, RequestCount: count}, nil

	case exportFormatOpenAPI:
		ws, err := collections.LoadWorkspace(s.root)
		if err != nil {
			return nil, fmt.Errorf("load workspace: %w", err)
		}
		requests, count, err := flattenExportRequests(ws, req.Collection)
		if err != nil {
			return nil, err
		}
		title := ""
		baseURL := ""
		for _, c := range ws.Collections {
			if req.Collection != "" && c.Config.Name != req.Collection {
				continue
			}
			if title == "" && c.Config.Name != "" {
				title = c.Config.Name
			}
			if baseURL == "" && c.Config.BaseURL != "" {
				baseURL = c.Config.BaseURL
			}
		}
		name := exportTitle(title, ws)
		data, err := exporter.ExportOpenAPI(name, baseURL, requests)
		if err != nil {
			return nil, fmt.Errorf("export openapi: %w", err)
		}
		path := filepath.Join(outDir, req.outName(defaultExportName(name), ".yaml"))
		if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
			return nil, err
		}
		return &ExportResult{Format: format, Path: path, RequestCount: count}, nil

	case exportFormatHar:
		h := s.hist()
		if h == nil {
			return nil, fmt.Errorf("no history found: send a request before exporting HAR")
		}
		const limit = 500
		entries, err := h.List(context.Background(), limit, 0, nil)
		if err != nil {
			return nil, fmt.Errorf("load history: %w", err)
		}
		data, err := exporter.ExportHAR(entries, version.Version, nil)
		if err != nil {
			return nil, fmt.Errorf("export har: %w", err)
		}
		path := filepath.Join(outDir, req.outName("traffic", ".har"))
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return nil, err
		}
		return &ExportResult{Format: format, Path: path, EntryCount: len(entries)}, nil

	case exportFormatWorkspace:
		dest := filepath.Join(outDir, req.outName(filepath.Base(s.root)+"-export", ""))
		ws, err := collections.LoadWorkspace(s.root)
		if err != nil {
			return nil, fmt.Errorf("load workspace: %w", err)
		}
		retargetExportPaths(ws, s.root, dest)
		if err := collections.SaveWorkspace(dest, ws); err != nil {
			return nil, fmt.Errorf("export workspace: %w", err)
		}
		count := 0
		for _, c := range ws.Collections {
			count += len(c.Requests)
		}
		return &ExportResult{Format: format, Path: dest, RequestCount: count}, nil

	default:
		return nil, fmt.Errorf("unknown format %q: pick one of postman, openapi, har, workspace", req.Format)
	}
}

func (r ExportRequest) outName(base, ext string) string {
	name := strings.TrimSpace(r.OutName)
	if name == "" {
		name = base
	}
	return name + ext
}

// defaultExportName maps a human title to a filesystem-safe slug.
func defaultExportName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			return r
		default:
			return '-'
		}
	}, n)
	return strings.Trim(n, "-")
}

func exportTitle(collectionTitle string, ws *collections.Workspace) string {
	if collectionTitle != "" {
		return collectionTitle
	}
	if ws.Config.Name != "" {
		return ws.Config.Name
	}
	return "Reqly workspace"
}

// flattenExportRequests resolves every request under the workspace (or a
// single named collection), applying inherited auth/headers/base URL — the
// same walk `reqly export` performs in the CLI.
func flattenExportRequests(ws *collections.Workspace, collection string) ([]request.Request, int, error) {
	var requests []request.Request
	var walkFolders func(coll *collections.Collection, chain []*collections.Folder, folders []*collections.Folder)
	collect := func(coll *collections.Collection, chain []*collections.Folder, entry *collections.RequestEntry) {
		resolved, err := ws.ResolveRequest(coll, chain, entry)
		if err != nil {
			return
		}
		if resolved.Request.Name == "" {
			resolved.Request.Name = entry.Name
		}
		requests = append(requests, resolved.Request)
	}
	walkFolders = func(coll *collections.Collection, chain []*collections.Folder, folders []*collections.Folder) {
		for _, f := range folders {
			childChain := append(chain, f)
			for _, entry := range f.Requests {
				collect(coll, childChain, entry)
			}
			walkFolders(coll, childChain, f.Folders)
		}
	}
	for _, coll := range ws.Collections {
		if collection != "" && coll.Config.Name != collection {
			continue
		}
		for _, entry := range coll.Requests {
			collect(coll, nil, entry)
		}
		walkFolders(coll, nil, coll.Folders)
	}
	return requests, len(requests), nil
}

// retargetExportPaths rewrites every request entry's absolute source path to
// its destination twin so SaveWorkspace copies files into dest instead of
// rewriting the originals in place (RequestEntry.Path is absolute after load).
func retargetExportPaths(ws *collections.Workspace, srcRoot, dest string) {
	rewrite := func(entries []*collections.RequestEntry) {
		for _, e := range entries {
			if e.Path == "" || !filepath.IsAbs(e.Path) {
				continue
			}
			if rel, err := filepath.Rel(srcRoot, e.Path); err == nil && !strings.HasPrefix(rel, "..") {
				e.Path = filepath.Join(dest, rel)
			}
		}
	}
	var walkFolders func(folders []*collections.Folder)
	walkFolders = func(folders []*collections.Folder) {
		for _, f := range folders {
			rewrite(f.Requests)
			walkFolders(f.Folders)
		}
	}
	for _, coll := range ws.Collections {
		rewrite(coll.Requests)
		walkFolders(coll.Folders)
	}
}
