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
)

func writeTempPaginationFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "req.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestPaginationRun_Page(t *testing.T) {
	// httptest server paginates via ?page=
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		page := r.URL.Query().Get("page")
		if page == "3" {
			fmt.Fprint(w, "[]")
			return
		}
		fmt.Fprint(w, `[{"id":1}]`)
	}))
	defer srv.Close()

	content := fmt.Sprintf(`
name: test
request:
  url: %s
  method: GET
  pagination:
    strategy: page
    maxPages: 3
`, srv.URL)
	path := writeTempPaginationFile(t, content)

	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"pagination", "run", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v errBuf=%q", err, errBuf.String())
	}
	if !strings.Contains(out.String(), "step 1:") || !strings.Contains(out.String(), "step 3:") {
		t.Fatalf("output missing steps: got %q", out.String())
	}
	if count != 3 {
		t.Fatalf("server calls: got %d want 3", count)
	}
}

func TestPaginationRun_Cursor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		if cursor == "b" {
			fmt.Fprint(w, `{"items":[{"id":3}], "nextCursor": ""}`)
			return
		}
		if cursor == "a" {
			fmt.Fprint(w, `{"items":[{"id":2}], "nextCursor": "b"}`)
			return
		}
		fmt.Fprint(w, `{"items":[{"id":1}], "nextCursor": "a"}`)
	}))
	defer srv.Close()

	content := fmt.Sprintf(`
name: test
request:
  url: %s
  method: GET
  pagination:
    strategy: cursor
    cursorParam: cursor
    nextPath: $.nextCursor
    maxPages: 3
`, srv.URL)
	path := writeTempPaginationFile(t, content)

	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"pagination", "run", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "step 3:") {
		t.Fatalf("cursor pagination steps: got %q", out.String())
	}
}

func TestPaginationRun_LinkHeader(t *testing.T) {
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `[]`)
			return
		}
		w.Header().Set("Link", fmt.Sprintf(`<%s?page=2>; rel="next"`, srvURL))
		fmt.Fprint(w, `[{"id":1}]`)
	}))
	srvURL = srv.URL
	defer srv.Close()

	content := fmt.Sprintf(`
name: test
request:
  url: %s?page=1
  method: GET
  pagination:
    strategy: link-header
    maxPages: 2
`, srv.URL)
	path := writeTempPaginationFile(t, content)

	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"pagination", "run", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "step 2:") {
		t.Fatalf("link-header steps: got %q", out.String())
	}
}

func TestPaginationRun_MaxPagesOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"id":1}]`)
	}))
	defer srv.Close()

	content := fmt.Sprintf(`
name: test
request:
  url: %s
  method: GET
  pagination:
    strategy: page
    maxPages: 100
`, srv.URL)
	path := writeTempPaginationFile(t, content)

	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"pagination", "run", path, "--max-pages", "2"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	// should only do 2 steps due to override
	if strings.Contains(out.String(), "step 3:") {
		t.Fatalf("max-pages override failed: got %q", out.String())
	}
	if !strings.Contains(out.String(), "step 2:") {
		t.Fatalf("expected 2 steps: got %q", out.String())
	}
}

func TestPaginationRun_Help(t *testing.T) {
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"pagination", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(out.String(), "run") {
		t.Fatalf("help should list run: got %q", out.String())
	}
}

func TestPaginationRun_MissingPaginationBlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[]`)
	}))
	defer srv.Close()
	content := fmt.Sprintf(`
name: test
request:
  url: %s
  method: GET
`, srv.URL)
	path := writeTempPaginationFile(t, content)

	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"pagination", "run", path})
	if err := rootCmd.Execute(); err == nil {
		t.Fatalf("expected error for missing pagination block")
	}
}
