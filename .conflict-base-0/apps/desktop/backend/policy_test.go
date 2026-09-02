// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/Its-Satyajit/reqly/internal/policy"
)

func TestPolicy_Desktop(t *testing.T) {
	dir := t.TempDir()
	svc := NewAppService()
	svc.root = dir

	// Default should be permissive
	p, err := svc.PolicyGet()
	if err != nil {
		t.Fatalf("PolicyGet: %v", err)
	}
	if p.MaxWorkflowSteps != 0 {
		t.Fatalf("unexpected %+v", p)
	}
	// Save a restrictive policy
	toSave := policy.Policy{AllowedActions: []string{"request.send"}, MaxWorkflowSteps: 2}
	if err := svc.PolicySave(toSave); err != nil {
		t.Fatalf("PolicySave: %v", err)
	}
	// Enforce allowed
	if err := svc.PolicyEnforce("request.send", "users/list"); err != nil {
		t.Fatalf("should allow: %v", err)
	}
	// Enforce denied
	if err := svc.PolicyEnforce("theme.import", "x"); err == nil {
		t.Fatalf("should deny")
	}
	// Enforce workflow steps via policy
	loaded, _ := svc.PolicyGet()
	if err := policy.EnforceWorkflow(loaded, 3); err == nil {
		t.Fatalf("should deny 3 steps >2")
	}
}
