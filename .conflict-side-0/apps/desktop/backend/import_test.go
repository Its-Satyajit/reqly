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
	"runtime"
	"strings"
	"testing"
)

// fixture reads an importer test fixture so bridge tests exercise the same
// bytes the parsers are validated against. Anchored to this source file so
// t.Chdir cannot break resolution.
func fixture(t *testing.T, parts ...string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve source file location")
	}
	all := append([]string{
		filepath.Dir(thisFile), "..", "..", "..",
		"internal", "importer", "testdata", "import-suite",
	}, parts...)
	data, err := os.ReadFile(filepath.Join(all...))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(data)
}

// newServiceInWorkspace creates a minimal reqly workspace in a fresh temp
// directory and an AppService rooted there.
func newServiceInWorkspace(t *testing.T) (*AppService, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("name: test-workspace\n"), 0o644); err != nil {
		t.Fatalf("write reqly.yaml: %v", err)
	}
	t.Chdir(dir)
	return NewAppService(), dir
}

func TestImportDryRunPreviewsWithoutWriting(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)

	before := dirEntries(t, wsDir)

	res, err := svc.Import(ImportRequest{
		Content: fixture(t, "postman", "fixtures", "collection-v2.json"),
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("Import dry-run: %v", err)
	}
	if res.Kind != ImportKindWorkspace {
		t.Errorf("Kind = %q, want %q", res.Kind, ImportKindWorkspace)
	}
	if res.Format != "postman" {
		t.Errorf("Format = %q, want %q", res.Format, "postman")
	}
	if res.RequestCount == 0 {
		t.Error("RequestCount = 0, want requests previewed")
	}
	if res.TargetDir == "" {
		t.Error("TargetDir empty, want sanitized title suggestion")
	}
	if res.Report == nil {
		t.Error("Report nil, want structured import report")
	}

	after := dirEntries(t, wsDir)
	if len(after) != len(before) {
		t.Errorf("dry run changed the workspace (%d → %d entries), want no writes", len(before), len(after))
	}
}

// dirEntries lists a directory's entry names in sorted order.
func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestImportCommitWritesWorkspaceFolder(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)

	_, err := svc.Import(ImportRequest{
		Content:   fixture(t, "bruno", "fixtures", "bruno-testbench.json"),
		DryRun:    true,
		TargetDir: "my-import",
	})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	res, err := svc.Import(ImportRequest{
		Content:   fixture(t, "bruno", "fixtures", "bruno-testbench.json"),
		TargetDir: "my-import",
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	written := filepath.Join(wsDir, "my-import")
	if _, err := os.Stat(filepath.Join(written, "reqly.yaml")); err != nil {
		t.Errorf("commit did not create %s: %v", filepath.Join(written, "reqly.yaml"), err)
	}
	if res.TargetDir != "my-import" {
		t.Errorf("TargetDir = %q, want my-import", res.TargetDir)
	}
}

func TestImportCommitConflictFailsFast(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)
	content := fixture(t, "insomnia", "fixtures", "insomnia-v5.yaml")

	if _, err := svc.Import(ImportRequest{Content: content}); err != nil {
		t.Fatalf("first commit: %v", err)
	}
	_, err := svc.Import(ImportRequest{Content: content})
	if err == nil {
		t.Fatal("second commit to same target succeeded, want conflict error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v, want it to mention the existing folder", err)
	}
}

func TestImportHARPreviewAndCommit(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)

	har := `{"log":{"version":"1.2","creator":{"name":"t"},"entries":[` +
		`{"startedDateTime":"2026-01-01T00:00:00Z","time":0,` +
		`"request":{"method":"GET","url":"https://example.com/a"},` +
		`		"response":{"status":200}}]}}`
	res, err := svc.Import(ImportRequest{Content: har, DryRun: true})
	if err != nil {
		t.Fatalf("HAR dry-run: %v", err)
	}
	if res.Format != "har" || res.RequestCount != 1 {
		t.Errorf("dry-run = (%q, %d), want (har, 1)", res.Format, res.RequestCount)
	}
	if _, err := svc.Import(ImportRequest{Content: har}); err != nil {
		t.Fatalf("HAR commit: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wsDir, res.TargetDir, "reqly.yaml")); err != nil {
		t.Errorf("commit did not create a standalone workspace folder: %v", err)
	}
}

