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
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
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

func TestParseEnvironmentField(t *testing.T) {
	f, err := Parse([]byte(`
environment: prod
variables:
  token: abc
request:
  method: GET
  url: https://api.example.com/users
`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Environment != "prod" {
		t.Fatalf("environment: got %q, want %q", f.Environment, "prod")
	}
}

func TestParseEnvironmentFieldOptional(t *testing.T) {
	f, err := Parse([]byte(`
request:
  method: GET
  url: https://api.example.com/users
`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Environment != "" {
		t.Fatalf("environment: got %q, want empty", f.Environment)
	}
}

func TestSaveJSONPreservesFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.json")
	original := `{"name":"list","request":{"method":"GET","url":"https://api.example.com/users"}}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := Parse([]byte(original))
	if err != nil {
		t.Fatal(err)
	}
	f.Request.Method = "POST"
	if err := Save(path, f); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("saved file is not valid JSON: %v\n%s", err, data)
	}
	req := m["request"].(map[string]any)
	if req["method"] != "POST" || req["url"] != "https://api.example.com/users" {
		t.Fatalf("unexpected saved request: %s", data)
	}
}

func TestSaveYAMLPreservesFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.yaml")
	original := "name: list\nrequest:\n  url: https://api.example.com/users\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := Parse([]byte(original))
	if err != nil {
		t.Fatal(err)
	}
	f.Request.Method = "GET"
	if err := Save(path, f); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "{") || strings.Contains(string(data), "}") {
		t.Fatalf("saved file looks like JSON, expected YAML:\n%s", data)
	}
	if !strings.Contains(string(data), "url: https://api.example.com/users") {
		t.Fatalf("saved YAML missing url:\n%s", data)
	}
}

func TestSaveYAMLOmitsZeroValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.yaml")
	if err := os.WriteFile(path, []byte("request:\n  url: https://x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, &File{Request: request.Request{URL: "https://x"}}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, unwanted := range []string{"id:", "method:", "timeout:", "headers:", "query:", "body:", "auth:"} {
		if strings.Contains(string(data), unwanted) {
			t.Fatalf("saved YAML contains zero-value %q:\n%s", unwanted, data)
		}
	}
}

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.yaml")
	src := `name: list
variables:
  token: abc
environment: staging
preRequest: console.log("pre")
postRequest: reqly.test("ok", true)
request:
  method: GET
  url: https://api.example.com/users?page={{page}}
  headers:
    - key: Accept
      value: application/json
  query:
    - key: page
      value: "2"
  body: "{\"a\":1}"
  auth:
    type: bearer
    config:
      token: "{{token}}"
  timeout: 5000
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(path, orig); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(data)
	if err != nil {
		t.Fatalf("saved file does not parse: %v\n%s", err, data)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Fatalf("round trip mismatch:\n%s", data)
	}
}

func TestSaveClearedAuthOmitsBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.yaml")
	src := `request:
  method: GET
  url: https://x
  auth:
    type: basic
    config:
      username: u
      password: p
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	// Inherit: the draft carries no auth, so the saved file must drop the
	// auth block rather than keep it.
	orig.Request.Auth = request.Auth{}
	if err := Save(path, orig); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, unwanted := range []string{"auth:", "username:", "password:"} {
		if strings.Contains(string(data), unwanted) {
			t.Fatalf("saved file still contains %q after clearing auth:\n%s", unwanted, data)
		}
	}
	got, err := Parse(data)
	if err != nil {
		t.Fatalf("saved file does not parse: %v\n%s", err, data)
	}
	if got.Request.Auth.Type != "" || len(got.Request.Auth.Config) != 0 {
		t.Fatalf("auth not cleared on round trip: %+v", got.Request.Auth)
	}
}

func TestSaveRequiresURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.yaml")
	if err := os.WriteFile(path, []byte("request:\n  url: https://x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, &File{Request: request.Request{Method: "GET"}}); err == nil {
		t.Fatal("expected error saving a request without a url")
	}
}

func TestSaveMissingDirErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nope", "list.yaml")
	if err := Save(path, &File{Request: request.Request{Method: "GET", URL: "https://x"}}); err == nil {
		t.Fatal("expected error saving into a nonexistent directory")
	}
}

func TestFingerprint(t *testing.T) {
	a := Fingerprint([]byte("hello"))
	b := Fingerprint([]byte("hello"))
	c := Fingerprint([]byte("world"))
	if a != b {
		t.Fatal("fingerprint is not stable for identical bytes")
	}
	if a == c {
		t.Fatal("fingerprint collides for different bytes")
	}
	if len(a) != 64 {
		t.Fatalf("fingerprint length: got %d, want 64", len(a))
	}
}
