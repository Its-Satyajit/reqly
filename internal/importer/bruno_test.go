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

const brunoInline = `{
  "name": "inline-bruno",
  "version": 1,
  "root": {
    "request": {
      "auth": { "mode": "bearer", "bearer": { "token": "{{tok}}" } },
      "headers": [ { "name": "X-Default", "value": "yes", "enabled": true } ]
    }
  },
  "items": [
    {
      "type": "folder",
      "name": "api",
      "items": [
        {
          "type": "http",
          "name": "Create",
          "request": {
            "url": "{{host}}/things",
            "method": "POST",
            "headers": [
              { "name": "Accept", "value": "application/json", "enabled": true },
              { "name": "X-Drop", "value": "y", "enabled": false }
            ],
            "params": [ { "name": "page", "value": "2", "enabled": true } ],
            "body": { "mode": "json", "json": "{\"a\":1}" },
            "auth": { "mode": "apikey", "apikey": { "key": "k", "value": "v", "placement": "queryparams" } }
          }
        }
      ]
    },
    {
      "type": "http",
      "name": "Form",
      "request": {
        "url": "{{host}}/form",
        "method": "POST",
        "body": {
          "mode": "formUrlEncoded",
          "formUrlEncoded": [ { "name": "a b", "value": "c d", "enabled": true }, { "name": "off", "value": "x", "enabled": false } ]
        }
      }
    },
    {
      "type": "graphql",
      "name": "gql",
      "request": {
        "url": "{{host}}/graphql",
        "method": "POST",
        "body": { "mode": "graphql", "graphql": { "query": "{ q }", "variables": "{\"n\":1}" } }
      }
    },
    {
      "type": "http",
      "name": "Scripted",
      "request": {
        "url": "{{host}}/s",
        "method": "GET",
        "script": { "request": "console.log(1);" },
        "assertions": [ { "name": "res.status", "value": "200" } ],
        "docs": "some docs"
      }
    }
  ],
  "environments": [
    {
      "name": "Local",
      "variables": [
        { "name": "host", "value": "http://localhost:8080", "enabled": true, "secret": false },
        { "name": "api_key", "value": "s3cret", "enabled": true, "secret": true },
        { "name": "disabled_var", "value": "x", "enabled": false, "secret": false }
      ]
    }
  ]
}`

func TestParseBruno(t *testing.T) {
	res, report, err := ParseBruno([]byte(brunoInline))
	if err != nil {
		t.Fatal(err)
	}
	if res.Title != "inline-bruno" {
		t.Fatalf("title = %q", res.Title)
	}
	if len(res.Root.Folders) != 1 || res.Root.Folders[0].Name != "api" {
		t.Fatalf("folders = %+v", res.Root.Folders)
	}
	create := res.Root.Folders[0].Requests[0].Request
	if create.Method != "POST" || !hasHeader(create.Headers, "Content-Type", "application/json") {
		t.Fatalf("create = %s %v", create.Method, create.Headers)
	}
	if hasHeader(create.Headers, "X-Drop", "") {
		t.Fatal("disabled header must be skipped")
	}
	if create.Body != `{"a":1}` {
		t.Fatalf("body = %q", create.Body)
	}
	if create.Auth.Type != "apikey" || create.Auth.Config["in"] != "query" {
		t.Fatalf("apikey auth = %+v", create.Auth)
	}

	form := res.Root.Requests[0].Request
	if form.Body != "a+b=c+d" && form.Body != "a+b=c%20d" {
		t.Fatalf("urlencoded body = %q", form.Body)
	}

	gql := res.Root.Requests[1].Request
	if !strings.Contains(gql.Body, `"query"`) || !strings.Contains(gql.Body, `{ q }`) {
		t.Fatalf("graphql body = %q", gql.Body)
	}

	scripted := res.Root.Requests[2]
	if scripted.Request.URL != "{{host}}/s" || scripted.Request.Auth.Type != "" && scripted.Request.Auth.Type != "bearer" {
		// Scripted request inherits the collection-level bearer auth.
		if scripted.Request.Auth.Type != "bearer" {
			t.Fatalf("scripted auth = %+v (want inherited bearer)", scripted.Request.Auth)
		}
	}

	joined := strings.Join(report.Messages(), "\n")
	for _, want := range []string{"script not imported", "assertions not imported", "docs not imported"} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings missing %q:\n%s", want, joined)
		}
	}

	if len(res.Environments) != 1 {
		t.Fatalf("environments = %d", len(res.Environments))
	}
	env := res.Environments[0]
	if env.Variables["host"] != "http://localhost:8080" {
		t.Fatalf("vars = %v", env.Variables)
	}
	if env.Secrets["api_key"] != "s3cret" {
		t.Fatalf("secrets = %v", env.Secrets)
	}
	if _, ok := env.Variables["disabled_var"]; ok {
		t.Fatal("disabled variable must be skipped")
	}
}

