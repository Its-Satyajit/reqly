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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func writeTempBulkFiles(t *testing.T, reqContent, dataContent, dataExt string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	reqPath := filepath.Join(dir, "req.yaml")
	dataPath := filepath.Join(dir, "data"+dataExt)
	if err := os.WriteFile(reqPath, []byte(reqContent), 0o644); err != nil {
		t.Fatalf("write req: %v", err)
	}
	if err := os.WriteFile(dataPath, []byte(dataContent), 0o644); err != nil {
		t.Fatalf("write data: %v", err)
	}
	return reqPath, dataPath
}

func TestBulkRun_CSVSequential(t *testing.T) {
	var urls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urls = append(urls, r.URL.String())
		fmt.Fprint(w, `ok`)
	}))
	defer srv.Close()

	reqContent := fmt.Sprintf(`
name: test
request:
  url: %s/users/{{id}}
  method: GET
`, srv.URL)
	dataContent := "id,name\n1,alice\n2,bob\n"
	reqPath, dataPath := writeTempBulkFiles(t, reqContent, dataContent, ".csv")

	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"bulk", "run", reqPath, "--data", dataPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v errBuf=%q", err, errBuf.String())
	}
	if !strings.Contains(out.String(), "step 1:") || !strings.Contains(out.String(), "step 2:") {
		t.Fatalf("output missing steps: got %q", out.String())
	}
	if len(urls) != 2 {
		t.Fatalf("calls: got %d want 2", len(urls))
	}
	if !strings.Contains(urls[0], "/1") || !strings.Contains(urls[1], "/2") {
		t.Fatalf("urls: got %v", urls)
	}
}

func TestBulkRun_JSON(t *testing.T) {
	var urls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urls = append(urls, r.URL.String())
		fmt.Fprint(w, `ok`)
	}))
	defer srv.Close()

	reqContent := fmt.Sprintf(`
name: test
request:
  url: %s/users/{{id}}
  method: GET
`, srv.URL)
	arr := []map[string]any{{"id": 1}, {"id": 2}}
	b, _ := json.Marshal(arr)
	reqPath, dataPath := writeTempBulkFiles(t, reqContent, string(b), ".json")

	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"bulk", "run", reqPath, "--data", dataPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(urls) != 2 {
		t.Fatalf("json calls: got %d", len(urls))
	}
}

func TestBulkRun_Parallel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `ok`)
	}))
	defer srv.Close()

	reqContent := fmt.Sprintf(`
name: test
request:
  url: %s/users/{{id}}
  method: GET
`, srv.URL)
	dataContent := "id\n1\n2\n3\n"
	reqPath, dataPath := writeTempBulkFiles(t, reqContent, dataContent, ".csv")

	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"bulk", "run", reqPath, "--data", dataPath, "--parallel", "--concurrency", "2"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("parallel: %v", err)
	}
	if !strings.Contains(out.String(), "step 3:") {
		t.Fatalf("parallel steps: got %q", out.String())
	}
}

func TestBulkRun_ContinueOnError(t *testing.T) {
	var count atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		if r.URL.Path == "/users/2" {
			w.WriteHeader(500)
			fmt.Fprint(w, `err`)
			return
		}
		fmt.Fprint(w, `ok`)
	}))
	defer srv.Close()

	reqContent := fmt.Sprintf(`
name: test
request:
  url: %s/users/{{id}}
  method: GET
`, srv.URL)
	dataContent := "id\n1\n2\n3\n"
	reqPath, dataPath := writeTempBulkFiles(t, reqContent, dataContent, ".csv")

	// without continue, should stop at 2
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"bulk", "run", reqPath, "--data", dataPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if strings.Contains(out.String(), "step 3:") {
		t.Fatalf("should stop at 2 without continue: got %q", out.String())
	}

	// with continue, should see step 3
	count.Store(0)
	var out2, errBuf2 bytes.Buffer
	rootCmd.SetOut(&out2)
	rootCmd.SetErr(&errBuf2)
	rootCmd.SetArgs([]string{"bulk", "run", reqPath, "--data", dataPath, "--continue-on-error"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("continue: %v", err)
	}
	if !strings.Contains(out2.String(), "step 3:") {
		t.Fatalf("continue should see step 3: got %q", out2.String())
	}
}

func TestBulkRun_Help(t *testing.T) {
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"bulk", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(out.String(), "run") {
		t.Fatalf("help should list run: got %q", out.String())
	}
}

func TestBulkRun_MissingData(t *testing.T) {
	dir := t.TempDir()
	reqPath := filepath.Join(dir, "req.yaml")
	if err := os.WriteFile(reqPath, []byte("name: test\nrequest:\n  url: https://example.com\n  method: GET\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"bulk", "run", reqPath})
	if err := rootCmd.Execute(); err == nil {
		t.Fatalf("expected error for missing --data")
	}
}
