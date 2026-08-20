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
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
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

// writeResolvableWorkspace builds a workspace with a workspace + collection +
// folder inheritance chain (base URL join, header merge, auth inheritance,
// variable scopes) and a file that pins an environment.
func writeResolvableWorkspace(t *testing.T, root string) {
	t.Helper()
	files := map[string]string{
		"reqly.yaml": `
name: demo
baseURL: https://api.example.com
variables:
  ENV: prod
  SHARED: ws
headers:
  - key: X-Workspace
    value: ws
`,
		"collections/users/reqly.yaml": `
name: users
baseURL: v1
variables:
  SHARED: users
headers:
  - key: X-Collection
    value: users
auth:
  type: bearer
  config:
    token: "{{TOKEN}}"
`,
		"collections/users/list-users.yaml": `
name: List Users
environment: staging
variables:
  TOKEN: req-token
request:
  method: GET
  url: users
`,
		"collections/users/auth/reqly.yaml": `
name: auth
variables:
  TOKEN: folder-token
`,
		"collections/users/auth/login.yaml": `
request:
  method: POST
  url: auth/login
`,
	}
	for name, contents := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWorkspaceServiceOpenRequestResolvesInheritance(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	opened, err := svc.OpenRequest("users/list-users")
	if err != nil {
		t.Fatal(err)
	}

	if opened.Path != "users/list-users" || opened.Name != "list-users" {
		t.Fatalf("path/name = %q/%q", opened.Path, opened.Name)
	}
	if opened.FileEnv != "staging" {
		t.Fatalf("fileEnv = %q, want staging", opened.FileEnv)
	}

	req := opened.Request
	if req.URL != "https://api.example.com/v1/users" {
		t.Fatalf("URL = %q, want https://api.example.com/v1/users", req.URL)
	}
	if req.Method != "GET" {
		t.Fatalf("method = %q, want GET", req.Method)
	}
	// Header merge: workspace + collection (collection wins on X-Workspace).
	headers := map[string]string{}
	for _, h := range req.Headers {
		headers[h.Key] = h.Value
	}
	if headers["X-Workspace"] != "ws" || headers["X-Collection"] != "users" {
		t.Fatalf("headers = %v", headers)
	}
	// Inherited auth applied silently (bearer with a token placeholder).
	if req.Auth.Type != "bearer" {
		t.Fatalf("auth type = %q, want bearer", req.Auth.Type)
	}
}

func TestWorkspaceServiceOpenRequestVariablesInScopeOrder(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	opened, err := svc.OpenRequest("users/auth/login")
	if err != nil {
		t.Fatal(err)
	}

	// Scope order low → high: global, collection, folder, request.
	var scopes []string
	byName := map[string]string{}
	for _, v := range opened.Variables {
		scopes = append(scopes, v.Scope)
		byName[v.Name] = v.Scope
	}
	if byName["SHARED"] != "collection" {
		t.Fatalf("SHARED scope = %q, want collection (wins over workspace)", byName["SHARED"])
	}
	if byName["TOKEN"] != "folder" {
		t.Fatalf("TOKEN scope = %q, want folder (login has no request vars)", byName["TOKEN"])
	}
	if byName["ENV"] != "global" {
		t.Fatalf("ENV scope = %q, want global", byName["ENV"])
	}
	if len(scopes) == 0 {
		t.Fatal("expected variables")
	}
	// No scope appears before an earlier-scope entry.
	order := map[string]int{"global": 0, "collection": 1, "folder": 2, "request": 3}
	prev := -1
	for _, s := range scopes {
		if order[s] < prev {
			t.Fatalf("scopes out of order: %v", scopes)
		}
		prev = order[s]
	}
}

func TestWorkspaceServiceOpenRequestMissingErrors(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	for _, path := range []string{"users/nope", "missing/thing", ""} {
		if _, err := svc.OpenRequest(path); err == nil {
			t.Fatalf("expected error for path %q, got nil", path)
		}
	}
}

func TestWorkspaceServiceOpenRequestWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml

	svc := NewWorkspaceService(dir)
	if _, err := svc.OpenRequest("users/list-users"); err == nil {
		t.Fatal("expected error without a workspace, got nil")
	}
}

