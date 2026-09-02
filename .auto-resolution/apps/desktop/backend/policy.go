// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/Its-Satyajit/reqly/internal/policy"
)

// PolicyGet returns the current policy for the workspace.
func (s *AppService) PolicyGet() (policy.Policy, error) {
	root := s.root
	if root == "" {
		root = "."
	}
	return policy.Load(policy.DefaultPath(root))
}

// PolicySave writes a policy for the workspace (0600).
func (s *AppService) PolicySave(p policy.Policy) error {
	root := s.root
	if root == "" {
		root = "."
	}
	return policy.Save(policy.DefaultPath(root), p)
}

// PolicyEnforce checks if an action is allowed.
func (s *AppService) PolicyEnforce(action, resource string) error {
	root := s.root
	if root == "" {
		root = "."
	}
	p, err := policy.Load(policy.DefaultPath(root))
	if err != nil {
		return err
	}
	return policy.Enforce(p, action, resource)
}
