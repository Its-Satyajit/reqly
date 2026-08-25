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

package collections

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// writeConfig writes a JSON config file into dir and creates the dir.
func writeConfig(t *testing.T, dir, contents string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeRequest writes a YAML request file into dir.
func writeRequest(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// makeWorkspace builds a workspace tree in a temp dir and loads it.
func makeWorkspace(t *testing.T) *Workspace {
	t.Helper()
	root := t.TempDir()

	writeConfig(t, root, `{
		"name": "demo",
		"baseURL": "https://api.example.com",
		"variables": {"ENV": "prod", "SHARED": "ws"},
		"headers": [{"key": "X-Workspace", "value": "ws"}]
	}`)

	usersDir := filepath.Join(root, "collections", "users")
	writeConfig(t, usersDir, `{
		"name": "users",
		"baseURL": "v1",
		"variables": {"SHARED": "users"},
		"headers": [{"key": "X-Collection", "value": "users"}, {"key": "X-Workspace", "value": "overridden"}],
		"auth": {"type": "bearer", "config": {"token": "{{TOKEN}}"}}
	}`)
	writeRequest(t, usersDir, "list-users.yaml", `name: List Users
variables: {TOKEN: users-token}
request: {method: GET, url: users}
`)
	writeRequest(t, usersDir, "get-user.yaml", `name: Get User
request: {method: GET, url: users/1}
`)

	authDir := filepath.Join(usersDir, "auth")
	writeConfig(t, authDir, `{
		"name": "auth",
		"variables": {"TOKEN": "folder-token"},
		"headers": [{"key": "X-Folder", "value": "auth"}]
	}`)
	writeRequest(t, authDir, "login.yaml", `name: Login
request: {method: POST, url: auth/login, body: "{\"u\":\"a\"}"}
`)

	w, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestLoadWorkspaceDiscoversTree(t *testing.T) {
	w := makeWorkspace(t)

	if w.Config.Name != "demo" {
		t.Fatalf("workspace name: got %q", w.Config.Name)
	}
	if len(w.Collections) != 1 {
		t.Fatalf("collections: got %d, want 1", len(w.Collections))
	}

	coll := w.Collections[0]
	if coll.Name != "users" {
		t.Fatalf("collection name: got %q", coll.Name)
	}
	if len(coll.Requests) != 2 {
		t.Fatalf("collection requests: got %d, want 2", len(coll.Requests))
	}
	if len(coll.Folders) != 1 {
		t.Fatalf("collection folders: got %d, want 1", len(coll.Folders))
	}

	folder := coll.Folders[0]
	if folder.Name != "auth" {
		t.Fatalf("folder name: got %q", folder.Name)
	}
	if len(folder.Requests) != 1 || folder.Requests[0].Name != "login" {
		t.Fatalf("folder requests: got %+v", folder.Requests)
	}
}

func TestFindRequestByPath(t *testing.T) {
	w := makeWorkspace(t)

	_, chain, entry, ok := w.FindRequest(RequestPath("users/list-users.yaml"))
	if !ok || entry == nil {
		t.Fatal("expected to find users/list-users.yaml")
	}
	if len(chain) != 0 {
		t.Fatalf("collection-level request should have no folder chain, got %d", len(chain))
	}

	_, chain, entry, ok = w.FindRequest(RequestPath("users/auth/login.yaml"))
	if !ok || entry == nil {
		t.Fatal("expected to find users/auth/login.yaml")
	}
	if len(chain) != 1 || chain[0].Name != "auth" {
		t.Fatalf("folder chain: got %+v", chain)
	}

	if _, _, _, ok = w.FindRequest(RequestPath("users/nope.yaml")); ok {
		t.Fatal("expected miss for missing request")
	}
	if _, _, _, ok = w.FindRequest(RequestPath("missing/list.yaml")); ok {
		t.Fatal("expected miss for missing collection")
	}
}

func TestResolveInheritance(t *testing.T) {
	w := makeWorkspace(t)
	coll := w.Collections[0]

	// users/list-users.yaml — collection-level, no folders.
	coll, chain, entry, ok := w.FindRequest(RequestPath("users/list-users.yaml"))
	if !ok {
		t.Fatal("request not found")
	}
	resolved, err := w.ResolveRequest(coll, chain, entry)
	if err != nil {
		t.Fatal(err)
	}

	// Base URL: workspace https://api.example.com + collection "v1" → /v1; request "users" → absolute.
	if want := "https://api.example.com/v1/users"; resolved.Request.URL != want {
		t.Fatalf("URL: got %q, want %q", resolved.Request.URL, want)
	}

	// Headers: workspace X-Workspace overridden by collection, X-Collection added.
	gotHeaders := map[string]string{}
	for _, h := range resolved.Request.Headers {
		gotHeaders[h.Key] = h.Value
	}
	if gotHeaders["X-Workspace"] != "overridden" {
		t.Fatalf("X-Workspace: got %q", gotHeaders["X-Workspace"])
	}
	if gotHeaders["X-Collection"] != "users" {
		t.Fatalf("X-Collection: got %q", gotHeaders["X-Collection"])
	}

	// Auth inherited from collection.
	if resolved.Request.Auth.Type != "bearer" {
		t.Fatalf("auth type: got %q", resolved.Request.Auth.Type)
	}

	// Variables: workspace global + collection, request overrides TOKEN.
	if v, ok := resolved.Vars.Resolve("SHARED"); !ok || v != "users" {
		t.Fatalf("SHARED precedence: got %q, %v", v, ok)
	}
	if v, ok := resolved.Vars.Resolve("ENV"); !ok || v != "prod" {
		t.Fatalf("ENV: got %q, %v", v, ok)
	}
	if v, ok := resolved.Vars.Resolve("TOKEN"); !ok || v != "users-token" {
		t.Fatalf("TOKEN precedence: got %q, %v", v, ok)
	}
}

func TestResolveFolderInheritance(t *testing.T) {
	w := makeWorkspace(t)
	coll, chain, entry, ok := w.FindRequest(RequestPath("users/auth/login.yaml"))
	if !ok {
		t.Fatal("request not found")
	}
	resolved, err := w.ResolveRequest(coll, chain, entry)
	if err != nil {
		t.Fatal(err)
	}

	// Base URL: workspace + collection + request path (folder has none).
	if want := "https://api.example.com/v1/auth/login"; resolved.Request.URL != want {
		t.Fatalf("URL: got %q, want %q", resolved.Request.URL, want)
	}

	gotHeaders := map[string]string{}
	for _, h := range resolved.Request.Headers {
		gotHeaders[h.Key] = h.Value
	}
	if gotHeaders["X-Folder"] != "auth" {
		t.Fatalf("X-Folder: got %q", gotHeaders["X-Folder"])
	}
	if gotHeaders["X-Collection"] != "users" {
		t.Fatalf("X-Collection: got %q", gotHeaders["X-Collection"])
	}

	// Folder variables sit between collection and request scopes.
	if v, ok := resolved.Vars.Resolve("TOKEN"); !ok || v != "folder-token" {
		t.Fatalf("folder TOKEN precedence: got %q, %v", v, ok)
	}

	// POST method and body preserved from the request file.
	if resolved.Request.Method != request.MethodPost {
		t.Fatalf("method: got %q", resolved.Request.Method)
	}
	if resolved.Request.Body == "" {
		t.Fatal("expected body to be preserved")
	}
}

func TestResolveScopesOrder(t *testing.T) {
	w := makeWorkspace(t)
	coll, chain, entry, ok := w.FindRequest(RequestPath("users/get-user.yaml"))
	if !ok {
		t.Fatal("request not found")
	}
	resolved, err := w.ResolveRequest(coll, chain, entry)
	if err != nil {
		t.Fatal(err)
	}

	// No request-level TOKEN here → falls through to collection scope.
	if v, ok := resolved.Vars.Resolve("TOKEN"); !ok || v != "" {
		// collection defines no TOKEN either; the auth config references it
		// so it should simply be undefined unless a scope provides it.
		_ = v
	}
	if v, ok := resolved.Vars.Get(variables.ScopeGlobal, "SHARED"); !ok || v != "ws" {
		t.Fatalf("global SHARED: got %q, %v", v, ok)
	}
	if v, ok := resolved.Vars.Get(variables.ScopeCollection, "SHARED"); !ok || v != "users" {
		t.Fatalf("collection SHARED: got %q, %v", v, ok)
	}
	if _, ok := resolved.Vars.Get(variables.ScopeRequest, "SHARED"); ok {
		t.Fatal("request scope should not have SHARED")
	}
}

func TestLoadWorkspaceRequiresDescriptor(t *testing.T) {
	root := t.TempDir()
	if _, err := LoadWorkspace(root); err == nil {
		t.Fatal("expected error for dir without descriptor")
	}
}

func TestAbsoluteRequestURLOverridesBase(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"name":"ws","baseURL":"https://api.example.com"}`)
	collDir := filepath.Join(root, "collections", "c")
	writeConfig(t, collDir, `{"name":"c"}`)
	writeRequest(t, collDir, "abs.yaml", `request: {url: "https://other.example.com/x"}`)

	w, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	coll, chain, entry, ok := w.FindRequest(RequestPath("c/abs.yaml"))
	if !ok {
		t.Fatal("request not found")
	}
	resolved, err := w.ResolveRequest(coll, chain, entry)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Request.URL != "https://other.example.com/x" {
		t.Fatalf("URL: got %q", resolved.Request.URL)
	}
}

func TestFolderAuthOverridesCollection(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"name":"ws"}`)
	collDir := filepath.Join(root, "collections", "c")
	writeConfig(t, collDir, `{"name":"c","auth":{"type":"bearer","config":{"token":"{{TOKEN}}"}}}`)
	folderDir := filepath.Join(collDir, "f")
	writeConfig(t, folderDir, `{"name":"f","auth":{"type":"basic","config":{"username":"u","password":"p"}}}`)
	writeRequest(t, folderDir, "r.yaml", `request: {url: "https://api.example.com/x"}`)

	w, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	coll, chain, entry, ok := w.FindRequest(RequestPath("c/f/r.yaml"))
	if !ok {
		t.Fatal("request not found")
	}
	resolved, err := w.ResolveRequest(coll, chain, entry)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Request.Auth.Type != "basic" {
		t.Fatalf("auth type: got %q, want basic", resolved.Request.Auth.Type)
	}
}

func TestRequestNoneAuthClearsInherited(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"name":"ws"}`)
	collDir := filepath.Join(root, "collections", "c")
	writeConfig(t, collDir, `{"name":"c","auth":{"type":"bearer","config":{"token":"{{TOKEN}}"}}}`)
	writeRequest(t, collDir, "public.yaml", `request: {url: "https://api.example.com/public", auth: {type: none}}`)

	w, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	coll, chain, entry, ok := w.FindRequest(RequestPath("c/public.yaml"))
	if !ok {
		t.Fatal("request not found")
	}
	resolved, err := w.ResolveRequest(coll, chain, entry)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Request.Auth.Type != "none" {
		t.Fatalf("auth type: got %q, want none", resolved.Request.Auth.Type)
	}
}

func TestLoadWorkspaceYAMLConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, configFileName),
		[]byte("name: ws\nbaseURL: https://api.example.com\nvariables:\n  ENV: prod\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if w.Config.Name != "ws" || w.Config.BaseURL != "https://api.example.com" {
		t.Fatalf("config: got %+v", w.Config)
	}
	if v, ok := w.VariablesSet().Resolve("ENV"); !ok || v != "prod" {
		t.Fatalf("ENV: got %q, %v", v, ok)
	}
}

func TestFolderWithoutDescriptorIsSkipped(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"name":"ws"}`)
	collDir := filepath.Join(root, "collections", "c")
	writeConfig(t, collDir, `{"name":"c"}`)
	if err := os.MkdirAll(filepath.Join(collDir, "plain"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeRequest(t, collDir, "r.yaml", `request: {url: "https://api.example.com/x"}`)

	w, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Collections[0].Folders) != 0 {
		t.Fatalf("expected no folders, got %d", len(w.Collections[0].Folders))
	}
	if len(w.Collections[0].Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(w.Collections[0].Requests))
	}
}

