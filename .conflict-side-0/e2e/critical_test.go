// Reqly - E2E critical workflows (Go net/http smoke, deferred Playwright).
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/collab"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/workflow"
)

// TestE2E_CriticalWorkflows covers the three critical user journeys
// that Playwright will later drive in the browser: (1) send a request,
// (2) run a workflow, (3) check collaboration health.
// This Go smoke is the TDD red→green for Milestones/06 E2E deferred.
func TestE2E_CriticalWorkflows(t *testing.T) {
	dir := t.TempDir()

	// 1. Request send via workflow (login → profile)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"token": "e2e-tok"}`))
		case "/profile":
			if r.Header.Get("Authorization") == "Bearer e2e-tok" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"user": "e2e"}`))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	wf := &workflow.Workflow{
		Name: "E2E Flow",
		Steps: []workflow.WorkflowStep{
			{ID: "login", Name: "Login", Request: request.Request{Method: "POST", URL: srv.URL + "/login"}, Extract: map[string]string{"token": "token"}},
			{ID: "profile", Name: "Profile", Request: request.Request{Method: "GET", URL: srv.URL + "/profile", Headers: []request.Header{{Key: "Authorization", Value: "Bearer {{token}}"}}}},
		},
	}
	report, err := workflow.NewWorkflowExecutor(nil).Execute(t.Context(), wf, workflow.WorkflowOptions{})
	if err != nil {
		t.Fatalf("workflow: %v", err)
	}
	if !report.Passed {
		t.Fatalf("workflow should pass: %+v", report)
	}

	// 2. Collaboration health
	collabWS := collab.SharedWorkspace{Path: dir, Collaborators: []collab.Collaborator{{User: "e2e", Role: "admin"}}}
	if err := collab.Save(collab.DefaultPath(dir), collabWS); err != nil {
		t.Fatalf("collab save: %v", err)
	}
	cs := collab.NewServer(dir)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	cs.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("health want 200, got %d", w.Code)
	}
	var out map[string]string
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil || out["status"] != "ok" {
		t.Fatalf("health unexpected: %v %v", err, out)
	}

	// 3. Verify workspace file persisted via collab server /workspace
	if err := os.MkdirAll(filepath.Join(dir, "collections", "e2e"), 0o755); err != nil {
		t.Fatal(err)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/workspace", nil)
	w2 := httptest.NewRecorder()
	cs.Handler().ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("workspace want 200, got %d", w2.Code)
	}
}
