// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkflowRun_Desktop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token": "desktop123"}`))
	}))
	defer srv.Close()

	yamlStr := fmt.Sprintf(`
name: Desktop Flow
steps:
  - id: s1
    name: S1
    request:
      method: GET
      url: %s
    extract:
      token: token
`, srv.URL)

	svc := NewAppService()
	report, err := svc.WorkflowRun(yamlStr)
	if err != nil {
		t.Fatalf("WorkflowRun: %v", err)
	}
	if report == nil || !report.Passed {
		t.Fatalf("expected passed report, got %+v", report)
	}
	if report.ExtractedVars["token"] != "desktop123" {
		t.Fatalf("want desktop123, got %q", report.ExtractedVars["token"])
	}
	if len(report.Steps) != 1 {
		t.Fatalf("want 1 step, got %d", len(report.Steps))
	}
}

func TestWorkflowRun_EmptyYaml(t *testing.T) {
	svc := NewAppService()
	if _, err := svc.WorkflowRun(""); err == nil {
		t.Fatalf("expected error for empty yaml")
	}
}

func TestWorkflowRun_InvalidYaml(t *testing.T) {
	svc := NewAppService()
	if _, err := svc.WorkflowRun("not: [valid: yaml:"); err == nil {
		t.Fatalf("expected error for invalid yaml")
	}
}
