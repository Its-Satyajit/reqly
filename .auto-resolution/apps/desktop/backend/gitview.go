package main

import (
	"fmt"

	"github.com/Its-Satyajit/reqly/internal/git"
)

// GitStatus returns porcelain status lines.
func (s *AppService) GitStatus() ([]string, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a git workspace to view status")
	}
	return git.Status(s.root)
}

// GitDiff returns git diff (staged when staged true).
func (s *AppService) GitDiff(staged bool) (string, error) {
	if s == nil || s.root == "" {
		return "", fmt.Errorf("no workspace found: open a git workspace to view diff")
	}
	return git.Diff(s.root, staged)
}

// GitLog returns recent log lines.
func (s *AppService) GitLog(limit, offset int) ([]string, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a git workspace to view log")
	}
	return git.Log(s.root, limit, offset)
}

// GitCommit stages files and commits.
func (s *AppService) GitCommit(message string, files []string) error {
	if s == nil || s.root == "" {
		return fmt.Errorf("no workspace found: open a git workspace to commit")
	}
	return git.Commit(s.root, message, files)
}
