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

package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	reqtesting "github.com/Its-Satyajit/reqly/internal/testing"
)

func writeTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func testFileForURL(url string) string {
	return fmt.Sprintf(`{
		"name": "suite",
		"request": {"method": "GET", "url": %q},
		"tests": [
			{"name": "ok", "assertions": [
				{"kind": "status", "expected": 200},
				{"kind": "json", "path": "$.id", "exact": true, "value": "42"}
			]}
		]
	}`, url)
}

func TestTestCommandPasses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":42}`))
	}))
	defer srv.Close()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"test", writeTestFile(t, testFileForURL(srv.URL))})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !strings.Contains(out.String(), "1/1 tests passed") {
		t.Fatalf("expected pass summary, got:\n%s", out.String())
	}
}

func TestTestCommandFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"id":99}`))
	}))
	defer srv.Close()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"test", writeTestFile(t, testFileForURL(srv.URL))})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(out.String(), "0/1 tests passed") {
		t.Fatalf("expected fail summary, got:\n%s", out.String())
	}
}

func TestTestCommandPartialPass(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":7}`))
	}))
	defer srv.Close()

	content := fmt.Sprintf(`{
		"name": "suite",
		"request": {"method": "GET", "url": %q},
		"tests": [
			{"name": "first", "assertions": [{"kind": "status", "expected": 200}]},
			{"name": "second", "assertions": [{"kind": "status", "expected": 404}]}
		]
	}`, srv.URL)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"test", writeTestFile(t, content)})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected failure for partial pass")
	}
	if !strings.Contains(out.String(), "1/2 tests passed") {
		t.Fatalf("expected 1/2 summary, got:\n%s", out.String())
	}
}

func TestTestCommandMasksSecretsInAssertionMessages(t *testing.T) {
	secret := "top-secret-value"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"api_key":"` + secret + `"}`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	writeEnv(t, dir, "dev", "secrets:\n  API_KEY: "+secret+"\n")
	envDir := filepath.Join(dir, "environments")
	t.Chdir(dir)

	content := fmt.Sprintf(`{
		"name": "suite",
		"environment": "dev",
		"request": {"method": "GET", "url": %q},
		"tests": [
			{"name": "secret-echo", "assertions": [
				{"kind": "status", "expected": 200},
				{"kind": "json", "path": "$.api_key", "exact": true, "value": "wrong-value"}
			]}
		]
	}`, srv.URL)
	testPath := filepath.Join(dir, "test.json")
	if err := os.WriteFile(testPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = envDir

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"test", testPath})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected failure for failing assertion")
	}
	if strings.Contains(out.String(), secret) {
		t.Fatalf("secret leaked in test output:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "[SECRET]") {
		t.Fatalf("expected [SECRET] masking:\n%s", out.String())
	}
}

func TestTestCommandMissingFile(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"test", "/nonexistent/test.json"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestTestCommandYAMLWithVariables(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	content := fmt.Sprintf(`
name: yaml suite
variables:
  token: abc123
request:
  method: GET
  url: %s
  headers:
    - key: Authorization
      value: Bearer {{token}}
tests:
  - name: ok
    assertions:
      - kind: status
        expected: 200
`, srv.URL)

	path := filepath.Join(t.TempDir(), "test.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"test", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if gotAuth != "Bearer abc123" {
		t.Fatalf("variable not interpolated: auth header %q", gotAuth)
	}
	if !strings.Contains(out.String(), "1/1 tests passed") {
		t.Fatalf("expected pass summary, got:\n%s", out.String())
	}
}

func TestTestCommandRequestError(t *testing.T) {
	content := fmt.Sprintf(`{
		"name": "suite",
		"request": {"method": "GET", "url": "http://127.0.0.1:1"},
		"tests": [{"name": "t", "assertions": [{"kind": "status", "expected": 200}]}]
	}`)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"test", writeTestFile(t, content)})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestCountPassed(t *testing.T) {
	results := []reqtesting.TestResult{
		{Passed: true},
		{Passed: false},
		{Passed: true},
	}
	if got := countPassed(results); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}
