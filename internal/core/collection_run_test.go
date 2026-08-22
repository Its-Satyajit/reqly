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

package core

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// writeRunWorkspace builds a temp workspace with a base URL, one collection
// (users) with two requests, and a nested folder (users/auth) with one
// request. Returns the workspace root.
func writeRunWorkspace(t *testing.T, baseURL string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"reqly.yaml":                        fmt.Sprintf("name: demo\nbaseURL: %s\n", baseURL),
		"collections/users/reqly.yaml":      "name: users\n",
		"collections/users/a.yaml":          "request: {method: GET, url: /a}",
		"collections/users/b.yaml":          "request: {method: GET, url: /b}",
		"collections/users/auth/reqly.yaml": "name: auth\n",
		"collections/users/auth/login.yaml": "request: {method: POST, url: /login}",
	}
	for name, contents := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func okServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"path":%q}`, r.URL.Path)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestCollectionRunServiceRunCollection(t *testing.T) {
	srv := okServer(t)
	dir := writeRunWorkspace(t, srv.URL)

	svc := NewCollectionRunService(dir)
	var streamed []string
	report, err := svc.Run(context.Background(), "users", RunOptions{
		OnStep: func(step RunStep) { streamed = append(streamed, step.Name) },
	})
	if err != nil {
		t.Fatal(err)
	}

	if report.Total != 3 || report.Passed != 3 || report.Failed != 0 || !report.OK {
		t.Fatalf("report = %+v", report)
	}
	wantOrder := []string{"a", "b", "login"}
	if len(streamed) != 3 || streamed[0] != "a" || streamed[1] != "b" || streamed[2] != "login" {
		t.Fatalf("streamed = %v, want %v", streamed, wantOrder)
	}
	for i, step := range report.Steps {
		if step.Name != wantOrder[i] {
			t.Fatalf("steps[%d] = %q, want %q", i, step.Name, wantOrder[i])
		}
		if step.Response == nil || !step.Response.OK || step.Response.StatusCode != 200 {
			t.Fatalf("steps[%d] response = %+v", i, step.Response)
		}
	}
	if report.DurationMS < 0 || report.Started.After(report.Finished) {
		t.Fatalf("timing fields invalid: started=%v finished=%v durationMs=%d", report.Started, report.Finished, report.DurationMS)
	}
}

func TestCollectionRunServiceRunFolder(t *testing.T) {
	srv := okServer(t)
	dir := writeRunWorkspace(t, srv.URL)

	svc := NewCollectionRunService(dir)
	report, err := svc.Run(context.Background(), "users/auth", RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 1 || report.Passed != 1 {
		t.Fatalf("report = %+v", report)
	}
	if report.Steps[0].Name != "login" || report.Steps[0].RequestPath != "users/auth/login" {
		t.Fatalf("step = %+v", report.Steps[0])
	}
}

func TestCollectionRunServiceFailFast(t *testing.T) {
	var hits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	dir := writeRunWorkspace(t, srv.URL)

	svc := NewCollectionRunService(dir)
	report, err := svc.Run(context.Background(), "users", RunOptions{FailFast: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 3 || len(report.Steps) != 1 || report.Failed != 1 {
		t.Fatalf("report = %+v", report)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %v", hits)
	}
}

func TestCollectionRunServiceSingleFlight(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	dir := writeRunWorkspace(t, srv.URL)

	svc := NewCollectionRunService(dir)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = svc.Run(ctx, "users", RunOptions{})
	}()

	<-started
	_, err := svc.Run(context.Background(), "users", RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("expected single-flight error, got %v", err)
	}
	close(release)
	wg.Wait()
}

func TestCollectionRunServiceMissingTarget(t *testing.T) {
	srv := okServer(t)
	dir := writeRunWorkspace(t, srv.URL)
	svc := NewCollectionRunService(dir)

	for _, path := range []string{"nope", "users/nope", "nope/deep"} {
		if _, err := svc.Run(context.Background(), path, RunOptions{}); err == nil {
			t.Fatalf("expected error for %q", path)
		}
	}
}

func TestCollectionRunServiceErrorStepDTO(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "boom")
	}))
	defer srv.Close()
	dir := writeRunWorkspace(t, srv.URL)

	svc := NewCollectionRunService(dir)
	report, err := svc.Run(context.Background(), "users", RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if report.OK || report.Failed != 3 {
		t.Fatalf("report = %+v", report)
	}
	step := report.Steps[0]
	if step.Passed {
		t.Fatal("expected failed step")
	}
	if step.RequestError != "" {
		t.Fatalf("transport errors should be empty for HTTP failures, got %q", step.RequestError)
	}
	if step.Response == nil || step.Response.StatusCode != 500 || step.Response.Body != "boom" {
		t.Fatalf("step response = %+v", step.Response)
	}
}

func TestCollectionRunServicePreScriptErrorDTO(t *testing.T) {
	srv := okServer(t)
	dir := t.TempDir()
	files := map[string]string{
		"reqly.yaml":                   fmt.Sprintf("name: demo\nbaseURL: %s\n", srv.URL),
		"collections/users/reqly.yaml": "name: users\n",
		"collections/users/x.yaml":     "request: {method: GET, url: /x}\npreRequest: throw new Error('boom')\n",
	}
	for name, contents := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	svc := NewCollectionRunService(dir)
	report, err := svc.Run(context.Background(), "users", RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	step := report.Steps[0]
	if step.Passed || step.RequestError == "" {
		t.Fatalf("expected pre-script error string, got %+v", step)
	}
	if !strings.Contains(step.RequestError, "boom") {
		t.Fatalf("error = %q, want boom", step.RequestError)
	}
}

func TestCollectionRunServiceMasksSecrets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"token":"superSecret"}`)
	}))
	defer srv.Close()
	dir := t.TempDir()
	files := map[string]string{
		"reqly.yaml":                   fmt.Sprintf("name: demo\nbaseURL: %s\nenvironment: prod\n", srv.URL),
"environments/prod.yaml":         "variables:\n  env: prod\nsecrets:\n  token: superSecret\n",
		"collections/users/reqly.yaml": "name: users\n",
		"collections/users/a.yaml":     "request: {method: GET, url: /a}\npostRequest: |\n  console.log(reqly.response.body);\n",
	}
	for name, contents := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	svc := NewCollectionRunService(dir)
	report, err := svc.Run(context.Background(), "users", RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	step := report.Steps[0]
	if strings.Contains(step.Response.Body, "superSecret") {
		t.Fatalf("body leaks secret: %q", step.Response.Body)
	}
	if !strings.Contains(step.Response.Body, "[SECRET]") {
		t.Fatalf("body should contain masked secret: %q", step.Response.Body)
	}
	for _, log := range step.Logs {
		if strings.Contains(log, "superSecret") {
			t.Fatalf("log leaks secret: %q", log)
		}
	}
}

func TestCollectionRunServiceNoRoot(t *testing.T) {
	svc := NewCollectionRunService("")
	if _, err := svc.Run(context.Background(), "users", RunOptions{}); err == nil {
		t.Fatal("expected error for missing workspace root")
	}
}

func TestCollectionRunServiceCancel(t *testing.T) {
	started := make(chan struct{})
	var startedOnce sync.Once
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedOnce.Do(func() { close(started) })
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	dir := writeRunWorkspace(t, srv.URL)

	svc := NewCollectionRunService(dir)
	ctx, cancel := context.WithCancel(context.Background())

	var streamed []string
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = svc.Run(ctx, "users", RunOptions{OnStep: func(step RunStep) { streamed = append(streamed, step.Name) }})
	}()

	<-started
	cancel()
	<-done
	close(release)

	// The current step (a) may complete and stream; no further steps schedule.
	if len(streamed) != 1 || streamed[0] != "a" {
		t.Fatalf("streamed = %v, want [a]", streamed)
	}
	// A cancelled run must not leak the single-flight lock.
	if _, err := svc.Run(context.Background(), "users", RunOptions{}); err != nil {
		t.Fatalf("single-flight lock leaked: %v", err)
	}
}
