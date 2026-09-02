// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
)

func TestRBAC_Desktop(t *testing.T) {
	dir := t.TempDir()
	svc := NewAppService()
	svc.root = dir

	roles, err := svc.RBACList()
	if err != nil {
		t.Fatalf("RBACList: %v", err)
	}
	if len(roles) != 3 {
		t.Fatalf("want 3 roles, got %v", roles)
	}
	if err := svc.RBACCheck("unknown", "request.send", "x"); err != nil {
		t.Fatalf("unknown viewer should allow request.send: %v", err)
	}
	if err := svc.RBACCheck("unknown", "workflow.run", "x"); err == nil {
		t.Fatalf("should deny workflow.run for viewer")
	}
}