func TestWorkspaceServiceOpenRequestIncludesFileRequestAndVersion(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	opened, err := svc.OpenRequest("users/list-users")
	if err != nil {
		t.Fatal(err)
	}

	// FileRequest is the raw, unmerged file-owned request: the editor seed,
	// without inherited base URL, headers, or auth.
	if opened.FileRequest.URL != "users" {
		t.Fatalf("fileRequest URL = %q, want users (unmerged)", opened.FileRequest.URL)
	}
	if len(opened.FileRequest.Headers) != 0 {
		t.Fatalf("fileRequest headers = %v, want none", opened.FileRequest.Headers)
	}
	if opened.FileRequest.Auth.Type != "" {
		t.Fatalf("fileRequest auth = %q, want none (inherited only)", opened.FileRequest.Auth.Type)
	}

	// Version fingerprints the raw file bytes so saves can detect concurrent
	// on-disk edits.
	raw, err := os.ReadFile(filepath.Join(dir, "collections/users/list-users.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if opened.Version != requestfile.Fingerprint(raw) {
		t.Fatalf("version = %q, want fingerprint of the raw file", opened.Version)
	}
}

func TestWorkspaceServiceSaveRequestWritesBuilderFieldsOnly(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)
	// A file exercising every preserved field.
	path := "collections/users/edit-me.yaml"
	full := `name: Edit Me
environment: staging
variables:
  TOKEN: file-token
preRequest: console.log("pre")
postRequest: reqly.test("ok", true)
request:
  method: POST
  url: edit
  headers:
    - key: X-Own
      value: mine
  auth:
    type: basic
    config:
      username: u
      password: p
  timeout: 2500
`
	if err := os.WriteFile(filepath.Join(dir, path), []byte(full), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewWorkspaceService(dir)
	opened, err := svc.OpenRequest("users/edit-me")
	if err != nil {
		t.Fatal(err)
	}

	draft := opened.FileRequest
	draft.URL = "edited"
	draft.Method = "PUT"
	draft.Headers = []request.Header{{Key: "X-Own", Value: "changed"}}
	draft.Query = []request.Parameter{{Key: "q", Value: "1"}}
	draft.Body = `{"x":1}`

	version, err := svc.SaveRequest("users/edit-me", draft, opened.Version)
	if err != nil {
		t.Fatal(err)
	}
	if version == "" {
		t.Fatal("expected a new version fingerprint")
	}

	reloaded, err := svc.OpenRequest("users/edit-me")
	if err != nil {
		t.Fatal(err)
	}
	// Builder fields written.
	if reloaded.Request.URL != "https://api.example.com/v1/edited" {
		t.Fatalf("saved URL = %q, want https://api.example.com/v1/edited", reloaded.Request.URL)
	}
	if reloaded.Request.Method != "PUT" {
		t.Fatalf("saved method = %q, want PUT", reloaded.Request.Method)
	}
	if reloaded.Request.Body != `{"x":1}` {
		t.Fatalf("saved body = %q", reloaded.Request.Body)
	}
	if len(reloaded.Request.Query) != 1 || reloaded.Request.Query[0].Key != "q" {
		t.Fatalf("saved query = %v", reloaded.Request.Query)
	}
	headers := map[string]string{}
	for _, h := range reloaded.Request.Headers {
		headers[h.Key] = h.Value
	}
	if headers["X-Own"] != "changed" {
		t.Fatalf("saved headers = %v", reloaded.Request.Headers)
	}
	// Non-editable fields preserved verbatim.
	if reloaded.FileRequest.Auth.Type != "basic" || reloaded.FileRequest.Auth.Config["username"] != "u" {
		t.Fatalf("auth not preserved: %+v", reloaded.FileRequest.Auth)
	}
	if reloaded.FileRequest.Timeout != 2500 {
		t.Fatalf("timeout not preserved: %d", reloaded.FileRequest.Timeout)
	}
	if reloaded.FileEnv != "staging" {
		t.Fatalf("environment not preserved: %q", reloaded.FileEnv)
	}
	// Request-file-level fields (name, variables, scripts) preserved on disk.
	raw, err := os.ReadFile(filepath.Join(dir, path))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"name: Edit Me", "TOKEN: file-token", "preRequest", "postRequest", `console.log("pre")`} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("saved file lost %q:\n%s", want, raw)
		}
	}
}

