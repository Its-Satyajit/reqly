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

package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/collections"
)

func fixtureBytes(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "import-suite", rel))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

const insomniaV4Inline = `{
  "_type": "export",
  "__export_format": 4,
  "__export_source": "test:v0",
  "resources": [
    { "_type": "workspace", "_id": "wrk_1", "name": "Inline API" },
    { "_type": "request_group", "_id": "fld_1", "parentId": "wrk_1", "name": "Auth" },
    {
      "_type": "request", "_id": "req_1", "parentId": "fld_1",
      "name": "Login", "method": "post", "url": "{{ _.base_url }}/login",
      "headers": [
        { "name": "Accept", "value": "application/json" },
        { "name": "X-Drop", "value": "y", "disabled": true }
      ],
      "parameters": [ { "name": "debug", "value": "1" }, { "name": "off", "value": "x", "disabled": true } ],
      "body": { "mimeType": "application/json", "text": "{\"u\":\"a\"}" },
      "authentication": { "type": "basic", "username": "u", "password": "p" }
    },
    {
      "_type": "request", "_id": "req_orphan", "parentId": "wrk_missing",
      "name": "Orphan", "method": "GET", "url": "/orphan"
    },
    {
      "_type": "environment", "_id": "env_1", "parentId": "wrk_1",
      "name": "Base Env",
      "data": { "base_url": "https://api.test", "user": { "name": "admin" }, "retries": 3 }
    },
    { "_type": "cookie_jar", "_id": "jar_1", "parentId": "wrk_1", "name": "Jar" }
  ]
}`

func TestParseInsomniaV4(t *testing.T) {
	res, warnings, err := ParseInsomnia([]byte(insomniaV4Inline))
	if err != nil {
		t.Fatal(err)
	}
	if res.Title != "Inline API" {
		t.Fatalf("title = %q", res.Title)
	}
	if len(res.Root.Folders) != 1 || res.Root.Folders[0].Name != "Auth" {
		t.Fatalf("folders = %+v", res.Root.Folders)
	}
	folder := res.Root.Folders[0]
	if len(folder.Requests) != 1 {
		t.Fatalf("folder requests = %d", len(folder.Requests))
	}
	login := folder.Requests[0].Request
	if login.Method != "POST" || !hasHeader(login.Headers, "Accept", "application/json") {
		t.Fatalf("login = %s %s %v", login.Method, login.URL, login.Headers)
	}
	if hasHeader(login.Headers, "X-Drop", "") {
		t.Fatal("disabled header must be skipped")
	}
	if login.Body != `{"u":"a"}` || !hasHeader(login.Headers, "Content-Type", "application/json") {
		t.Fatalf("body=%q headers=%v", login.Body, login.Headers)
	}
	if login.Auth.Type != "basic" || login.Auth.Config["password"] != "p" {
		t.Fatalf("auth = %+v", login.Auth)
	}

	// Orphaned request (parent missing) lands at root.
	if len(res.Root.Requests) != 1 || res.Root.Requests[0].Request.URL != "/orphan" {
		t.Fatalf("root requests = %+v", res.Root.Requests)
	}

	// Environment flattened: nested map → dotted key + warning; number → string.
	if len(res.Environments) != 1 {
		t.Fatalf("environments = %d", len(res.Environments))
	}
	env := res.Environments[0]
	if env.Variables["base_url"] != "https://api.test" {
		t.Fatalf("vars = %v", env.Variables)
	}
	if env.Variables["user.name"] != "admin" || env.Variables["retries"] != "3" {
		t.Fatalf("flattened vars = %v", env.Variables)
	}
	joined := strings.Join(append(warnings, env.Warnings...), "\n")
	if !strings.Contains(joined, "dotted key") {
		t.Fatalf("expected flattening warning, got %s", joined)
	}
}

func TestParseInsomniaV5YAML(t *testing.T) {
	data := fixtureBytes(t, filepath.Join("insomnia", "fixtures", "insomnia-v5.yaml"))
	res, warnings, err := ParseInsomnia(data)
	if err != nil {
		t.Fatal(err)
	}
	if res.Title != "Test API Collection v5" {
		t.Fatalf("title = %q", res.Title)
	}
	// collection: [API Tests {children}, Data Management {children}]
	if len(res.Root.Folders) != 2 {
		t.Fatalf("root folders = %d", len(res.Root.Folders))
	}
	api := res.Root.Folders[0]
	if api.Name != "API Tests" || len(api.Folders) != 1 || len(api.Requests) < 1 {
		t.Fatalf("tree = %+v / %d requests", api.Name, len(api.Requests))
	}
	authFolder := api.Folders[0]
	if authFolder.Name != "Authentication" || len(authFolder.Requests) != 1 {
		t.Fatalf("auth folder = %+v", authFolder)
	}
	loginReq := authFolder.Requests[0].Request
	if loginReq.Method != "POST" || !strings.Contains(loginReq.URL, "auth/login") {
		t.Fatalf("login request = %s %s", loginReq.Method, loginReq.URL)
	}
	bodyWarnFree := strings.Join(warnings, "\n")
	_ = bodyWarnFree
}

func TestParseInsomniaEnvironmentsToFiles(t *testing.T) {
	data := fixtureBytes(t, filepath.Join("insomnia", "fixtures", "insomnia-v4-with-envs.json"))
	res, _, err := ParseInsomnia(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Environments) != 3 {
		t.Fatalf("environments = %d, want 3", len(res.Environments))
	}
	dir := t.TempDir()
	if err := res.Write(dir); err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "environments", "*.yaml"))
	if len(matches) != 3 {
		t.Fatalf("env files = %v", matches)
	}
	ws, err := collections.LoadWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Config.Name == "" {
		t.Fatal("workspace name empty")
	}
}

func TestParseInsomniaInvalidInputs(t *testing.T) {
	if _, _, err := ParseInsomnia([]byte(insomniaV4Inline[:40])); err == nil {
		t.Fatal("truncated JSON must error")
	}
	bad := fixtureBytes(t, filepath.Join("insomnia", "fixtures", "insomnia-malformed.json"))
	if _, _, err := ParseInsomnia(bad); err == nil {
		t.Fatal("malformed v4 JSON must error")
	}
	missingCollection := fixtureBytes(t, filepath.Join("insomnia", "fixtures", "insomnia-v5-invalid-missing-collection.yaml"))
	res, _, err := ParseInsomnia(missingCollection)
	if err == nil {
		t.Fatalf("missing collection block must error, got %+v", res)
	}
	if !strings.Contains(err.Error(), "collection") {
		t.Fatalf("error should mention collection: %v", err)
	}
}

func TestParseInsomniaDatesFixture(t *testing.T) {
	data := fixtureBytes(t, filepath.Join("insomnia", "fixtures", "insomnia-v5-dates.yaml"))
	res, warnings, err := ParseInsomnia(data)
	if err != nil {
		t.Fatal(err)
	}
	if res.RequestCount() == 0 && len(res.Root.Folders) == 0 {
		t.Fatal("dates fixture imported nothing")
	}
	_ = warnings
}

func TestParseInsomniaSniffingAndEmpty(t *testing.T) {
	if _, _, err := ParseInsomnia(nil); err == nil {
		t.Fatal("nil input must error")
	}
	if _, _, err := ParseInsomnia([]byte("   \n")); err == nil {
		t.Fatal("blank input must error")
	}
}
