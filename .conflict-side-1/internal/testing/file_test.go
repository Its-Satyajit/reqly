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

func TestParseTestFileEnvironmentField(t *testing.T) {
	data := `
name: users
environment: staging
request:
  method: GET
  url: https://api.example.com/users
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
	if tf.Environment != "staging" {
		t.Fatalf("environment: got %q, want %q", tf.Environment, "staging")
	}
}

func TestParseTestFile_AssertionAliases(t *testing.T) {
	data := `
name: users
request:
  method: GET
  url: https://api.example.com/users
  headers:
    Accept: application/json
tests:
  - name: aliases
    assertions:
      - kind: header
        header: Content-Type
        value: application/json
      - kind: header
        name: X-Custom-Header
        value: custom
      - kind: response_time
        max: 5000
      - kind: response_time
        threshold: 3000
      - kind: status
        status: 200
`
	tf, err := ParseTestFile([]byte(data))
	if err != nil {
		t.Fatalf("failed to parse test file with aliases: %v", err)
	}
	if len(tf.Tests[0].Assertions) != 5 {
		t.Fatalf("expected 5 assertions, got %d", len(tf.Tests[0].Assertions))
	}
	a0 := tf.Tests[0].Assertions[0]
	if a0.Kind != AssertHeader || a0.Path != "Content-Type" || a0.Value != "application/json" {
		t.Fatalf("assertion 0 alias mismatch: %+v", a0)
	}
	a1 := tf.Tests[0].Assertions[1]
	if a1.Kind != AssertHeader || a1.Path != "X-Custom-Header" || a1.Value != "custom" {
		t.Fatalf("assertion 1 alias mismatch: %+v", a1)
	}
	a2 := tf.Tests[0].Assertions[2]
	if a2.Kind != AssertResponseTime || a2.Expected != 5000 {
		t.Fatalf("assertion 2 alias mismatch: %+v", a2)
	}
	a3 := tf.Tests[0].Assertions[3]
	if a3.Kind != AssertResponseTime || a3.Expected != 3000 {
		t.Fatalf("assertion 3 alias mismatch: %+v", a3)
	}
	a4 := tf.Tests[0].Assertions[4]
	if a4.Kind != AssertStatus || a4.Expected != 200 {
		t.Fatalf("assertion 4 alias mismatch: %+v", a4)
	}
}