func TestParseBrunoCollectionDefaults(t *testing.T) {
	res, report, err := ParseBruno([]byte(brunoInline))
	if err != nil {
		t.Fatal(err)
	}
	if res.Auth.Type != "bearer" || res.Auth.Config["token"] != "{{tok}}" {
		t.Fatalf("collection auth = %+v", res.Auth)
	}
	if len(res.Headers) != 1 || res.Headers[0].Key != "X-Default" {
		t.Fatalf("collection headers = %v", res.Headers)
	}
	_ = report
}

func TestParseBrunoWriteRoundTrip(t *testing.T) {
	res, _, err := ParseBruno([]byte(brunoInline))
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
	if ws.Config.Name != "inline-bruno" || len(ws.Collections) != 1 {
		t.Fatalf("workspace = %+v", ws.Config.Name)
	}
	coll := ws.Collections[0]
	if coll.Config.Auth.Type != "bearer" || coll.Config.Auth.Config["token"] != "{{tok}}" {
		t.Fatalf("descriptor auth = %+v", coll.Config.Auth)
	}
	if len(coll.Config.Headers) != 1 || coll.Config.Headers[0].Key != "X-Default" {
		t.Fatalf("descriptor headers = %v", coll.Config.Headers)
	}
	envFiles, _ := filepath.Glob(filepath.Join(dir, "environments", "*.yaml"))
	if len(envFiles) != 1 {
		t.Fatalf("env files = %v", envFiles)
	}
	envData, _ := os.ReadFile(envFiles[0])
	if !strings.Contains(string(envData), "secrets:") || !strings.Contains(string(envData), "api_key") {
		t.Fatalf("env file missing secrets block: %s", envData)
	}
}

func TestParseBrunoFixtures(t *testing.T) {
	type expectation int
	const (
		mustParse expectation = iota
		mustError
	)
	cases := map[string]expectation{
		"bruno-testbench.json":                        mustParse,
		"bruno-v2-json-collection-with-proxy.json":    mustParse,
		"bruno-with-examples.json":                    mustParse,
		"bruno-http-request-missing-url.json":         mustParse,
		"bruno-http-example-request-missing-url.json": mustParse,
		"bruno-grpc-request-missing-url.json":         mustParse,
		"descriptions-collection-bru.json":            mustParse,
		"bruno-invalid-corrupted.json":                mustParse,
		"bruno-malformed.json":                        mustError,
		"bruno-missing-required-fields.json":          mustParse,
	}
	names, err := filepath.Glob(filepath.Join("testdata", "import-suite", "bruno", "fixtures", "*.json"))
	if err != nil || len(names) == 0 {
		t.Fatalf("fixtures missing: %v %v", names, err)
	}
	seen := map[string]bool{}
	for _, path := range names {
		base := filepath.Base(path)
		want, ok := cases[base]
		if !ok {
			t.Errorf("fixture %s has no expectation entry", base)
			continue
		}
		seen[base] = true
		t.Run(base, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			res, _, err := ParseBruno(data)
			if want == mustError {
				if err == nil {
					t.Fatalf("expected error, got %+v", res)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			if err := res.Write(dir); err != nil {
				t.Fatal(err)
			}
			if _, err := collections.LoadWorkspace(dir); err != nil {
				t.Fatal(err)
			}
		})
	}
	for name := range cases {
		if !seen[name] {
			t.Errorf("expected fixture %s not found", name)
		}
	}
}
