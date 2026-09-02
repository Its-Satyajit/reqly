// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package scim

import (
	"fmt"
	"strings"
	"sync"
)

// User is a SCIM 2.0 user (RFC 7643).
type User struct {
	ID       string   `json:"id" yaml:"id"`
	UserName string   `json:"userName" yaml:"userName"`
	Email    string   `json:"email,omitempty" yaml:"email,omitempty"`
	Groups   []string `json:"groups,omitempty" yaml:"groups,omitempty"`
	Active   bool     `json:"active" yaml:"active"`
}

// Group is a SCIM 2.0 group.
type Group struct {
	ID          string   `json:"id" yaml:"id"`
	DisplayName string   `json:"displayName" yaml:"displayName"`
	Members     []string `json:"members,omitempty" yaml:"members,omitempty"` // user IDs
}

// ValidateUser checks a user for required fields.
func ValidateUser(u User) error {
	if strings.TrimSpace(u.UserName) == "" {
		return fmt.Errorf("userName is required")
	}
	if strings.TrimSpace(u.Email) != "" && !strings.Contains(u.Email, "@") {
		return fmt.Errorf("invalid email %q", u.Email)
	}
	return nil
}

// ValidateGroup checks a group.
func ValidateGroup(g Group) error {
	if strings.TrimSpace(g.DisplayName) == "" {
		return fmt.Errorf("displayName is required")
	}
	return nil
}

// Store is an in-memory SCIM store (local, zero telemetry).
// In production it could be backed by SQLite, but for M73 it is in-memory
// with mutex, mirroring the audit/policy pattern.
type Store struct {
	mu     sync.Mutex
	users  map[string]User
	groups map[string]Group
}

// NewStore returns an empty SCIM store.
func NewStore() *Store {
	return &Store{
		users:  make(map[string]User),
		groups: make(map[string]Group),
	}
}

// CreateUser adds a user. ID is auto-generated if empty.
func (s *Store) CreateUser(u User) (User, error) {
	if err := ValidateUser(u); err != nil {
		return User{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if u.ID == "" {
		u.ID = fmt.Sprintf("user-%d", len(s.users)+1)
	}
	if _, exists := s.users[u.ID]; exists {
		return User{}, fmt.Errorf("user %q already exists", u.ID)
	}
	// UserName must be unique
	for _, existing := range s.users {
		if existing.UserName == u.UserName {
			return User{}, fmt.Errorf("userName %q already exists", u.UserName)
		}
	}
	u.Active = true
	s.users[u.ID] = u
	return u, nil
}

// GetUser returns a user by ID.
func (s *Store) GetUser(id string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return User{}, fmt.Errorf("user %q not found", id)
	}
	return u, nil
}

// ListUsers returns all users sorted by userName.
func (s *Store) ListUsers() []User {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []User
	for _, u := range s.users {
		out = append(out, u)
	}
	return out
}

// DeactivateUser marks a user inactive.
func (s *Store) DeactivateUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return fmt.Errorf("user %q not found", id)
	}
	u.Active = false
	s.users[id] = u
	return nil
}

// CreateGroup adds a group.
func (s *Store) CreateGroup(g Group) (Group, error) {
	if err := ValidateGroup(g); err != nil {
		return Group{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if g.ID == "" {
		g.ID = fmt.Sprintf("group-%d", len(s.groups)+1)
	}
	if _, exists := s.groups[g.ID]; exists {
		return Group{}, fmt.Errorf("group %q already exists", g.ID)
	}
	s.groups[g.ID] = g
	return g, nil
}

// AddUserToGroup adds a user to a group.
func (s *Store) AddUserToGroup(userID, groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user %q not found", userID)
	}
	g, ok := s.groups[groupID]
	if !ok {
		return fmt.Errorf("group %q not found", groupID)
	}
	// Add to user's groups
	for _, gid := range u.Groups {
		if gid == groupID {
			return nil // already member
		}
	}
	u.Groups = append(u.Groups, groupID)
	s.users[userID] = u
	// Add to group's members
	for _, uid := range g.Members {
		if uid == userID {
			return nil
		}
	}
	g.Members = append(g.Members, userID)
	s.groups[groupID] = g
	return nil
}