func TestWorkspaceServiceSaveRequestRejectsChangedOnDisk(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	opened, err := svc.OpenRequest("users/list-users")
	if err != nil {
		t.Fatal(err)
	}

	// Another editor changes the file on disk after it was opened.
	path := filepath.Join(dir, "collections/users/list-users.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, []byte("\n# external edit\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.SaveRequest("users/list-users", opened.FileRequest, opened.Version); !errors.Is(err, ErrFileChangedOnDisk) {
		t.Fatalf("expected ErrFileChangedOnDisk, got %v", err)
	}
	// The file must be untouched after a rejected save.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "external edit") {
		t.Fatal("file was modified despite rejected save")
	}
}

func TestWorkspaceServiceSaveRequestMissingFileErrors(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	if _, err := svc.SaveRequest("users/nope", request.Request{}, "v"); err == nil {
		t.Fatal("expected error saving an unknown request")
	}
}

func TestWorkspaceServiceResolveSendAppliesDraft(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	draft := request.Request{
		Method:  "POST",
		URL:     "custom",
		Headers: []request.Header{{Key: "X-Own", Value: "mine"}},
	}
	resolved, err := svc.ResolveSend("users/list-users", draft)
	if err != nil {
		t.Fatal(err)
	}

	if resolved.Request.URL != "https://api.example.com/v1/custom" {
		t.Fatalf("URL = %q, want https://api.example.com/v1/custom", resolved.Request.URL)
	}
	if resolved.Request.Method != "POST" {
		t.Fatalf("method = %q, want POST", resolved.Request.Method)
	}
	headers := map[string]string{}
	for _, h := range resolved.Request.Headers {
		headers[h.Key] = h.Value
	}
	if headers["X-Workspace"] != "ws" || headers["X-Collection"] != "users" || headers["X-Own"] != "mine" {
		t.Fatalf("headers = %v", resolved.Request.Headers)
	}
	if resolved.Request.Auth.Type != "bearer" {
		t.Fatalf("auth = %q, want inherited bearer", resolved.Request.Auth.Type)
	}
	if v, ok := resolved.Vars.Resolve("TOKEN"); !ok || v != "req-token" {
		t.Fatalf("TOKEN = %q, %v; want req-token from the request scope", v, ok)
	}
}

