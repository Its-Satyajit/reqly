// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/Its-Satyajit/reqly/internal/audit"
)

// AuditList returns audit entries for the current workspace.
func (s *AppService) AuditList() ([]audit.Entry, error) {
	root := s.root
	if root == "" {
		root = "."
	}
	store, err := audit.NewStore(root)
	if err != nil {
		return nil, err
	}
	return store.List()
}

// AuditAdd appends an entry and returns it.
func (s *AppService) AuditAdd(action, resource, details string) (audit.Entry, error) {
	root := s.root
	if root == "" {
		root = "."
	}
	store, err := audit.NewStore(root)
	if err != nil {
		return audit.Entry{}, err
	}
	return store.Add(audit.Entry{Action: action, Resource: resource, Details: details})
}

// AuditClear truncates the log.
func (s *AppService) AuditClear() error {
	root := s.root
	if root == "" {
		root = "."
	}
	store, err := audit.NewStore(root)
	if err != nil {
		return err
	}
	return store.Clear()
}

// AuditExport returns the audit log as JSONL string for downloads.
func (s *AppService) AuditExport() (string, error) {
	entries, err := s.AuditList()
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", nil
	}
	var out string
	for _, e := range entries {
		out += fmt.Sprintf("%s %s %s %s\n", e.Timestamp.Format("2006-01-02T15:04:05Z07:00"), e.Action, e.Resource, e.Details)
	}
	return out, nil
}
