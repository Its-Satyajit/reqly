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

package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// writeFile writes a file into a temp dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// buildWorkspace creates an on-disk workspace with the given collection and
// returns its directory (files added afterwards require a reload).
func buildWorkspace(t *testing.T, baseURL string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "reqly.yaml", fmt.Sprintf("name: test-ws\nbaseURL: %s\n", baseURL))
	writeFile(t, dir, "collections/main/reqly.yaml", "name: main\n")
	return dir
}

// loadWorkspace loads and returns the workspace plus its single collection.
func loadWorkspace(t *testing.T, dir string) (*collections.Workspace, *collections.Collection) {
	t.Helper()
	ws, err := collections.LoadWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws.Collections) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(ws.Collections))
	}
	return ws, ws.Collections[0]
}

func TestRunCollectionSequential(t *testing.T) {
	var hits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "a.yaml", "request:\n  method: GET\n  url: /a\n")
	writeFile(t, collDir, "b.yaml", "request:\n  method: GET\n  url: /b\n")
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 2 || !report.OK() || report.Passed != 2 {
		t.Fatalf("report = %+v", report)
	}
	if len(hits) != 2 {
		t.Fatalf("expected 2 requests, got %v", hits)
	}
}

func TestRunCollectionFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "ok.yaml", "request:\n  method: GET\n  url: /ok\n")
	writeFile(t, collDir, "bad.yaml", "request:\n  method: GET\n  url: /bad\n")
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if report.OK() {
		t.Fatalf("expected failures, report = %+v", report)
	}
	if report.Passed != 1 || report.Failed != 1 {
		t.Fatalf("passed=%d failed=%d", report.Passed, report.Failed)
	}
	if len(report.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(report.Steps))
	}
}

func TestRunCollectionFailFast(t *testing.T) {
	var hits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "a.yaml", "request:\n  method: GET\n  url: /a\n")
	writeFile(t, collDir, "b.yaml", "request:\n  method: GET\n  url: /b\n")
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{FailFast: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 2 || len(report.Steps) != 1 {
		t.Fatalf("expected run to stop after first failure, report = %+v", report)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 request hit, got %v", hits)
	}
}

func TestRunCollectionOnStep(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "a.yaml", "request:\n  method: GET\n  url: /a\n")
	writeFile(t, collDir, "b.yaml", "request:\n  method: GET\n  url: /b\n")
	ws, coll := loadWorkspace(t, dir)

	var onStep []string
	report, err := RunCollection(context.Background(), ws, coll, nil, Options{OnStep: func(s StepResult) { onStep = append(onStep, s.Name) }})
	if err != nil {
		t.Fatal(err)
	}
	if len(onStep) != 2 || onStep[0] != "a" || onStep[1] != "b" {
		t.Fatalf("onStep = %v", onStep)
	}
	for i, s := range report.Steps {
		if s.Name != onStep[i] {
			t.Fatalf("onStep order %v != report order %v", onStep, s.Name)
		}
	}
}

func TestRunCollectionOnStepNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "a.yaml", "request:\n  method: GET\n  url: /a\n")
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 1 || !report.OK() {
		t.Fatalf("report = %+v", report)
	}
}

func TestRunCollectionCancelStopsScheduling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var onStep []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cancel()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "a.yaml", "request:\n  method: GET\n  url: /a\n")
	writeFile(t, collDir, "b.yaml", "request:\n  method: GET\n  url: /b\n")
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(ctx, ws, coll, nil, Options{OnStep: func(s StepResult) { onStep = append(onStep, s.Name) }})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Steps) != 1 || len(onStep) != 1 || onStep[0] != "a" {
		t.Fatalf("expected only step a, steps=%d onStep=%v", len(report.Steps), onStep)
	}
}

func TestRunCollectionFolderOrder(t *testing.T) {
	var hits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "z-root.yaml", "request:\n  method: GET\n  url: /z-root\n")
	writeFile(t, collDir, "sub/reqly.yaml", "name: sub\n")
	writeFile(t, collDir, "sub/folder.yaml", "request:\n  method: GET\n  url: /sub-folder\n")
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d: %+v", len(report.Steps), report.Steps)
	}
	// Root request runs before folder requests (folder traversal happens after
	// the container's own requests).
	if hits[0] != "/z-root" || hits[1] != "/sub-folder" {
		t.Fatalf("unexpected order: %v", hits)
	}
}

func TestRunFolder(t *testing.T) {
	var hits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "root.yaml", "request:\n  method: GET\n  url: /root\n")
	writeFile(t, collDir, "sub/reqly.yaml", "name: sub\n")
	writeFile(t, collDir, "sub/a.yaml", "request:\n  method: GET\n  url: /sub/a\n")
	writeFile(t, collDir, "sub/nested/reqly.yaml", "name: nested\n")
	writeFile(t, collDir, "sub/nested/b.yaml", "request:\n  method: GET\n  url: /sub/nested/b\n")
	ws, coll := loadWorkspace(t, dir)

	folder := coll.Folders[0]
	report, err := RunFolder(context.Background(), ws, coll, folder, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Steps) != 2 {
		t.Fatalf("expected 2 steps (folder + nested), got %d: %+v", len(report.Steps), report.Steps)
	}
	// Folder's own requests run before its nested folders.
	if hits[0] != "/sub/a" || hits[1] != "/sub/nested/b" {
		t.Fatalf("unexpected order: %v", hits)
	}
}

