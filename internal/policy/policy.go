// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Policy is a local organization policy file (e.g. .reqly/policy.yaml, 0600).
// It is Git-native when committed, but the file itself is 0600 for secrets.
// Policies are local-only, zero telemetry, and enforced by the core before
// sensitive actions.
type Policy struct {
	RequireAudit      bool     `json:"requireAudit" yaml:"requireAudit"`
	MaxWorkflowSteps  int      `json:"maxWorkflowSteps,omitempty" yaml:"maxWorkflowSteps,omitempty"`
	AllowedActions    []string `json:"allowedActions,omitempty" yaml:"allowedActions,omitempty"`
	RequireAuth       bool     `json:"requireAuth,omitempty" yaml:"requireAuth,omitempty"`
	AllowCustomThemes bool     `json:"allowCustomThemes,omitempty" yaml:"allowCustomThemes,omitempty"`
}

// DefaultPolicy returns the permissive default (no restrictions).
func DefaultPolicy() Policy {
	return Policy{
		RequireAudit:      false,
		MaxWorkflowSteps:  0, // 0 = no limit
		AllowedActions:    nil,
		RequireAuth:       false,
		AllowCustomThemes: true,
	}
}

// Validate checks a policy for semantic errors.
func Validate(p Policy) error {
	if p.MaxWorkflowSteps < 0 {
		return fmt.Errorf("maxWorkflowSteps must be >= 0")
	}
	for _, a := range p.AllowedActions {
		if strings.TrimSpace(a) == "" {
			return fmt.Errorf("allowedActions cannot contain empty")
		}
	}
	return nil
}

// Enforce checks whether an action on a resource is allowed by the policy.
// It returns an error when denied, nil when allowed.
func Enforce(p Policy, action, resource string) error {
	if len(p.AllowedActions) > 0 {
		allowed := false
		for _, a := range p.AllowedActions {
			if a == action || a == "*" {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("action %q denied by policy (resource %q)", action, resource)
		}
	}
	return nil
}

// EnforceWorkflow checks workflow step count against MaxWorkflowSteps.
func EnforceWorkflow(p Policy, stepCount int) error {
	if p.MaxWorkflowSteps > 0 && stepCount > p.MaxWorkflowSteps {
		return fmt.Errorf("workflow has %d steps, exceeds policy max %d", stepCount, p.MaxWorkflowSteps)
	}
	return nil
}

// Load loads a policy from a file (YAML or JSON). Missing file returns DefaultPolicy.
func Load(path string) (Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultPolicy(), nil
		}
		return Policy{}, err
	}
	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Policy{}, fmt.Errorf("parse policy: %w", err)
	}
	if err := Validate(p); err != nil {
		return Policy{}, err
	}
	return p, nil
}

// Save writes a policy to a file with 0600.
func Save(path string, p Policy) error {
	if err := Validate(p); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// DefaultPath returns the default policy path for a workspace root.
func DefaultPath(root string) string {
	return filepath.Join(root, ".reqly", "policy.yaml")
}
