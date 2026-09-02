// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/Its-Satyajit/reqly/internal/rbac"
)

// RBACList returns RBAC roles for the workspace.
func (s *AppService) RBACList() ([]string, error) {
	root := s.root
	if root == "" {
		root = "."
	}
	r, err := rbac.Load(rbac.DefaultPath(root))
	if err != nil {
		return nil, err
	}
	return rbac.ListRoles(r), nil
}

// RBACCheck checks if a user can perform an action.
func (s *AppService) RBACCheck(user, action, resource string) error {
	root := s.root
	if root == "" {
		root = "."
	}
	r, err := rbac.Load(rbac.DefaultPath(root))
	if err != nil {
		return err
	}
	return rbac.Enforce(r, user, action, resource)
}

// RBACGet returns the full RBAC config.
func (s *AppService) RBACGet() (rbac.RBAC, error) {
	root := s.root
	if root == "" {
		root = "."
	}
	return rbac.Load(rbac.DefaultPath(root))
}
