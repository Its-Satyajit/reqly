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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// makeTestWorkspace builds a workspace tree pointing at a test server.
func makeTestWorkspace(t *testing.T, srvURL string) string {
	t.Helper()
	root := t.TempDir()

	writeFile := func(rel, contents string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeFile("reqly.yaml", `{"name":"demo","baseURL":"`+srvURL+`","headers":[{"key":"X-Workspace","value":"ws"}]}`)
	writeFile("collections/users/reqly.yaml", `{"name":"users","baseURL":"/api","variables":{"SHARED":"users"}}`)
	writeFile("collections/users/list-users.yaml", `{"request":{"method":"GET","url":"users"}}`)
	writeFile("collections/users/auth/reqly.yaml", `{"name":"auth","headers":[{"key":"Authorization","value":"Bearer admin-token"}]}`)
	writeFile("collections/users/auth/me.yaml", `{"request":{"method":"GET","url":"auth/me"}}`)

	return root
}

func TestCollectionList(t *testing.T) {
	resetCollectionFlags()
	root := makeTestWorkspace(t, "http://example.com")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "list", "--workspace", root})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"demo/", "users/", "list-users", "auth/", "me"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out.String())
		}
	}
}

func TestCollectionListMissingWorkspace(t *testing.T) {
	resetCollectionFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "list", "--workspace", t.TempDir()})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not a workspace") {
		t.Fatalf("expected 'not a workspace' error, got %v", err)
	}
}

func TestCollectionRunResolvesInheritance(t *testing.T) {
	resetCollectionFlags()
	var gotHeader, gotURL, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Workspace")
		gotMethod = r.Method
		gotURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	root := makeTestWorkspace(t, srv.URL)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "run", "--workspace", root, "users/list-users.yaml"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if gotMethod != "GET" || gotURL != "/api/users" {
		t.Fatalf("got method %q url %q", gotMethod, gotURL)
	}
	if gotHeader != "ws" {
		t.Fatalf("workspace header not inherited: got %q", gotHeader)
	}
	if !strings.Contains(out.String(), "200 OK") {
		t.Fatalf("expected 200 OK in output, got:\n%s", out.String())
	}
}

func TestCollectionRunFolderInheritance(t *testing.T) {
	resetCollectionFlags()
	var gotAuth, gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	root := makeTestWorkspace(t, srv.URL)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "run", "--workspace", root, "users/auth/me.yaml"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if gotURL != "/api/auth/me" {
		t.Fatalf("url: got %q", gotURL)
	}
	if gotAuth != "Bearer admin-token" {
		t.Fatalf("folder auth header not inherited: got %q", gotAuth)
	}
}

func TestCollectionRunRequestNotFound(t *testing.T) {
	resetCollectionFlags()
	root := makeTestWorkspace(t, "http://example.com")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"collection", "run", "--workspace", root, "users/nope.yaml"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' error, got %v", err)
	}
}

// resetCollectionFlags clears package-level flag state that persists across tests.
func resetCollectionFlags() {
	collectionWorkspace, _ = filepath.Abs(".")
	for _, cmd := range []*cobra.Command{collectionRunCmd, collectionListCmd} {
		if flag := cmd.Flags().Lookup("workspace"); flag != nil {
			flag.Changed = false
		}
	}
}
