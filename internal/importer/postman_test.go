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
	"github.com/Its-Satyajit/reqly/internal/request"
)

const postmanV21 = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"

func postmanCollection(t string) string {
	return `{
  "info": { "name": "Demo API", "schema": "` + t + `" },
  "variable": [
    { "key": "baseUrl", "value": "https://api.example.com" },
    { "key": "off", "value": "ignored", "disabled": true }
  ],
  "item": [
    {
      "name": "users",
      "item": [
        {
          "name": "list users",
          "request": {
            "method": "GET",
            "url": { "raw": "{{baseUrl}}/users?page=1", "host": ["{{baseUrl}}"], "path": ["users"],
              "query": [ { "key": "page", "value": "1" }, { "key": "old", "value": "x", "disabled": true } ] }
          }
        },
        {
          "name": "create user",
          "request": {
            "method": "POST",
            "url": "{{baseUrl}}/users",
            "header": [ { "key": "Accept", "value": "application/json" }, { "key": "X-Drop", "value": "y", "disabled": true } ],
            "body": { "mode": "raw", "raw": "{\"name\":\"ada\"}", "options": { "raw": { "language": "json" } } }
          },
          "event": [ { "listen": "test", "script": { "exec": ["pm.test(\"x\", () => true);"] } } ]
        }
      ]
    },
    {
      "name": "auth",
      "request": {
        "method": "POST",
        "url": "{{baseUrl}}/login",
        "body": { "mode": "urlencoded", "urlencoded": [ { "key": "user", "value": "a b" }, { "key": "skip", "value": "z", "disabled": true } ] },
        "auth": { "type": "basic", "basic": { "username": "u", "password": "p" } }
      }
    },
    {
      "name": "upload",
      "request": {
        "method": "POST",
        "url": "{{baseUrl}}/upload",
        "body": { "mode": "formdata", "formdata": [ { "key": "note", "value": "hi", "type": "text" }, { "key": "doc", "type": "file", "src": "/tmp/a.pdf" } ] }
      }
    },
    {
      "name": "graphql",
      "request": {
        "method": "POST",
        "url": "{{baseUrl}}/graphql",
        "body": { "mode": "graphql", "graphql": { "query": "{ users { id } }", "variables": "{\"first\":5}" } }
      }
    },
    {
      "name": "binary file",
      "request": {
        "method": "POST",
        "url": "{{baseUrl}}/bin",
        "body": { "mode": "file", "file": { "src": "/tmp/big.bin" } }
      }
    }
  ]
}`
}

func TestParsePostmanV21(t *testing.T) {
	res, warnings, err := ParsePostman([]byte(postmanCollection(postmanV21)))
	if err != nil {
		t.Fatal(err)
	}
	if res.Title != "Demo API" {
		t.Fatalf("title = %q", res.Title)
	}
	if res.Variables["baseUrl"] != "https://api.example.com" {
		t.Fatalf("variables = %v", res.Variables)
	}
	if _, ok := res.Variables["off"]; ok {
		t.Fatal("disabled collection variable must be skipped")
	}
	if len(res.Root.Folders) != 1 || res.Root.Folders[0].Name != "users" {
		t.Fatalf("folders = %+v", res.Root.Folders)
	}
	users := res.Root.Folders[0]
	if len(users.Requests) != 2 {
		t.Fatalf("users requests = %d", len(users.Requests))
	}
	if len(res.Root.Requests) != 4 {
		t.Fatalf("root requests = %d, want 4", len(res.Root.Requests))
	}

	list := users.Requests[0].Request
	if list.Method != "GET" || list.URL != "{{baseUrl}}/users?page=1" {
		t.Fatalf("list = %s %s", list.Method, list.URL)
	}
	if len(list.Query) != 1 || list.Query[0].Key != "page" {
		t.Fatalf("query = %v (disabled must be skipped)", list.Query)
	}

	create := users.Requests[1].Request
	if create.Body != `{"name":"ada"}` {
		t.Fatalf("raw body = %q", create.Body)
	}
	if !hasHeader(create.Headers, "Content-Type", "application/json") {
		t.Fatalf("implied Content-Type missing: %v", create.Headers)
	}
	if hasHeader(create.Headers, "X-Drop", "") {
		t.Fatal("disabled header must be skipped")
	}

	login := res.Root.Requests[0].Request
	if login.Body != "user=a+b" && login.Body != "user=a%20b" {
		t.Fatalf("urlencoded body = %q", login.Body)
	}
	if !hasHeader(login.Headers, "Content-Type", "application/x-www-form-urlencoded") {
		t.Fatalf("urlencoded Content-Type missing: %v", login.Headers)
	}
	if login.Auth.Type != "basic" || login.Auth.Config["username"] != "u" || login.Auth.Config["password"] != "p" {
		t.Fatalf("basic auth = %+v", login.Auth)
	}

	gql := res.Root.Requests[2].Request
	if !strings.Contains(gql.Body, `"query"`) || !strings.Contains(gql.Body, `{ users { id } }`) {
		t.Fatalf("graphql body = %q", gql.Body)
	}

	joined := strings.Join(warnings, "\n")
	for _, want := range []string{"script not imported", "file field", "local path", "file mode"} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings missing %q:\n%s", want, joined)
		}
	}
}

