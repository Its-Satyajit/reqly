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
	"sync/atomic"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/history"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// newRunWS creates a minimal workspace: a descriptor pinning no
// environment plus a "dev" environment carrying one secret and one variable.
func newRunWS(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	desc := "name: ws\nenvironment: dev\n"
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte(desc), 0o644); err != nil {
		t.Fatal(err)
	}
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	env := "name: dev\nvariables:\n  base: example\ndescription: \"\"\nsecrets:\n  api_key: supersecret\n"
	if err := os.WriteFile(filepath.Join(envDir, "dev.yaml"), []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func echoServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestRun_MasksSecretsInBodyAndHeaders(t *testing.T) {
	dir := newRunWS(t)
	srv := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Echo", "bearer supersecret")
		fmt.Fprint(w, `{"key":"supersecret"}`)
	})

	svc := NewRunService(dir)
	defer svc.Close()
	res, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL + "/{{base}}",
	}, RunRequestOptions{FileEnv: "dev"})
	if err != nil {
		t.Fatal(err)
	}
	out := string(res.Response.Body) + strings.Join(res.Response.Headers["X-Echo"], "")
	if strings.Contains(out, "supersecret") {
		t.Fatalf("secret leaked into result: %q", out)
	}
}

func TestRun_InterpolatesEnvironmentVariables(t *testing.T) {
	dir := newRunWS(t)
	var gotPath atomic.Value
	srv := echoServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath.Store(r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})
	svc := NewRunService(dir)
	defer svc.Close()
	if _, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL + "/{{base}}",
	}, RunRequestOptions{FileEnv: "dev"}); err != nil {
		t.Fatal(err)
	}
	if p := gotPath.Load(); p == nil || p.(string) != "/example" {
		t.Fatalf("expected /example interpolated from env var, got %v", gotPath.Load())
	}
}

func TestRun_RuntimeVarsWinOverEnvironmentScope(t *testing.T) {
	dir := newRunWS(t)
	var gotPath atomic.Value
	srv := echoServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath.Store(r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})

	runtime := variables.NewSet()
	runtime.Set(variables.ScopeRuntime, "base", "override")

	svc := NewRunService(dir)
	defer svc.Close()
	if _, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL + "/{{base}}",
	}, RunRequestOptions{FileEnv: "dev", RuntimeVars: runtime}); err != nil {
		t.Fatal(err)
	}
	if p := gotPath.Load(); p == nil || p.(string) != "/override" {
		t.Fatalf("expected runtime var to win, got %v", gotPath.Load())
	}
}

func TestRun_RecordsHistoryWithMaskedShow(t *testing.T) {
	dir := newRunWS(t)
	srv := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"key":"supersecret"}`)
	})

	record := true
	svc := NewRunService(dir)
	defer svc.Close()
	res, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
	}, RunRequestOptions{FileEnv: "dev", RequestPath: "flaky.yaml", RecordHistory: &record})
	if err != nil {
		t.Fatal(err)
	}
	if res.Response.Attempts != 1 {
		t.Fatalf("expected Attempts=1, got %d", res.Response.Attempts)
	}
	if svc.historySvc == nil || svc.historySvc.store == nil {
		t.Fatal("expected history store to open lazily")
	}
	entries, err := svc.historySvc.List(context.Background(), 10, 0, nil)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected exactly one history entry, got %v (%v)", entries, err)
	}
	e := entries[0]
	if e.RequestPath != "flaky.yaml" {
		t.Fatalf("unexpected entry %+v", e)
	}
	// Stored raw bytes must contain the secret (exact traffic), while Show
	// masks only headers — body masking is a display concern of callers.
	store, _ := history.NewStore(filepath.Join(dir, ".reqly", "history.db"))
	defer store.Close()
	shown, err := store.Show(context.Background(), e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(shown.RespBody), "supersecret") {
		t.Fatalf("expected raw bytes recorded, got %q", shown.RespBody)
	}
	if res.Warning != "" {
		t.Fatalf("unexpected warning %q", res.Warning)
	}
}

func TestRun_RecordHistoryOptOut(t *testing.T) {
	dir := newRunWS(t)
	srv := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	disabled := false
	svc := NewRunService(dir)
	defer svc.Close()
	if _, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
	}, RunRequestOptions{RecordHistory: &disabled}); err != nil {
		t.Fatal(err)
	}
	if svc.historySvc != nil {
		t.Fatal("history store should not be needed when recording is off")
	}
}

func TestRun_OnRetryObserved(t *testing.T) {
	dir := newRunWS(t)
	var calls atomic.Int32
	srv := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	retries := 0
	svc := NewRunService(dir)
	defer svc.Close()
	res, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
		Retry:  &request.Retry{Count: 3, DelayMs: 1, Strategy: request.RetryStrategyFixed},
	}, RunRequestOptions{
		OnRetry: func(request.RetryEvent) { retries++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	if retries != 1 || res.Response.Attempts != 2 {
		t.Fatalf("expected 1 retry notice and 2 attempts, got %d/%d", retries, res.Response.Attempts)
	}
}

func TestRun_MissingEnvIsHardError(t *testing.T) {
	dir := newRunWS(t)
	svc := NewRunService(dir)
	defer svc.Close()
	_, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    "https://localhost:1",
	}, RunRequestOptions{FileEnv: "nonexistent"})
	if err == nil {
		t.Fatal("expected missing selected environment to be a hard error")
	}
}

func TestNewRunService_UnboundRootStillRuns(t *testing.T) {
	srv := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `ok`)
	})
	svc := NewRunService("")
	defer svc.Close()
	if _, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
	}, RunRequestOptions{}); err != nil {
		t.Fatal(err)
	}
	if svc.historySvc != nil {
		t.Fatal("no workspace → no history recording")
	}
}
