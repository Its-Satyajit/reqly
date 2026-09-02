// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/audit"
	"github.com/Its-Satyajit/reqly/internal/collab"
	"github.com/Its-Satyajit/reqly/internal/policy"
	"github.com/Its-Satyajit/reqly/internal/rbac"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/workflow"
)

// TestPipeline_WorkflowWithPolicyRBACAuditCollab exercises the full
// core ↔ persistence ↔ engine pipeline in one integration test.
// It mirrors the "core ↔ persistence ↔ engine" integration coverage
// mentioned in Milestones/06.
func TestPipeline_WorkflowWithPolicyRBACAuditCollab(t *testing.T) {
	dir := t.TempDir()

	// Mock server for login → profile chain
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"token": "tok123"}`))
		case "/profile":
			if r.Header.Get("Authorization") == "Bearer tok123" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"user": "alice"}`))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Policy: allow all, audit required
	pol := policy.Policy{RequireAudit: true, MaxWorkflowSteps: 5}
	if err := policy.Validate(pol); err != nil {
		t.Fatalf("policy validate: %v", err)
	}
	if err := policy.Enforce(pol, "workflow.run", "test-workflow"); err != nil {
		t.Fatalf("policy enforce: %v", err)
	}
	if err := policy.Save(filepath.Join(dir, ".reqly", "policy.yaml"), pol); err != nil {
		t.Fatalf("policy save: %v", err)
	}

	// RBAC: alice is admin
	rb := rbac.DefaultRBAC()
	rb.UserRoles["alice"] = "admin"
	if err := rbac.Enforce(rb, "alice", "workflow.run", "test-workflow"); err != nil {
		t.Fatalf("rbac enforce: %v", err)
	}

	// Collab: add alice as admin
	ws := collab.SharedWorkspace{Path: dir}
	if err := collab.AddCollaborator(&ws, "alice", "admin"); err != nil {
		t.Fatalf("collab add: %v", err)
	}
	if err := collab.Save(collab.DefaultPath(dir), ws); err != nil {
		t.Fatalf("collab save: %v", err)
	}

	// Workflow: login → profile with token extraction
	wf := &workflow.Workflow{
		Name: "Integration Flow",
		Steps: []workflow.WorkflowStep{
			{
				ID:      "login",
				Name:    "Login",
				Request: request.Request{Method: "POST", URL: srv.URL + "/login"},
				Extract: map[string]string{"token": "token"},
			},
			{
				ID:   "profile",
				Name: "Profile",
				Request: request.Request{
					Method:  "GET",
					URL:     srv.URL + "/profile",
					Headers: []request.Header{{Key: "Authorization", Value: "Bearer {{token}}"}},
				},
			},
		},
	}
	exec := workflow.NewWorkflowExecutor(nil)
	report, err := exec.Execute(context.Background(), wf, workflow.WorkflowOptions{})
	if err != nil {
		t.Fatalf("workflow execute: %v", err)
	}
	if !report.Passed || len(report.Steps) != 2 {
		t.Fatalf("unexpected report %+v", report)
	}
	if report.ExtractedVars["token"] != "tok123" {
		t.Fatalf("want tok123, got %q", report.ExtractedVars["token"])
	}

	// Audit: record the workflow run
	store, err := audit.NewStore(dir)
	if err != nil {
		t.Fatalf("audit store: %v", err)
	}
	entry, err := store.Add(audit.Entry{Action: "workflow.run", Resource: "Integration Flow", Actor: "alice", Details: "2 steps"})
	if err != nil {
		t.Fatalf("audit add: %v", err)
	}
	if entry.ID == "" {
		t.Fatalf("expected audit ID")
	}
	list, err := store.List()
	if err != nil || len(list) != 1 {
		t.Fatalf("audit list: %v, len %d", err, len(list))
	}

	// Collab: verify alice is collaborator
	loadedWS, err := collab.Load(collab.DefaultPath(dir))
	if err != nil {
		t.Fatalf("collab load: %v", err)
	}
	if !collab.IsCollaborator(loadedWS, "alice") {
		t.Fatalf("alice should be collaborator")
	}

	// Policy workflow step count enforcement
	if err := policy.EnforceWorkflow(pol, len(wf.Steps)); err != nil {
		t.Fatalf("policy workflow enforce: %v", err)
	}
}