func TestWorkspaceServiceResolveSendInheritAppliesContainerAuth(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)
	// A file with its own auth that the draft leaves unset (Inherit): the
	// send must treat the draft as authoritative, so the container's inherited
	// auth applies rather than the file's now-dropped auth.
	path := "collections/users/own-auth.yaml"
	full := `request:
  method: GET
  url: own
  auth:
    type: basic
    config:
      username: file-user
      password: file-pass
`
	if err := os.WriteFile(filepath.Join(dir, path), []byte(full), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewWorkspaceService(dir)
	// The draft carries builder fields and no auth — that is what the Auth tab
	// sends for Inherit.
	resolved, err := svc.ResolveSend("users/own-auth", request.Request{Method: "GET", URL: "edited"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Request.Auth.Type != "bearer" {
		t.Fatalf("auth = %q, want inherited bearer (draft Inherit drops file auth)", resolved.Request.Auth.Type)
	}
	if resolved.Request.URL != "https://api.example.com/v1/edited" {
		t.Fatalf("URL = %q, want https://api.example.com/v1/edited", resolved.Request.URL)
	}
}

func TestWorkspaceServiceSaveRequestAuthSemantics(t *testing.T) {
	cases := []struct {
		name         string
		initial      string
		draftAuth    request.Auth
		wantAuth     request.Auth
		wantContains []string
		wantNot      []string
	}{
		{
			name: "typed draft auth becomes the file's",
			initial: `request:
  method: GET
  url: own
`,
			draftAuth:    request.Auth{Type: "bearer", Config: map[string]string{"token": "t0ken"}},
			wantAuth:     request.Auth{Type: "bearer", Config: map[string]string{"token": "t0ken"}},
			wantContains: []string{"type: bearer", "token: t0ken"},
		},
		{
			name: "none writes an explicit block",
			initial: `request:
  method: GET
  url: own
`,
			draftAuth:    request.Auth{Type: "none"},
			wantAuth:     request.Auth{Type: "none"},
			wantContains: []string{"type: none"},
		},
		{
			name: "inherit removes an existing block",
			initial: `request:
  method: GET
  url: own
  auth:
    type: basic
    config:
      username: u
      password: p
`,
			draftAuth:    request.Auth{},
			wantAuth:     request.Auth{},
			wantContains: []string{"url: own"},
			wantNot:      []string{"auth:", "username:"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeResolvableWorkspace(t, dir)
			path := "collections/users/own-auth.yaml"
			if err := os.WriteFile(filepath.Join(dir, path), []byte(tc.initial), 0o644); err != nil {
				t.Fatal(err)
			}

			svc := NewWorkspaceService(dir)
			opened, err := svc.OpenRequest("users/own-auth")
			if err != nil {
				t.Fatal(err)
			}

			draft := opened.FileRequest
			draft.Auth = tc.draftAuth
			if _, err := svc.SaveRequest("users/own-auth", draft, opened.Version); err != nil {
				t.Fatal(err)
			}

			reloaded, err := svc.OpenRequest("users/own-auth")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(reloaded.FileRequest.Auth, tc.wantAuth) {
				t.Fatalf("auth = %+v, want %+v", reloaded.FileRequest.Auth, tc.wantAuth)
			}
			raw, err := os.ReadFile(filepath.Join(dir, path))
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(string(raw), want) {
					t.Fatalf("saved file lost %q:\n%s", want, raw)
				}
			}
			for _, want := range tc.wantNot {
				if strings.Contains(string(raw), want) {
					t.Fatalf("saved file still holds %q:\n%s", want, raw)
				}
			}
		})
	}
}

func TestWorkspaceServiceResolveSendDraftAuthOverridesInherited(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	draft := request.Request{
		Method: "GET",
		URL:    "custom",
		Auth:   request.Auth{Type: "basic", Config: map[string]string{"username": "u", "password": "p"}},
	}
	resolved, err := svc.ResolveSend("users/list-users", draft)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Request.Auth.Type != "basic" || resolved.Request.Auth.Config["username"] != "u" {
		t.Fatalf("auth = %+v, want draft basic overriding inherited bearer", resolved.Request.Auth)
	}
}

func TestWorkspaceServiceResolveSendNoneDisablesInherited(t *testing.T) {
	dir := t.TempDir()
	writeResolvableWorkspace(t, dir)

	svc := NewWorkspaceService(dir)
	draft := request.Request{
		Method: "GET",
		URL:    "custom",
		Auth:   request.Auth{Type: "none"},
	}
	resolved, err := svc.ResolveSend("users/list-users", draft)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Request.Auth.Type != "none" {
		t.Fatalf("auth = %+v, want explicit none disabling the inherited bearer", resolved.Request.Auth)
	}
}
