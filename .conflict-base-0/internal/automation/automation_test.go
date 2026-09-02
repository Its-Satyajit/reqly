package automation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/workflow"
)

func TestAutomation_Validate(t *testing.T) {
	tests := []struct {
		name    string
		auto    Automation
		wantErr bool
	}{
		{
			name: "valid minimal",
			auto: Automation{
				Name: "a",
				Workflow: workflow.Workflow{
					Name: "wf",
					Steps: []workflow.WorkflowStep{
						{ID: "s1", Name: "S1", Request: request.Request{Method: "GET", URL: "http://example.com"}},
					},
				},
				Interval: "0",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			auto: Automation{
				Workflow: workflow.Workflow{
					Steps: []workflow.WorkflowStep{{ID: "s1", Request: request.Request{Method: "GET", URL: "http://example.com"}}},
				},
			},
			wantErr: true,
		},
		{
			name: "missing steps",
			auto: Automation{
				Name:     "a",
				Workflow: workflow.Workflow{Name: "wf"},
			},
			wantErr: true,
		},
		{
			name: "bad interval",
			auto: Automation{
				Name:     "a",
				Workflow: workflow.Workflow{Name: "wf", Steps: []workflow.WorkflowStep{{ID: "s1", Request: request.Request{Method: "GET", URL: "http://example.com"}}}},
				Interval: "not-a-duration",
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.auto.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestAutomation_IsEnabled(t *testing.T) {
	a := Automation{Name: "x", Workflow: workflow.Workflow{Steps: []workflow.WorkflowStep{{ID: "s1"}}}}
	if !a.IsEnabled() {
		t.Fatalf("nil enabled should be true")
	}
	f := false
	a.Enabled = &f
	if a.IsEnabled() {
		t.Fatalf("expected disabled")
	}
	tr := true
	a.Enabled = &tr
	if !a.IsEnabled() {
		t.Fatalf("expected enabled")
	}
}

func TestScheduler_RunOnce(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	auto := &Automation{
		Name: "once",
		Workflow: workflow.Workflow{
			Name: "wf",
			Steps: []workflow.WorkflowStep{
				{ID: "s1", Name: "S1", Request: request.Request{Method: "GET", URL: srv.URL}},
			},
		},
		Interval: "0",
	}
	sched := NewScheduler(nil)
	var reports int
	if err := sched.Run(context.Background(), auto, func(r *workflow.WorkflowReport) { reports++ }); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if reports != 1 {
		t.Fatalf("want 1 report, got %d", reports)
	}
}

func TestScheduler_RunInterval(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	auto := &Automation{
		Name: "interval",
		Workflow: workflow.Workflow{
			Name: "wf",
			Steps: []workflow.WorkflowStep{
				{ID: "s1", Name: "S1", Request: request.Request{Method: "GET", URL: srv.URL}},
			},
		},
		Interval: "10ms",
		MaxRuns:  3,
	}
	sched := NewScheduler(nil)
	var reports int
	start := time.Now()
	if err := sched.Run(context.Background(), auto, func(r *workflow.WorkflowReport) { reports++ }); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if reports != 3 {
		t.Fatalf("want 3 reports, got %d", reports)
	}
	if time.Since(start) < 15*time.Millisecond {
		t.Fatalf("expected at least 15ms for 3 runs at 10ms interval, took %v", time.Since(start))
	}
}

func TestScheduler_Disabled(t *testing.T) {
	f := false
	auto := &Automation{
		Name:     "disabled",
		Workflow: workflow.Workflow{Name: "wf", Steps: []workflow.WorkflowStep{{ID: "s1", Request: request.Request{Method: "GET", URL: "http://example.com"}}}},
		Enabled:  &f,
	}
	sched := NewScheduler(nil)
	if err := sched.Run(context.Background(), auto, nil); err == nil {
		t.Fatalf("expected error for disabled")
	}
}

func TestScheduler_NilAutomation(t *testing.T) {
	sched := NewScheduler(nil)
	if err := sched.Run(context.Background(), nil, nil); err == nil {
		t.Fatalf("expected error for nil")
	}
}
