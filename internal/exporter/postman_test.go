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