func TestParsePostmanAuthMappingAndOverrides(t *testing.T) {
	data := `{
  "info": { "name": "A", "schema": "` + postmanV21 + `" },
  "auth": { "type": "bearer", "bearer": { "token": "t0" } },
  "item": [
    { "name": "plain", "request": { "method": "GET", "url": "https://x.test/" } },
    { "name": "own", "request": { "method": "GET", "url": "https://x.test/",
        "auth": { "type": "apikey", "apikey": [ { "key": "key", "value": "k" }, { "key": "value", "value": "v" }, { "key": "in", "value": "header" } ] } } },
    { "name": "hawk", "request": { "method": "GET", "url": "https://x.test/", "auth": { "type": "hawk", "hawk": {} } } }
  ]
}`
	res, warnings, err := ParsePostman([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	plain := res.Root.Requests[0].Request
	if plain.Auth.Type != "bearer" || plain.Auth.Config["token"] != "t0" {
		t.Fatalf("inherited auth = %+v", plain.Auth)
	}
	own := res.Root.Requests[1].Request
	if own.Auth.Type != "apikey" || own.Auth.Config["in"] != "header" {
		t.Fatalf("request auth override = %+v", own.Auth)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "hawk") {
		t.Fatalf("unsupported auth type must warn, got %v", warnings)
	}
}

func TestParsePostmanRejectsNonJSONAndOldSchema(t *testing.T) {
	if _, _, err := ParsePostman([]byte("not json")); err == nil {
		t.Fatal("expected error for non-JSON input")
	}
	old := `{"info": {"name": "old", "schema": "https://schema.getpostman.com/json/collection/v2.0.0/collection.json"}, "item": []}`
	res, warnings, err := ParsePostman([]byte(old))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "v2.0") {
		t.Fatalf("expected schema-version warning, got %v", warnings)
	}
	if res.Title != "old" {
		t.Fatalf("title = %q", res.Title)
	}
}

func hasHeader(headers []request.Header, key, value string) bool {
	for _, h := range headers {
		if strings.EqualFold(h.Key, key) && (value == "" || h.Value == value) {
			return true
		}
	}
	return false
}

func TestPostmanWriteRoundTrip(t *testing.T) {
	res, _, err := ParsePostman([]byte(postmanCollection(postmanV21)))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := res.Write(dir); err != nil {
		t.Fatal(err)
	}

	ws, err := collections.LoadWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Config.Name != "Demo API" {
		t.Fatalf("workspace name = %q", ws.Config.Name)
	}
	if len(ws.Collections) != 1 {
		t.Fatalf("collections = %d", len(ws.Collections))
	}
	coll := ws.Collections[0]
	if coll.Config.Variables["baseUrl"] != "https://api.example.com" {
		t.Fatalf("collection vars = %v", coll.Config.Variables)
	}
	if len(coll.Folders) != 1 || len(coll.Folders[0].Requests) != 2 {
		t.Fatalf("tree = %d folders, %d root requests", len(coll.Folders), len(coll.Requests))
	}
	if len(coll.Requests) != 4 {
		t.Fatalf("root requests = %d, want 4", len(coll.Requests))
	}
}

func TestPostmanWriteDedupesFilenames(t *testing.T) {
	data := `{
  "info": { "name": "dups", "schema": "` + postmanV21 + `" },
  "item": [
    { "name": "same name!", "request": { "method": "GET", "url": "https://a.test/1" } },
    { "name": "same name!", "request": { "method": "GET", "url": "https://a.test/2" } }
  ]
}`
	res, _, err := ParsePostman([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := res.Write(dir); err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "collections", "postman-import", "*.yaml"))
	if len(matches) != 3 { // 2 requests + reqly.yaml
		t.Fatalf("files = %v", matches)
	}
	ws, err := collections.LoadWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := ws.Collections[0].Requests; len(got) != 2 {
		t.Fatalf("requests after reload = %d, want 2", len(got))
	}
}

// TestParsePostmanOfficialExamples imports the example collections published
// in postmanlabs/postman-collection (Apache-2.0), vendored under
// testdata/postman, and asserts each yields a loadable workspace.
func TestParsePostmanOfficialExamples(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("testdata", "postman", "*.json"))
	if err != nil || len(matches) == 0 {
		t.Fatalf("fixtures missing: %v %v", matches, err)
	}
	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			res, _, err := ParsePostman(data)
			if err != nil {
				t.Fatal(err)
			}
			if res.Root == nil {
				t.Fatal("nil root")
			}
			dir := t.TempDir()
			if err := res.Write(dir); err != nil {
				t.Fatal(err)
			}
			ws, err := collections.LoadWorkspace(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(ws.Collections) != 1 {
				t.Fatalf("collections = %d, want 1", len(ws.Collections))
			}
			if countRequests(res.Root) == 0 {
				t.Fatal("expected at least one imported request")
			}
		})
	}
}

func countRequests(f *PostmanFolder) int {
	n := len(f.Requests)
	for _, sub := range f.Folders {
		n += countRequests(sub)
	}
	return n
}
