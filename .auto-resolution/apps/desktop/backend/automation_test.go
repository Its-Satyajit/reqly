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

func TestAutomationRun_Desktop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	yamlStr := fmt.Sprintf(`
name: desktop-auto
workflow:
  name: wf
  steps:
    - id: s1
      name: S1
      request:
        method: GET
        url: %s
interval: "0"
`, srv.URL)

	svc := NewAppService()
	report, err := svc.AutomationRun(yamlStr)
	if err != nil {
		t.Fatalf("AutomationRun: %v", err)
	}
	if report == nil || !report.Passed {
		t.Fatalf("expected passed, got %+v", report)
	}
	if len(report.Steps) != 1 {
		t.Fatalf("want 1 step, got %d", len(report.Steps))
	}
}

func TestAutomationRun_Empty(t *testing.T) {
	svc := NewAppService()
	if _, err := svc.AutomationRun(""); err == nil {
		t.Fatalf("expected error for empty")
	}
}
