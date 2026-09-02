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

package exporter

import (
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
)

func TestExportToPostman(t *testing.T) {
	reqs := []request.Request{
		{
			Name:   "List Users",
			Method: request.MethodGet,
			URL:    "https://api.example.com/users?page=1",
			Headers: []request.Header{
				{Key: "Accept", Value: "application/json"},
			},
			Query: []request.Parameter{{Key: "page", Value: "1"}},
		},
		{
			Method: request.MethodPost,
			URL:    "https://api.example.com/users",
			Body:   `{"name":"x"}`,
		},
	}

	col, err := ExportToPostman("My API", reqs)
	if err != nil {
		t.Fatal(err)
	}
	if col.Info.Name != "My API" {
		t.Fatalf("name: got %q", col.Info.Name)
	}
	if len(col.Item) != 2 {
		t.Fatalf("items: got %d", len(col.Item))
	}

	first := col.Item[0]
	if first.Name != "List Users" {
		t.Fatalf("first name: got %q", first.Name)
	}
	if first.Request.Method != "GET" {
		t.Fatalf("first method: got %q", first.Request.Method)
	}
	if first.Request.URL.Raw != "https://api.example.com/users?page=1" {
		t.Fatalf("first url: got %q", first.Request.URL.Raw)
	}
	if len(first.Request.URL.Query) != 1 || first.Request.URL.Query[0].Key != "page" {
		t.Fatalf("first query: got %+v", first.Request.URL.Query)
	}

	second := col.Item[1]
	if second.Name != "POST https://api.example.com/users" {
		t.Fatalf("second name: got %q", second.Name)
	}
	if second.Request.Body == nil || second.Request.Body.Raw != `{"name":"x"}` {
		t.Fatalf("second body: got %+v", second.Request.Body)
	}
}

func TestExportToPostmanJSONValid(t *testing.T) {
	data, err := ExportToPostmanJSON("X", []request.Request{
		{Method: request.MethodGet, URL: "https://api.example.com/"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"schema"`) {
		t.Fatalf("output missing schema:\n%s", data)
	}

	// Round-trip.
	col, err := ParsePostman(data)
	if err != nil {
		t.Fatal(err)
	}
	if col.Info.Name != "X" {
		t.Fatalf("round-trip name: got %q", col.Info.Name)
	}
	if len(col.Item) != 1 {
		t.Fatalf("round-trip items: got %d", len(col.Item))
	}
}

func TestParsePostmanNotACollection(t *testing.T) {
	if _, err := ParsePostman([]byte(`{"foo": 1}`)); err == nil {
		t.Fatal("expected error for non-Postman JSON")
	}
}

func TestPostmanToRequestsWithFoldersAndBase(t *testing.T) {
	data := `{
		"info": {"name": "C", "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},
		"item": [
			{"name": "Ping", "request": {"method": "GET", "url": {"raw": "/ping"}}},
			{"name": "Folder", "item": [
				{"name": "Create", "request": {"method": "POST", "url": {"raw": "/items"}, "body": {"mode": "raw", "raw": "{}"}}}
			]}
		]
	}`
	col, err := ParsePostman([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	reqs := col.ToRequests("https://api.example.com/v1")
	if len(reqs) != 2 {
		t.Fatalf("requests: got %d", len(reqs))
	}
	if reqs[0].URL != "https://api.example.com/v1/ping" {
		t.Fatalf("ping url: got %q", reqs[0].URL)
	}
	if reqs[0].Name != "Ping" {
		t.Fatalf("ping name: got %q", reqs[0].Name)
	}
	if reqs[1].URL != "https://api.example.com/v1/items" {
		t.Fatalf("create url: got %q", reqs[1].URL)
	}
	if reqs[1].Method != request.MethodPost || reqs[1].Body != "{}" {
		t.Fatalf("create req: got %+v", reqs[1])
	}
}

func TestDisplayName(t *testing.T) {
	cases := []struct {
		req  request.Request
		want string
	}{
		{request.Request{Name: "Named", Method: request.MethodGet}, "Named"},
		{request.Request{Method: request.MethodGet, URL: "https://x.com/"}, "GET https://x.com/"},
		{request.Request{Method: request.MethodDelete}, "DELETE"},
	}
	for _, c := range cases {
		if got := displayName(c.req); got != c.want {
			t.Fatalf("displayName(%+v): got %q, want %q", c.req, got, c.want)
		}
	}
}
