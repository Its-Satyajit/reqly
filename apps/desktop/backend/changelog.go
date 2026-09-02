// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/Its-Satyajit/reqly/internal/diffing"
)

// ChangelogResult is the bridge payload for the changelog view.
type ChangelogResult struct {
	Changelog *diffing.Changelog `json:"changelog"`
	Markdown  string             `json:"markdown"`
	JSON      string             `json:"json"`
}

// ChangelogGenerate generates a human-readable API changelog between two specs.
// Paths may be workspace-relative or absolute. Format is "markdown" or "json" (default markdown).
// When failOnBreaking is true and breaking changes exist, it returns an error.
func (s *AppService) ChangelogGenerate(oldPath, newPath string, format string, failOnBreaking bool) (*ChangelogResult, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to generate changelog")
	}
	if oldPath == "" || newPath == "" {
		return nil, fmt.Errorf("old and new spec paths are required")
	}
	absOld, err := s.resolveTestPath(oldPath)
	if err != nil {
		return nil, fmt.Errorf("old spec: %w", err)
	}
	absNew, err := s.resolveTestPath(newPath)
	if err != nil {
		return nil, fmt.Errorf("new spec: %w", err)
	}
	oldBytes, err := os.ReadFile(absOld)
	if err != nil {
		return nil, fmt.Errorf("read old spec: %w", err)
	}
	newBytes, err := os.ReadFile(absNew)
	if err != nil {
		return nil, fmt.Errorf("read new spec: %w", err)
	}
	cl, err := diffing.GenerateChangelog(oldBytes, newBytes)
	if err != nil {
		return nil, err
	}
	if failOnBreaking && len(cl.Breaking) > 0 {
		return nil, fmt.Errorf("breaking changes detected (%d) — fail-on-breaking is set", len(cl.Breaking))
	}
	md := cl.ToMarkdown()
	js, err := cl.ToJSON()
	if err != nil {
		js = "{}"
	}
	if format == "json" {
		// JSON is primary, markdown secondary
		return &ChangelogResult{Changelog: cl, Markdown: md, JSON: js}, nil
	}
	return &ChangelogResult{Changelog: cl, Markdown: md, JSON: js}, nil
}
