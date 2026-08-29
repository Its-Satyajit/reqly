// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
)

func TestCollab_Desktop(t *testing.T) {
	dir := t.TempDir()
	svc := NewAppService()
	svc.root = dir

	if err := svc.CollabAdd("alice", "admin"); err != nil {
		t.Fatalf("CollabAdd: %v", err)
	}
	list, err := svc.CollabList()
	if err != nil {
		t.Fatalf("CollabList: %v", err)
	}
	if len(list) != 1 || list[0].User != "alice" {
		t.Fatalf("unexpected %+v", list)
	}
	if err := svc.CollabRemove("alice"); err != nil {
		t.Fatalf("CollabRemove: %v", err)
	}
	list, _ = svc.CollabList()
	if len(list) != 0 {
		t.Fatalf("want 0, got %v", list)
	}
}
