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

	"gopkg.in/yaml.v3"
)

// seedExportWorkspace writes a collection with one request into the temp
// workspace so Postman/OpenAPI/workspace exports have something to flatten.
func seedExportWorkspace(t *testing.T, wsDir string) {
	t.Helper()
	collDir := filepath.Join(wsDir, "collections", "users")
	if err := os.MkdirAll(collDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	desc := "name: users\nbaseUrl: https://api.example.com\n"
	if err := os.WriteFile(filepath.Join(collDir, "reqly.yaml"), []byte(desc), 0o644); err != nil {
		t.Fatalf("write descriptor: %v", err)
	}
	reqFile, err := json.Marshal(map[string]any{
		"version": "1",
		"request": map[string]any{
			"name":   "list",
			"method": "GET",
			"url":    "https://api.example.com/users",
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(collDir, "list.json"), reqFile, 0o644); err != nil {
		t.Fatalf("write request: %v", err)
	}
}

func TestExportRequiresWorkspace(t *testing.T) {
	svc := &AppService{}
	if _, err := svc.Export(ExportRequest{Format: "postman"}); err == nil {
		t.Fatal("Export without a workspace succeeded, want error")
	}
}

func TestExportUnknownFormatErrors(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)
	_, err := svc.Export(ExportRequest{Format: "soap"})
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Fatalf("err = %v, want unknown-format error", err)
	}
}

func TestExportPostmanWritesJSONUnderExports(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)

	res, err := svc.Export(ExportRequest{Format: "postman"})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if res.RequestCount != 1 {
		t.Errorf("RequestCount = %d, want 1", res.RequestCount)
	}
	data, err := os.ReadFile(res.Path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	var coll struct {
		Info struct {
			Name string `json:"name"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &coll); err != nil {
		t.Fatalf("postman JSON invalid: %v", err)
	}
	if !strings.Contains(filepath.Dir(res.Path), filepath.Join(".reqly", "exports")) {
		t.Errorf("Path = %q, want it under .reqly/exports", res.Path)
	}
}

func TestExportOpenAPIScopedToCollection(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)

	res, err := svc.Export(ExportRequest{Format: "openapi", Collection: "users"})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if res.RequestCount != 1 {
		t.Errorf("RequestCount = %d, want 1", res.RequestCount)
	}
	data, err := os.ReadFile(res.Path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	var spec map[string]any
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatalf("openapi YAML invalid: %v", err)
	}
	if spec["openapi"] == nil || spec["paths"] == nil {
		t.Errorf("spec missing openapi/paths keys: %v", spec)
	}
}

func TestExportHarEmptyHistorySucceeds(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)
	res, err := svc.Export(ExportRequest{Format: "har"})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if res.EntryCount != 0 {
		t.Errorf("EntryCount = %d, want 0", res.EntryCount)
	}
	if !strings.HasSuffix(res.Path, ".har") {
		t.Errorf("Path = %q, want .har suffix", res.Path)
	}
}

func TestExportWorkspaceCopiesTree(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)

	res, err := svc.Export(ExportRequest{Format: "workspace"})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if _, err := os.Stat(filepath.Join(res.Path, "collections", "users", "list.json")); err != nil {
		t.Fatalf("exported workspace missing copied request: %v", err)
	}
}

func TestExportOutNameOverride(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)

	res, err := svc.Export(ExportRequest{Format: "postman", OutName: "custom"})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if filepath.Base(res.Path) != "custom.json" {
		t.Errorf("Path = %q, want custom.json", res.Path)
	}
}
