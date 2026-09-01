// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/automation"
	"github.com/Its-Satyajit/reqly/internal/workflow"
)

// AutomationRun executes an automation YAML string once (ignoring its interval)
// and returns the workflow report. Desktop parity for `reqly automation run`.
func (s *AppService) AutomationRun(yamlStr string) (*workflow.WorkflowReport, error) {
	if yamlStr == "" {
		return nil, fmt.Errorf("automation yaml is required")
	}
	var auto automation.Automation
	if err := yaml.Unmarshal([]byte(yamlStr), &auto); err != nil {
		return nil, fmt.Errorf("parse automation: %w", err)
	}
	// For desktop single-run, force interval 0 so scheduler runs once.
	auto.Interval = "0"
	if err := auto.Validate(); err != nil {
		return nil, err
	}
	sched := automation.NewScheduler(nil)
	var report *workflow.WorkflowReport
	if err := sched.Run(context.Background(), &auto, func(r *workflow.WorkflowReport) { report = r }); err != nil {
		return nil, err
	}
	return report, nil
}
