package workflow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
)

func TestExecuteWorkflow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"token": "secret123"}`))
		case "/profile":
			if r.Header.Get("Authorization") == "Bearer secret123" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"user": "alice"}`))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
	}))
	defer srv.Close()

	wf := &Workflow{
		Name: "Authentication Flow",
		Steps: []WorkflowStep{
			{
				ID:   "login",
				Name: "Login Step",
				Request: request.Request{
					Method: "POST",
					URL:    srv.URL + "/login",
				},
				Condition: "",
				Extract: map[string]string{
					"token": "token",
				},
			},
			{
				ID:   "profile",
				Name: "Profile Step",
				Request: request.Request{
					Method: "GET",
					URL:    srv.URL + "/profile",
					Headers: []request.Header{
						{Key: "Authorization", Value: "Bearer {{token}}"},
					},
				},
				Condition: "reqly.getVariable('token') !== ''",
			},
		},
	}

	exec := NewWorkflowExecutor(nil)
	report, err := exec.Execute(context.Background(), wf, WorkflowOptions{})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	if !report.Passed {
		t.Fatalf("expected workflow to pass, report: %+v", report)
	}
	if len(report.Steps) != 2 {
		t.Fatalf("want 2 steps, got %d", len(report.Steps))
	}
	if report.ExtractedVars["token"] != "secret123" {
		t.Fatalf("want secret123, got %q", report.ExtractedVars["token"])
	}
}

func TestExecuteWorkflow_ConditionSkip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	wf := &Workflow{
		Name:      "Skip",
		Variables: map[string]string{"run": ""},
		Steps: []WorkflowStep{
			{
				ID:   "s1",
				Name: "ShouldSkip",
				Request: request.Request{
					Method: "GET",
					URL:    srv.URL,
				},
				Condition: "reqly.getVariable('run') !== ''",
			},
			{
				ID:   "s2",
				Name: "ShouldRun",
				Request: request.Request{
					Method: "GET",
					URL:    srv.URL,
				},
			},
		},
	}
	exec := NewWorkflowExecutor(nil)
	report, err := exec.Execute(context.Background(), wf, WorkflowOptions{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(report.Steps) != 1 {
		t.Fatalf("want 1 step after skip, got %d", len(report.Steps))
	}
	if report.Steps[0].Name != "ShouldRun" {
		t.Fatalf("unexpected step: %q", report.Steps[0].Name)
	}
}

func TestExecuteWorkflow_ExtractMissingKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"a": "1"}`))
	}))
	defer srv.Close()

	wf := &Workflow{
		Name: "ExtractMissing",
		Steps: []WorkflowStep{
			{
				ID:   "s1",
				Name: "S1",
				Request: request.Request{
					Method: "GET",
					URL:    srv.URL,
				},
				Extract: map[string]string{"missing": "nonexistent"},
			},
		},
	}
	exec := NewWorkflowExecutor(nil)
	report, err := exec.Execute(context.Background(), wf, WorkflowOptions{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(report.ExtractedVars) != 0 {
		t.Fatalf("want 0 extracted, got %+v", report.ExtractedVars)
	}
}

func TestExecuteWorkflow_QueryInterpolation(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("token")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	wf := &Workflow{
		Name:      "QueryInterp",
		Variables: map[string]string{"token": "abc123"},
		Steps: []WorkflowStep{
			{
				ID:   "s1",
				Name: "S1",
				Request: request.Request{
					Method: "GET",
					URL:    srv.URL,
					Query: []request.Parameter{
						{Key: "token", Value: "{{token}}"},
					},
				},
			},
		},
	}
	exec := NewWorkflowExecutor(nil)
	report, err := exec.Execute(context.Background(), wf, WorkflowOptions{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !report.Passed {
		t.Fatalf("expected pass, got %+v", report)
	}
	if gotQuery != "abc123" {
		t.Fatalf("want abc123 query, got %q", gotQuery)
	}
}

func TestExecuteWorkflow_NilWorkflow(t *testing.T) {
	exec := NewWorkflowExecutor(nil)
	if _, err := exec.Execute(context.Background(), nil, WorkflowOptions{}); err == nil {
		t.Fatalf("expected error for nil workflow")
	}
}

func TestExecuteWorkflow_FailureStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`error`))
	}))
	defer srv.Close()

	wf := &Workflow{
		Name: "Fail",
		Steps: []WorkflowStep{
			{
				ID:   "s1",
				Name: "S1",
				Request: request.Request{
					Method: "GET",
					URL:    srv.URL,
				},
			},
		},
	}
	exec := NewWorkflowExecutor(nil)
	report, err := exec.Execute(context.Background(), wf, WorkflowOptions{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if report.Passed {
		t.Fatalf("expected failed report")
	}
	if len(report.Steps) != 1 || report.Steps[0].Passed {
		t.Fatalf("expected step to be failed, got %+v", report.Steps)
	}
	if report.Steps[0].RequestError != "" {
		t.Fatalf("unexpected request error for HTTP 500, should be empty, got %q", report.Steps[0].RequestError)
	}
}

func TestExecuteWorkflow_EnvironmentVars(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Env")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	wf := &Workflow{
		Name: "Env",
		Steps: []WorkflowStep{
			{
				ID:   "s1",
				Name: "S1",
				Request: request.Request{
					Method: "GET",
					URL:    srv.URL,
					Headers: []request.Header{
						{Key: "X-Env", Value: "{{envVal}}"},
					},
				},
			},
		},
	}
	exec := NewWorkflowExecutor(nil)
	report, err := exec.Execute(context.Background(), wf, WorkflowOptions{
		EnvironmentVars: map[string]string{"envVal": "prod"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !report.Passed {
		t.Fatalf("expected pass")
	}
	if gotHeader != "prod" {
		t.Fatalf("want prod, got %q", gotHeader)
	}
}
