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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func seedDocsWorkspace(t *testing.T, wsDir string) {
	t.Helper()
	for _, coll := range []string{"users", "billing"} {
		collDir := filepath.Join(wsDir, "collections", coll)
		if err := os.MkdirAll(collDir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		desc := "name: " + coll + "\nbaseUrl: https://api.example.com\n"
		if err := os.WriteFile(filepath.Join(collDir, "reqly.yaml"), []byte(desc), 0o644); err != nil {
			t.Fatalf("write descriptor: %v", err)
		}
		reqFile, err := json.Marshal(map[string]any{
			"version": "1",
			"request": map[string]any{
				"name":   "list",
				"method": "GET",
				"url":    "https://api.example.com/" + coll,
			},
		})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if err := os.WriteFile(filepath.Join(collDir, "list.json"), reqFile, 0o644); err != nil {
			t.Fatalf("write request: %v", err)
		}
	}
}

func TestDocsGenerateRequiresWorkspace(t *testing.T) {
	svc := &AppService{}
	if _, err := svc.DocsGenerate(DocsGenerateRequest{}); err == nil {
		t.Fatal("expected error without a workspace, got nil")
	}
}

func TestDocsGenerateWholeWorkspace(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedDocsWorkspace(t, wsDir)

	res, err := svc.DocsGenerate(DocsGenerateRequest{OutName: "site"})
	if err != nil {
		t.Fatalf("DocsGenerate: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(res.Path), ".reqly/docs/site") {
		t.Errorf("Path = %q", res.Path)
	}
	if res.RequestCount != 2 {
		t.Errorf("RequestCount = %d, want 2", res.RequestCount)
	}
	var names []string
	for _, f := range res.Files {
		names = append(names, f.Name)
	}
	if len(names) != 3 || names[0] != "billing.md" || names[1] != "index.md" || names[2] != "users.md" {
		t.Errorf("file list = %v", names)
	}
	idx := res.Files[1].Content
	if !strings.Contains(idx, "- [users](users.md) (1 requests)") ||
		!strings.Contains(idx, "- [billing](billing.md) (1 requests)") {
		t.Errorf("index content = %q", idx)
	}
	if _, err := os.Stat(filepath.Join(res.Path, "index.md")); err != nil {
		t.Errorf("docs not on disk: %v", err)
	}
}

func TestDocsGenerateCollectionFilter(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedDocsWorkspace(t, wsDir)

	res, err := svc.DocsGenerate(DocsGenerateRequest{Collections: []string{"users"}})
	if err != nil {
		t.Fatalf("DocsGenerate filtered: %v", err)
	}
	if len(res.Files) != 2 { // index.md + users.md
		t.Fatalf("file count = %d, want 2 (%+v)", len(res.Files), res.Files)
	}
	if strings.Contains(res.Files[0].Content, "billing") && res.Files[0].Name == "index.md" {
		t.Errorf("index still lists billing: %q", res.Files[0].Content)
	}
	if res.RequestCount != 1 {
		t.Errorf("RequestCount = %d, want 1", res.RequestCount)
	}
}

func TestDocsGenerateRejectsUnknownCollection(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedDocsWorkspace(t, wsDir)

	if _, err := svc.DocsGenerate(DocsGenerateRequest{Collections: []string{"nope"}}); err == nil {
		t.Fatal("expected error for unknown collection, got nil")
	}
}

func TestDocsGenerateRejectsPathyOutName(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedDocsWorkspace(t, wsDir)

	if _, err := svc.DocsGenerate(DocsGenerateRequest{OutName: "../escape"}); err == nil {
		t.Fatal("expected error for path traversal outName, got nil")
	}
}
