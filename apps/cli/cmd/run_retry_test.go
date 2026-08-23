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
	"sync/atomic"
	"testing"
)

func flakyOnceServer(t *testing.T, status int) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(status)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func TestRunRetriesFromFile(t *testing.T) {
	resetRunFlags()
	srv, calls := flakyOnceServer(t, http.StatusServiceUnavailable)
	dir := t.TempDir()
	path := filepath.Join(dir, "flaky.yaml")
	src := "request:\n  method: GET\n  url: " + srv.URL + "\n  retry:\n    count: 3\n    delayMs: 1\n    strategy: fixed\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	if !strings.Contains(output, "200 OK") {
		t.Fatalf("expected eventual success, got:\n%s", output)
	}
	if !strings.Contains(output, "retrying in") || !strings.Contains(output, "attempt 1/") {
		t.Fatalf("expected a retry notice, got:\n%s", output)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 sends, got %d", got)
	}
}

func TestRunRetryFlagsOverrideFile(t *testing.T) {
	resetRunFlags()
	srv, calls := flakyOnceServer(t, http.StatusBadGateway)
	dir := t.TempDir()
	path := filepath.Join(dir, "flaky.yaml")
	// File disables retry (no block). Flags enable it.
	src := "request:\n  method: GET\n  url: " + srv.URL + "\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", path, "--retries", "2", "--retry-delay", "1ms"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "200 OK") {
		t.Fatalf("expected success via flag retries, got:\n%s", out.String())
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected flag-driven retry (2 sends), got %d", got)
	}
}

func TestRunRetryFlagsURLMode(t *testing.T) {
	resetRunFlags()
	srv, calls := flakyOnceServer(t, http.StatusTooManyRequests)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", srv.URL, "--retries", "1", "--retry-delay", "1ms"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "200 OK") {
		t.Fatalf("expected success in URL mode with retries, got:\n%s", out.String())
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 sends in URL mode, got %d", got)
	}
}

func TestRunReportsAttemptsWhenRetried(t *testing.T) {
	resetRunFlags()
	srv, _ := flakyOnceServer(t, http.StatusServiceUnavailable)
	dir := t.TempDir()
	path := filepath.Join(dir, "flaky.yaml")
	src := "request:\n  method: GET\n  url: " + srv.URL + "\n  retry:\n    count: 2\n    delayMs: 1\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "2 attempts") {
		t.Fatalf("expected attempt count in status line, got:\n%s", out.String())
	}
}