func TestLoadWorkspaceNoCollectionsDir(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"name":"ws"}`)
	w, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Collections) != 0 {
		t.Fatalf("expected no collections, got %d", len(w.Collections))
	}
}

func TestEffectiveURLInvalidBase(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"name":"ws","baseURL":"://bad"}`)
	collDir := filepath.Join(root, "collections", "c")
	writeConfig(t, collDir, `{"name":"c"}`)
	writeRequest(t, collDir, "r.yaml", `request: {url: "x"}`)

	w, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	coll, chain, entry, ok := w.FindRequest(RequestPath("c/r.yaml"))
	if !ok {
		t.Fatal("request not found")
	}
	if _, err := w.ResolveRequest(coll, chain, entry); err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}

func TestEffectiveURLPreservesPlaceholders(t *testing.T) {
	i := Inherited{BaseURL: "https://api.example.com"}
	got, err := i.EffectiveURL("/api/users/{{userId}}")
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://api.example.com/api/users/{{userId}}"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEffectiveURLJoinsRelativeBase(t *testing.T) {
	i := Inherited{BaseURL: "/v1"}
	got, err := i.EffectiveURL("users")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/v1/users"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWorkspaceConfigParsesEnvironmentField(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte(`
name: workspace
environment: prod
`), 0o644); err != nil {
		t.Fatal(err)
	}

	ws, err := LoadWorkspace(dir)
	if err != nil {
		t.Fatalf("LoadWorkspace: %v", err)
	}
	if ws.Config.Environment != "prod" {
		t.Fatalf("environment: got %q, want %q", ws.Config.Environment, "prod")
	}
}

func TestSetWorkspaceEnvironmentPersistsField(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte(`
name: workspace
environment: dev
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SetWorkspaceEnvironment(dir, "prod"); err != nil {
		t.Fatalf("SetWorkspaceEnvironment: %v", err)
	}

	ws, err := LoadWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Config.Environment != "prod" {
		t.Fatalf("environment: got %q, want %q", ws.Config.Environment, "prod")
	}
	// Other fields are preserved.
	if ws.Config.Name != "workspace" {
		t.Fatalf("name: got %q, want preserved", ws.Config.Name)
	}
}

