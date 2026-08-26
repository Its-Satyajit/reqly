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

package openapi

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

const exploreSpec = `openapi: 3.0.3
info:
  title: Explore API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /users/{id}:
    get:
      operationId: getUser
      summary: Get a user
      tags: [users]
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: ok
    delete:
      operationId: deleteUser
      tags: [users, admin]
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "204":
          description: gone
  /ping:
    get:
      summary: Ping
      responses:
        "200":
          description: pong
`

func mustDoc(t *testing.T, data string) *openapi3.T {
	t.Helper()
	doc, err := Load([]byte(data))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return doc
}

func TestExploreListsOperationsSorted(t *testing.T) {
	eps := Explore(mustDoc(t, exploreSpec))
	if len(eps) != 3 {
		t.Fatalf("Explore() returned %d endpoints, want 3", len(eps))
	}
	want := []struct{ method, path, opID, summary, firstTag string }{
		{"GET", "/ping", "", "Ping", ""},
		{"GET", "/users/{id}", "getUser", "Get a user", "users"},
		{"DELETE", "/users/{id}", "deleteUser", "", "users"},
	}
	for i, w := range want {
		got := eps[i]
		if got.Method != w.method || got.Path != w.path || got.OperationID != w.opID || got.Summary != w.summary {
			t.Errorf("endpoint[%d] = %s %s opID=%q summary=%q, want %s %s opID=%q summary=%q",
				i, got.Method, got.Path, got.OperationID, got.Summary, w.method, w.path, w.opID, w.summary)
		}
		firstTag := ""
		if len(got.Tags) > 0 {
			firstTag = got.Tags[0]
		}
		if firstTag != w.firstTag {
			t.Errorf("endpoint[%d] first tag = %q, want %q", i, firstTag, w.firstTag)
		}
	}
}

func TestFilterByTag(t *testing.T) {
	users := FilterByTag(Explore(mustDoc(t, exploreSpec)), "users")
	if len(users) != 2 {
		t.Fatalf("FilterByTag(users) returned %d endpoints, want 2", len(users))
	}
	if users[0].OperationID != "getUser" || users[1].OperationID != "deleteUser" {
		t.Fatalf("unexpected filter result: %+v", users)
	}
	if got := FilterByTag(Explore(mustDoc(t, exploreSpec)), "nonexistent"); len(got) != 0 {
		t.Fatalf("FilterByTag(nonexistent) returned %d endpoints, want 0", len(got))
	}
}

func TestGenerateNoSelectorsErrorsListingOps(t *testing.T) {
	_, _, err := Generate(mustDoc(t, exploreSpec), GenerateOptions{})
	if err == nil {
		t.Fatal("expected error when no selectors given")
	}
	for _, want := range []string{"getUser", "--operation", "--all"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestGenerateUnknownOperationErrors(t *testing.T) {
	_, _, err := Generate(mustDoc(t, exploreSpec), GenerateOptions{Operations: []string{"nope"}})
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected unknown-operation error mentioning nope, got %v", err)
	}
}

func TestGenerateUnknownMethodPathErrors(t *testing.T) {
	_, _, err := Generate(mustDoc(t, exploreSpec), GenerateOptions{Method: "GET", Path: "/nope"})
	if err == nil || !strings.Contains(err.Error(), "/nope") {
		t.Fatalf("expected unknown-path error mentioning /nope, got %v", err)
	}
}
