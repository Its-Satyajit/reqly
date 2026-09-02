// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/workflow"
)

// Automation is a Git-native, self-hosted schedule for a Workflow.
// Stored as YAML (e.g. .reqly/automation/<name>.yaml) and executed locally
// without any cloud. Interval 0 means run once.
type Automation struct {
	Name     string            `json:"name" yaml:"name"`
	Workflow workflow.Workflow `json:"workflow" yaml:"workflow"`
	Interval string            `json:"interval,omitempty" yaml:"interval,omitempty"` // Go duration string, e.g. "10s", "5m"
	Enabled  *bool             `json:"enabled,omitempty" yaml:"enabled,omitempty"`   // nil = true
	MaxRuns  int               `json:"maxRuns,omitempty" yaml:"maxRuns,omitempty"`   // 0 = infinite (until ctx cancel)
}

// IsEnabled reports whether the automation should run. Nil Enabled means enabled.
func (a *Automation) IsEnabled() bool {
	if a.Enabled == nil {
		return true
	}
	return *a.Enabled
}

// IntervalDuration parses Interval as a Go duration. Empty or "0" means 0 (run once).
func (a *Automation) IntervalDuration() (time.Duration, error) {
	if a.Interval == "" || a.Interval == "0" {
		return 0, nil
	}
	d, err := time.ParseDuration(a.Interval)
	if err != nil {
		return 0, fmt.Errorf("interval %q: %w", a.Interval, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("interval must be >= 0")
	}
	return d, nil
}

// Validate checks required fields.
func (a *Automation) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(a.Workflow.Steps) == 0 {
		return fmt.Errorf("workflow.steps is required")
	}
	if _, err := a.IntervalDuration(); err != nil {
		return err
	}
	if a.MaxRuns < 0 {
		return fmt.Errorf("maxRuns must be >= 0")
	}
	return nil
}

// Scheduler executes an Automation's Workflow at the configured interval.
type Scheduler struct {
	client *request.Client
}

// NewScheduler returns a Scheduler. client nil means default client.
func NewScheduler(client *request.Client) *Scheduler {
	if client == nil {
		client = request.NewClient()
	}
	return &Scheduler{client: client}
}

// Run executes the automation's workflow. Interval 0 runs once and returns.
// Otherwise it ticks at interval, calling onReport after each run, until ctx
// is cancelled or MaxRuns is reached. onReport may be nil.
func (s *Scheduler) Run(ctx context.Context, a *Automation, onReport func(*workflow.WorkflowReport)) error {
	if a == nil {
		return fmt.Errorf("automation is nil")
	}
	if err := a.Validate(); err != nil {
		return err
	}
	if !a.IsEnabled() {
		return fmt.Errorf("automation %q is disabled", a.Name)
	}
	interval, err := a.IntervalDuration()
	if err != nil {
		return err
	}
	exec := workflow.NewWorkflowExecutor(s.client)

	runOnce := func() error {
		report, err := exec.Execute(ctx, &a.Workflow, workflow.WorkflowOptions{})
		if err != nil {
			return err
		}
		if onReport != nil {
			onReport(report)
		}
		return nil
	}

	if interval == 0 {
		return runOnce()
	}

	// Immediate first run, then ticker.
	if err := runOnce(); err != nil {
		return err
	}
	if a.MaxRuns == 1 {
		return nil
	}
	count := 1
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := runOnce(); err != nil {
				return err
			}
			count++
			if a.MaxRuns > 0 && count >= a.MaxRuns {
				return nil
			}
		}
	}
}
