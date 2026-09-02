// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAudit_Desktop(t *testing.T) {
	dir := t.TempDir()
	svc := NewAppService()
	svc.root = dir
	// Ensure .reqly exists via AuditAdd
	entry, err := svc.AuditAdd("request.send", "users/list", "GET /users")
	if err != nil {
		t.Fatalf("AuditAdd: %v", err)
	}
	if entry.ID == "" {
		t.Fatalf("expected ID")
	}
	// List
	list, err := svc.AuditList()
	if err != nil {
		t.Fatalf("AuditList: %v", err)
	}
	if len(list) != 1 || list[0].Action != "request.send" {
		t.Fatalf("unexpected list %+v", list)
	}
	// Export
	exp, err := svc.AuditExport()
	if err != nil {
		t.Fatalf("AuditExport: %v", err)
	}
	if exp == "" {
		t.Fatalf("expected export")
	}
	// Clear
	if err := svc.AuditClear(); err != nil {
		t.Fatalf("AuditClear: %v", err)
	}
	list, _ = svc.AuditList()
	if len(list) != 0 {
		t.Fatalf("want 0 after clear, got %d", len(list))
	}
	// Check file perm 0600
	fi, err := os.Stat(filepath.Join(dir, ".reqly", "audit.log"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("want 0600, got %o", fi.Mode().Perm())
	}
}