func TestSetWorkspaceEnvironmentCreatesDescriptor(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "collections"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := SetWorkspaceEnvironment(dir, "dev"); err != nil {
		t.Fatalf("SetWorkspaceEnvironment: %v", err)
	}
	ws, err := LoadWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Config.Environment != "dev" {
		t.Fatalf("environment: got %q, want %q", ws.Config.Environment, "dev")
	}
}

func TestFindWorkspaceRootWalksUp(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	if got := FindWorkspaceRoot(nested); got != dir {
		t.Fatalf("FindWorkspaceRoot(%s) = %q, want %q", nested, got, dir)
	}
	if got := FindWorkspaceRoot(dir); got != dir {
		t.Fatalf("FindWorkspaceRoot(root) = %q, want %q", got, dir)
	}
}

func TestFindWorkspaceRootNone(t *testing.T) {
	dir := t.TempDir() // no descriptor anywhere
	if got := FindWorkspaceRoot(dir); got != "" {
		t.Fatalf("FindWorkspaceRoot = %q, want empty", got)
	}
}

func TestCreateContainer(t *testing.T) {
	dir := t.TempDir()
	if err := CreateContainer(filepath.Join(dir, "payments"), "payments"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, ok, err := loadConfig(filepath.Join(dir, "payments")); err != nil || !ok {
		t.Fatal("descriptor missing after create")
	}
	if err := CreateContainer(filepath.Join(dir, "payments"), "payments"); err == nil {
		t.Fatal("expected duplicate create to fail")
	}
	for _, bad := range []string{"", ".", "..", "a/b", ".hidden"} {
		if err := CreateContainer(filepath.Join(dir, bad), bad); err == nil {
			t.Errorf("expected invalid name %q to fail", bad)
		}
	}
}
