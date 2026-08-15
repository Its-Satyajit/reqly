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

package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCollectionTestRunsAllRequests(t *testing.T) {
	resetCollectionTestFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()

	root := t.TempDir()
	write := func(rel, contents string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("reqly.yaml", `{"baseURL":"`+srv.URL+`"}`)
	write("collections/main/reqly.yaml", `{"name":"main"}`)
	write("collections/main/a.yaml", `
request:
  method: GET
  url: /a
postRequest: |
  reqly.test("status 200", function() { return reqly.response.status === 200; });
`)
	write("collections/main/b.yaml", `
request:
  method: GET
  url: /b
postRequest: |
  reqly.test("has ok", function() { return reqly.response.body.indexOf("ok") !== -1; });
`)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "test", "main", "--workspace", root})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := stripANSI(out.String())
	for _, want := range []string{"PASS a", "PASS b", "2 passed, 0 failed", "[ok] status 200", "[ok] has ok"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestCollectionTestFailingAssertion(t *testing.T) {
	resetCollectionTestFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "unexpected")
	}))
	defer srv.Close()

	root := t.TempDir()
	write := func(rel, contents string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("reqly.yaml", `{"baseURL":"`+srv.URL+`"}`)
	write("collections/main/reqly.yaml", `{"name":"main"}`)
	write("collections/main/a.yaml", `
request:
  method: GET
  url: /a
postRequest: |
  reqly.test("expects hello", function() { return reqly.response.body === "hello"; });
`)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "test", "main", "--workspace", root})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected failure exit for failing assertion")
	}
	if !strings.Contains(err.Error(), "1 step(s) failed") {
		t.Fatalf("unexpected error %q", err)
	}
	if !strings.Contains(stripANSI(out.String()), "[FAIL] expects hello") {
		t.Fatalf("expected failing test in output, got:\n%s", out.String())
	}
}

func TestCollectionTestNotFound(t *testing.T) {
	resetCollectionTestFlags()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "reqly.yaml"), []byte(`{"name":"demo"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "test", "nope", "--workspace", root})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestCollectionTestVariableChaining(t *testing.T) {
	resetCollectionTestFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"token":"tok-9"}`)
			return
		}
		if r.Header.Get("Authorization") == "Bearer tok-9" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"user":"reqly"}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	root := t.TempDir()
	write := func(rel, contents string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("reqly.yaml", `{"baseURL":"`+srv.URL+`"}`)
	write("collections/main/reqly.yaml", `{"name":"main"}`)
	write("collections/main/login.yaml", `
request:
  method: POST
  url: /login
postRequest: |
  reqly.setVariable("token", JSON.parse(reqly.response.body).token);
`)
	write("collections/main/me.yaml", `
request:
  method: GET
  url: /me
  headers:
    - key: Authorization
      value: "Bearer {{token}}"
`)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "test", "main", "--workspace", root})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "2 passed, 0 failed") {
		t.Fatalf("expected chaining to pass both, got:\n%s", out.String())
	}
}

func resetCollectionTestFlags() {
	collectionFailFast = false
	if flag := collectionTestCmd.Flags().Lookup("fail-fast"); flag != nil {
		flag.Changed = false
	}
}

// stripANSI removes ANSI escape sequences used for colored CLI output.
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}
