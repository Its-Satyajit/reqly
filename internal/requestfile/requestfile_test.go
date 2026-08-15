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

package requestfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseJSON(t *testing.T) {
	f, err := Parse([]byte(`{
		"name": "users",
		"variables": {"token": "abc"},
		"request": {
			"method": "GET",
			"url": "https://api.example.com/users?page={{page}}",
			"headers": [{"key": "Accept", "value": "application/json"}],
			"query": [{"key": "page", "value": "2"}],
			"body": "{\"a\":1}",
			"timeout": 5000
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Name != "users" {
		t.Fatalf("name: got %q", f.Name)
	}
	if f.Request.Method != "GET" || f.Request.URL == "" {
		t.Fatalf("unexpected request: %+v", f.Request)
	}
	if v, ok := f.Variable("token"); !ok || v != "abc" {
		t.Fatalf("variable token: got %q, %v", v, ok)
	}
	if got := f.VariableNames(); len(got) != 1 || got[0] != "token" {
		t.Fatalf("variable names: got %v", got)
	}
}

func TestParseYAML(t *testing.T) {
	f, err := Parse([]byte(`
name: users
variables:
  token: abc
  host: api.example.com
request:
  method: POST
  url: https://{{host}}/users
  headers:
    - key: Content-Type
      value: application/json
  body: '{"name":"reqly"}'
`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Name != "users" {
		t.Fatalf("name: got %q", f.Name)
	}
	if f.Request.Method != "POST" {
		t.Fatalf("method: got %q", f.Request.Method)
	}
	if len(f.Request.Headers) != 1 || f.Request.Headers[0].Key != "Content-Type" {
		t.Fatalf("headers: got %+v", f.Request.Headers)
	}
	if got := f.VariableNames(); len(got) != 2 || got[0] != "host" || got[1] != "token" {
		t.Fatalf("variable names: got %v", got)
	}
}

func TestParseRequiresURL(t *testing.T) {
	if _, err := Parse([]byte(`{"request": {"method": "GET"}}`)); err == nil {
		t.Fatal("expected error for missing url")
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse([]byte(`not yaml [ {`)); err == nil {
		t.Fatal("expected error for invalid content")
	}
}

func TestParseEmpty(t *testing.T) {
	if _, err := Parse([]byte(`{}`)); err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestLoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "req.json")
	if err := os.WriteFile(path, []byte(`{"request": {"method": "GET", "url": "https://api.example.com/x"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Request.URL != "https://api.example.com/x" {
		t.Fatalf("url: got %q", f.Request.URL)
	}
}

func TestLoadFileMissing(t *testing.T) {
	if _, err := LoadFile(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLooksLikeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "req.yaml")
	os.WriteFile(path, []byte("{}"), 0o644)
	if !LooksLikeFile(path) {
		t.Fatalf("expected %q to look like a file", path)
	}
	for _, url := range []string{"https://api.example.com/x", "http://localhost:8000"} {
		if LooksLikeFile(url) {
			t.Fatalf("expected %q to not look like a file", url)
		}
	}
	if LooksLikeFile(filepath.Join(t.TempDir(), "missing.txt")) {
		t.Fatal("expected missing non-request path to not look like a file")
	}
	if !LooksLikeFile(filepath.Join(t.TempDir(), "missing.yaml")) {
		t.Fatal("expected missing .yaml path to look like a file")
	}
}