func TestImportHintOverridesDetection(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)

	// Detection would classify this payload as Bruno; claiming it is Postman
	// must route to the Postman parser (which then rejects it), proving the
	// hint wins over sniffing.
	_, err := svc.Import(ImportRequest{
		Content:    `{"name":"c","version":"1","items":[]}`,
		FormatHint: "postman",
	})
	if err == nil {
		t.Fatal("bruno payload with postman hint parsed as bruno, want postman parse failure")
	}
	if !strings.Contains(err.Error(), "info") {
		t.Errorf("error = %v, want the Postman parser's missing-info complaint", err)
	}
}

func TestImportInvalidHintErrors(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)

	_, err := svc.Import(ImportRequest{
		Content:    "curl https://example.com",
		FormatHint: "yaml-thing",
	})
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Fatalf("error = %v, want unknown-format complaint about the hint", err)
	}
}

func TestDetectBridge(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	svc := NewAppService()

	format, ok := svc.Detect("curl https://example.com")
	if !ok || format != "curl" {
		t.Errorf("Detect(curl) = (%q, %v), want (curl, true)", format, ok)
	}
	if _, ok := svc.Detect(""); ok {
		t.Error("Detect(empty) reported a match, want none")
	}
}

func TestImportUndetectableContentErrors(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)

	for name, content := range map[string]string{
		"empty":        "",
		"garbage json": `{"foo":1}`,
		"plain text":   "hello world",
	} {
		_, err := svc.Import(ImportRequest{Content: content})
		if err == nil {
			t.Errorf("%s: Import succeeded, want undetectable-format error", name)
		}
	}
}

func TestImportOpenAPIPreviewOperations(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	spec := fixture(t, "openapi", "cli", "fixtures", "openapi.json")

	res, err := svc.Import(ImportRequest{Content: spec, DryRun: true})
	if err != nil {
		t.Fatalf("OpenAPI dry-run: %v", err)
	}
	if res.Format != "openapi" {
		t.Errorf("Format = %q, want openapi", res.Format)
	}
	if len(res.Operations) == 0 {
		t.Fatal("Operations empty, want flattened operation list for preview")
	}
	for _, op := range res.Operations[:1] {
		if op.Method == "" || op.Path == "" {
			t.Errorf("operation missing method/path: %+v", op)
		}
	}

	if _, err := svc.Import(ImportRequest{Content: spec}); err != nil {
		t.Fatalf("OpenAPI commit: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wsDir, res.TargetDir, "reqly.yaml")); err != nil {
		t.Errorf("commit did not create a standalone workspace folder: %v", err)
	}
	_, err = svc.Import(ImportRequest{Content: spec})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("second commit error = %v, want conflict", err)
	}
}

func TestImportCurlOpensAsRequest(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	before := dirEntries(t, wsDir)

	res, err := svc.Import(ImportRequest{
		Content: "curl -X POST -H 'x-a: b' -d '{\"k\":1}' https://example.com/api",
	})
	if err != nil {
		t.Fatalf("cURL import: %v", err)
	}
	if res.Kind != ImportKindRequest {
		t.Errorf("Kind = %q, want %q", res.Kind, ImportKindRequest)
	}
	if res.Request == nil {
		t.Fatal("Request nil, want parsed cURL request")
	}
	if res.Request.Method != "POST" {
		t.Errorf("Method = %q, want POST", res.Request.Method)
	}
	if len(res.TargetDir) != 0 {
		t.Errorf("TargetDir = %q, want empty (no filesystem target)", res.TargetDir)
	}

	after := dirEntries(t, wsDir)
	if len(after) != len(before) {
		t.Errorf("cURL import wrote %d entries, want none", len(after)-len(before))
	}
}
