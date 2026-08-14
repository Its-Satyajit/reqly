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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunCommandExecutesRequest(t *testing.T) {
	resetRunFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", srv.URL})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{"200 OK", "Content-Type: application/json", `{"ok":true}`} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestRunCommandSendsHeadersAndBody(t *testing.T) {
	resetRunFlags()
	var gotHeader, gotBody, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Test")
		gotMethod = r.Method
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		gotBody = string(buf)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{
		"run", "-m", "POST", "-H", "X-Test: hello", "-d", `{"a":1}`, srv.URL,
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if gotMethod != "POST" {
		t.Fatalf("expected POST, got %q", gotMethod)
	}
	if gotHeader != "hello" {
		t.Fatalf("expected header 'hello', got %q", gotHeader)
	}
	if gotBody != `{"a":1}` {
		t.Fatalf("expected body %q, got %q", `{"a":1}`, gotBody)
	}
}

func TestRunCommandInvalidHeader(t *testing.T) {
	resetRunFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", "-H", "BadHeader", "https://example.com"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for invalid header")
	}
}

func TestRunCommandNetworkError(t *testing.T) {
	resetRunFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", "http://127.0.0.1:1"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("unexpected error %q", err)
	}
}

func TestParseHeaders(t *testing.T) {
	headers, err := parseHeaders([]string{"Content-Type: application/json", " X-A : b "})
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(headers))
	}
	if headers[0].Key != "Content-Type" || headers[0].Value != "application/json" {
		t.Fatalf("unexpected header %+v", headers[0])
	}
	if headers[1].Key != "X-A" || headers[1].Value != "b" {
		t.Fatalf("unexpected header %+v", headers[1])
	}
}

func TestParseHeadersInvalid(t *testing.T) {
	if _, err := parseHeaders([]string{"NoColon"}); err == nil {
		t.Fatal("expected error for header without colon")
	}
}

// resetRunFlags clears flag state that persists across tests on the shared command.
func resetRunFlags() {
	runMethod = "GET"
	runHeaders = nil
	runBody = ""
	runTimeout = 30 * time.Second
}