func TestRunCollectionPreScriptMutatesRequest(t *testing.T) {
	var gotHeader, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Injected")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "pre.yaml", fmt.Sprintf(`
request:
  method: GET
  url: /orig
preRequest: |
  reqly.request.url = %q;
  reqly.request.headers.set("X-Injected", "yes");
`, srv.URL+"/injected"))
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("report = %+v", report)
	}
	if gotPath != "/injected" || gotHeader != "yes" {
		t.Fatalf("path=%q header=%q", gotPath, gotHeader)
	}
}

func TestRunCollectionPreScriptError(t *testing.T) {
	dir := buildWorkspace(t, "http://example.com")
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "bad.yaml", "request:\n  method: GET\n  url: /x\npreRequest: throw new Error('boom')\n")
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if report.OK() || report.Failed != 1 {
		t.Fatalf("report = %+v", report)
	}
	if report.Steps[0].RequestError == nil {
		t.Fatal("expected request error from script")
	}
}

func TestRunCollectionPostScriptTests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true,"count":3}`)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "post.yaml", `
request:
  method: GET
  url: /data
postRequest: |
  reqly.test("status 200", function() { return reqly.response.status === 200; });
  reqly.test("has ok", function() { return reqly.response.body.indexOf("ok") !== -1; });
`)
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("report = %+v", report)
	}
	if len(report.Steps[0].Tests) != 2 {
		t.Fatalf("expected 2 tests, got %+v", report.Steps[0].Tests)
	}
	for _, tr := range report.Steps[0].Tests {
		if !tr.Passed {
			t.Fatalf("test %q should pass", tr.Name)
		}
	}
}

func TestRunCollectionPostScriptFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello")
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "post.yaml", `
request:
  method: GET
  url: /data
postRequest: |
  reqly.test("expects bye", function() { return reqly.response.body === "bye"; });
`)
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if report.OK() {
		t.Fatalf("expected step failure from failing test, report = %+v", report)
	}
	if len(report.Steps[0].Tests) != 1 || report.Steps[0].Tests[0].Passed {
		t.Fatalf("tests = %+v", report.Steps[0].Tests)
	}
}

func TestRunCollectionVariableChaining(t *testing.T) {
	// /login returns a token; /me requires it. The post-request script on login
	// extracts the token into a runtime variable consumed by /me.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"token":"tok-123"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") == "Bearer tok-123" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"user":"reqly"}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"unauthorized"}`)
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "login.yaml", `
request:
  method: POST
  url: /login
postRequest: |
  var token = JSON.parse(reqly.response.body).token;
  reqly.setVariable("token", token);
`)
	writeFile(t, collDir, "me.yaml", `
variables:
  auth: "Bearer {{token}}"
request:
  method: GET
  url: /me
  headers:
    - key: Authorization
      value: "{{auth}}"
`)
	ws, coll := loadWorkspace(t, dir)

	report, err := RunCollection(context.Background(), ws, coll, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("report = %+v", report)
	}
	if len(report.Steps) != 2 || !report.Steps[1].Response.OK() {
		t.Fatalf("chained /me should pass, report = %+v", report)
	}
}

func TestRunCollectionSharedVarsProvided(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, r.Header.Get("Authorization"))
	}))
	defer srv.Close()

	dir := buildWorkspace(t, srv.URL)
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "me.yaml", `
variables:
  auth: "Bearer {{token}}"
request:
  method: GET
  url: /me
  headers:
    - key: Authorization
      value: "{{auth}}"
`)
	ws, coll := loadWorkspace(t, dir)

	vars := variables.NewSet()
	vars.Set(variables.ScopeEnvironment, "token", "from-env")
	report, err := RunCollection(context.Background(), ws, coll, vars, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("report = %+v", report)
	}
	if string(report.Steps[0].Response.Body) != "Bearer from-env" {
		t.Fatalf("body = %q", report.Steps[0].Response.Body)
	}
}

func TestRunCollectionUsesHTTPClientOption(t *testing.T) {
	// A custom transport that short-circuits without network.
	var gotBody string
	dir := buildWorkspace(t, "http://api.example.com")
	collDir := filepath.Join(dir, "collections/main")
	writeFile(t, collDir, "a.yaml", "request:\n  method: GET\n  url: /a\n")
	ws, coll := loadWorkspace(t, dir)

	client := request.NewClient(request.WithHTTPClient(&http.Client{
		Transport: &stubTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				gotBody = "called"
				return &http.Response{StatusCode: 204, Body: http.NoBody}, nil
			},
		},
	}))
	report, err := RunCollection(context.Background(), ws, coll, nil, Options{Client: client})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("report = %+v", report)
	}
	if gotBody != "called" {
		t.Fatalf("transport not used: %q", gotBody)
	}
}

type stubTransport struct {
	roundTrip func(req *http.Request) (*http.Response, error)
}

func (t *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.roundTrip(req)
}

func TestStepResultJSON(t *testing.T) {
	r := StepResult{Name: "a", Passed: true, RequestPath: "/x.yaml"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}
