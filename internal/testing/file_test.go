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

package testing

import (
	"strings"
	"testing"
)

func TestParseTestFile(t *testing.T) {
	data := `{
		"name": "users",
		"request": {
			"method": "GET",
			"url": "https://api.example.com/users",
			"timeout": 5000,
			"auth": {"type": "bearer", "config": {"token": "abc"}}
		},
		"tests": [
			{
				"name": "ok",
				"assertions": [
					{"kind": "status", "expected": 200},
					{"kind": "json", "path": "$.count", "exact": true, "value": "2"}
				]
			}
		]
	}`

	tf, err := ParseTestFile([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if tf.Name != "users" {
		t.Fatalf("expected name users, got %q", tf.Name)
	}
	if tf.Request.URL != "https://api.example.com/users" {
		t.Fatalf("unexpected url %q", tf.Request.URL)
	}
	if tf.Request.Timeout != 5000 {
		t.Fatalf("expected timeout 5000, got %d", tf.Request.Timeout)
	}
	if tf.Request.Auth.Type != "bearer" || tf.Request.Auth.Config["token"] != "abc" {
		t.Fatalf("unexpected auth %+v", tf.Request.Auth)
	}
	if len(tf.Tests) != 1 || len(tf.Tests[0].Assertions) != 2 {
		t.Fatalf("unexpected tests %+v", tf.Tests)
	}
	if tf.Tests[0].Assertions[0].Kind != AssertStatus {
		t.Fatalf("expected status assertion, got %+v", tf.Tests[0].Assertions[0])
	}
}

func TestParseTestFileMissingURL(t *testing.T) {
	data := `{"name":"x","request":{"method":"GET"},"tests":[{"name":"t","assertions":[]}]}`
	_, err := ParseTestFile([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing url")
	}
}

func TestParseTestFileNoTests(t *testing.T) {
	data := `{"name":"x","request":{"method":"GET","url":"https://x.com"},"tests":[]}`
	_, err := ParseTestFile([]byte(data))
	if err == nil {
		t.Fatal("expected error for no tests")
	}
}

func TestParseTestFileInvalidJSON(t *testing.T) {
	_, err := ParseTestFile([]byte(`{not json`))
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestParseTestFileYAMLWithVariables(t *testing.T) {
	data := `
name: users
variables:
  token: abc123
request:
  method: GET
  url: https://api.example.com/users
  headers:
    - key: Authorization
      value: Bearer {{token}}
tests:
  - name: ok
    assertions:
      - kind: status
        expected: 200
`
	tf, err := ParseTestFile([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if tf.Name != "users" {
		t.Fatalf("name: got %q", tf.Name)
	}
	if v, ok := tf.Variables["token"]; !ok || v != "abc123" {
		t.Fatalf("variable token: got %q, %v", v, ok)
	}
	if tf.Request.Method != "GET" || len(tf.Request.Headers) != 1 {
		t.Fatalf("unexpected request %+v", tf.Request)
	}
	if len(tf.Tests) != 1 {
		t.Fatalf("unexpected tests %+v", tf.Tests)
	}
}

func TestParseTestFileUnknownAssertionStillParses(t *testing.T) {
	data := `{"name":"x","request":{"url":"https://x.com"},"tests":[
		{"name":"t","assertions":[{"kind":"bogus"}]}
	]}`
	tf, err := ParseTestFile([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(tf.Tests[0].Assertions) != 1 {
		t.Fatalf("expected 1 assertion, got %+v", tf.Tests)
	}
}

func TestParseTestFileRoundTripAuth(t *testing.T) {
	data := `{"request":{"url":"https://x.com","auth":{"type":"basic","config":{"username":"u","password":"p"}}},"tests":[{"name":"t","assertions":[{"kind":"status","expected":200}]}]}`
	tf, err := ParseTestFile([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if tf.Request.Auth.Type != "basic" || tf.Request.Auth.Config["username"] != "u" {
		t.Fatalf("unexpected auth %+v", tf.Request.Auth)
	}
}

func TestLoadTestFileMissing(t *testing.T) {
	_, err := LoadTestFile("/nonexistent/path/reqly.json")
	if err == nil || !strings.Contains(err.Error(), "read test file") {
		t.Fatalf("expected read error, got %v", err)
	}
}
