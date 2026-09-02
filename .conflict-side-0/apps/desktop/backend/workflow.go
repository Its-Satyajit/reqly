// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/workflow"
)

// WorkflowRun executes a workflow YAML string and returns the report.
// The YAML may be JSON as well (YAML superset). It is the desktop parity
// for `reqly workflow <file>` and the Goja `reqly.workflow.run` binding.
func (s *AppService) WorkflowRun(yamlStr string) (*workflow.WorkflowReport, error) {
	if yamlStr == "" {
		return nil, fmt.Errorf("workflow yaml is required")
	}
	var wf workflow.Workflow
	if err := yaml.Unmarshal([]byte(yamlStr), &wf); err != nil {
		return nil, fmt.Errorf("parse workflow: %w", err)
	}
	exec := workflow.NewWorkflowExecutor(nil)
	report, err := exec.Execute(context.Background(), &wf, workflow.WorkflowOptions{})
	if err != nil {
		return nil, err
	}
	return report, nil
}
