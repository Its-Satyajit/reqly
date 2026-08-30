// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package collab

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

// Collaborator is a user with access to a shared workspace.
type Collaborator struct {
	User    string    `json:"user" yaml:"user"`
	Role    string    `json:"role" yaml:"role"` // viewer | editor | admin
	AddedAt time.Time `json:"addedAt" yaml:"addedAt"`
}

// SharedWorkspace is a Git-native shared workspace config.
// Stored as .reqly/collab.yaml (0600), committed to Git for sharing.
type SharedWorkspace struct {
	Path          string         `json:"path" yaml:"path"`
	Collaborators []Collaborator `json:"collaborators" yaml:"collaborators"`
}

// Validate checks a SharedWorkspace.
func Validate(s SharedWorkspace) error {
	if strings.TrimSpace(s.Path) == "" {
		return fmt.Errorf("path is required")
	}
	seen := make(map[string]bool)
	for _, c := range s.Collaborators {
		if strings.TrimSpace(c.User) == "" {
			return fmt.Errorf("collaborator user is required")
		}
		if c.Role != "viewer" && c.Role != "editor" && c.Role != "admin" {
			return fmt.Errorf("invalid role %q for user %q", c.Role, c.User)
		}
		if seen[c.User] {
			return fmt.Errorf("duplicate collaborator %q", c.User)
		}
		seen[c.User] = true
	}
	return nil
}

// AddCollaborator adds or updates a collaborator.
func AddCollaborator(s *SharedWorkspace, user, role string) error {
	if strings.TrimSpace(user) == "" {
		return fmt.Errorf("user is required")
	}
	if role != "viewer" && role != "editor" && role != "admin" {
		return fmt.Errorf("invalid role %q", role)
	}
	for i, c := range s.Collaborators {
		if c.User == user {
			s.Collaborators[i].Role = role
			return nil
		}
	}
	s.Collaborators = append(s.Collaborators, Collaborator{User: user, Role: role, AddedAt: time.Now().UTC()})
	return nil
}

// RemoveCollaborator removes a collaborator.
func RemoveCollaborator(s *SharedWorkspace, user string) error {
	for i, c := range s.Collaborators {
		if c.User == user {
			s.Collaborators = append(s.Collaborators[:i], s.Collaborators[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("collaborator %q not found", user)
}

// IsCollaborator reports whether user is a collaborator.
func IsCollaborator(s SharedWorkspace, user string) bool {
	for _, c := range s.Collaborators {
		if c.User == user {
			return true
		}
	}
	return false
}

// Load loads a SharedWorkspace from a file. Missing file returns empty with path.
func Load(path string) (SharedWorkspace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SharedWorkspace{Path: filepath.Dir(filepath.Dir(path))}, nil
		}
		return SharedWorkspace{}, err
	}
	var s SharedWorkspace
	if err := yaml.Unmarshal(data, &s); err != nil {
		return SharedWorkspace{}, fmt.Errorf("parse collab: %w", err)
	}
	if err := Validate(s); err != nil {
		return SharedWorkspace{}, err
	}
	return s, nil
}

// Save writes a SharedWorkspace to a file with 0600.
func Save(path string, s SharedWorkspace) error {
	if err := Validate(s); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// DefaultPath returns the default collab path for a workspace root.
func DefaultPath(root string) string {
	return filepath.Join(root, ".reqly", "collab.yaml")
}
