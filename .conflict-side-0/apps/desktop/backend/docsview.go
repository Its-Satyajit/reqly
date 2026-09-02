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
	"sort"
	"strings"
	"time"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/docs"
)

// DocsGenerateRequest describes one docs generation run. Collections narrows
// output to the named collections (empty → whole workspace). OutName
// overrides the default directory base name.
type DocsGenerateRequest struct {
	Collections []string `json:"collections,omitempty"`
	OutName     string   `json:"outName,omitempty"`
}

// DocsFile is one generated Markdown file returned for preview.
type DocsFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// DocsResult reports where the docs landed plus their content so the panel
// can preview and copy without a second file-read round trip.
type DocsResult struct {
	Path         string     `json:"path"`
	RequestCount int        `json:"requestCount"`
	Files        []DocsFile `json:"files"`
}

// DocsGenerate renders Markdown documentation (index.md + one file per
// collection) from the open workspace into `<root>/.reqly/docs/<name>/` —
// outside Git, like exports — and returns every file for in-app preview.
func (s *AppService) DocsGenerate(req DocsGenerateRequest) (*DocsResult, error) {
	if s == nil || s.workspace == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to generate docs")
	}
	ws, err := collections.LoadWorkspace(s.root)
	if err != nil {
		return nil, fmt.Errorf("load workspace: %w", err)
	}
	if len(req.Collections) > 0 {
		want := make(map[string]bool, len(req.Collections))
		for _, c := range req.Collections {
			want[strings.TrimSpace(c)] = true
		}
		filtered := make([]*collections.Collection, 0, len(ws.Collections))
		for _, c := range ws.Collections {
			if want[c.Config.Name] {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) == 0 {
			return nil, fmt.Errorf("no matching collections: %v", req.Collections)
		}
		out := *ws
		out.Collections = filtered
		ws = &out
	}

	name := strings.TrimSpace(req.OutName)
	if name == "" {
		name = "docs-" + time.Now().Format("20060102-150405")
	}
	if strings.ContainsAny(name, "/\\") {
		return nil, fmt.Errorf("outName must be a plain directory name")
	}
	outDir := filepath.Join(s.root, ".reqly", "docs", name)
	if err := docs.Generate(outDir, ws, ""); err != nil {
		return nil, fmt.Errorf("generate docs: %w", err)
	}

	count := 0
	for _, c := range ws.Collections {
		count += len(c.Requests) + countWorkspaceFolderRequests(c.Folders)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		return nil, fmt.Errorf("read docs dir: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	files := make([]DocsFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(outDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		files = append(files, DocsFile{Name: e.Name(), Content: string(data)})
	}
	return &DocsResult{Path: outDir, RequestCount: count, Files: files}, nil
}

// countWorkspaceFolderRequests counts requests under nested folders — same
// walk internal/docs uses for its index tally.
func countWorkspaceFolderRequests(folders []*collections.Folder) int {
	n := 0
	for _, f := range folders {
		n += len(f.Requests)
		n += countWorkspaceFolderRequests(f.Folders)
	}
	return n
}
