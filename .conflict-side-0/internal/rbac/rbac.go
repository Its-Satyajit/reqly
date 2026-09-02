// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package rbac

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Role defines a set of permissions.
type Role struct {
	Name        string   `json:"name" yaml:"name"`
	Permissions []string `json:"permissions" yaml:"permissions"`
}

// RBAC is a local role-based access control config.
// Stored as .reqly/rbac.yaml (0600), Git-native when committed.
type RBAC struct {
	Roles     map[string]Role   `json:"roles" yaml:"roles"`
	UserRoles map[string]string `json:"userRoles" yaml:"userRoles"` // user -> role
}

// DefaultRBAC returns the permissive default with admin/editor/viewer.
func DefaultRBAC() RBAC {
	return RBAC{
		Roles: map[string]Role{
			"admin":  {Name: "admin", Permissions: []string{"*"}},
			"editor": {Name: "editor", Permissions: []string{"request.send", "workflow.run", "automation.run", "collection.run", "theme.import", "theme.export"}},
			"viewer": {Name: "viewer", Permissions: []string{"request.send", "theme.export"}},
		},
		UserRoles: map[string]string{},
	}
}

// Validate checks RBAC for semantic errors.
func Validate(r RBAC) error {
	if len(r.Roles) == 0 {
		return fmt.Errorf("at least one role is required")
	}
	for name, role := range r.Roles {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("role name cannot be empty")
		}
		if role.Name != name {
			return fmt.Errorf("role key %q != name %q", name, role.Name)
		}
		if len(role.Permissions) == 0 {
			return fmt.Errorf("role %q must have at least one permission", name)
		}
		for _, p := range role.Permissions {
			if strings.TrimSpace(p) == "" {
				return fmt.Errorf("role %q has empty permission", name)
			}
		}
	}
	for user, role := range r.UserRoles {
		if strings.TrimSpace(user) == "" {
			return fmt.Errorf("user cannot be empty")
		}
		if _, ok := r.Roles[role]; !ok {
			return fmt.Errorf("user %q has unknown role %q", user, role)
		}
	}
	return nil
}

// Can checks if a user can perform an action. Unknown user uses "viewer" if exists, else deny.
func Can(r RBAC, user, action string) bool {
	roleName, ok := r.UserRoles[user]
	if !ok {
		// Default to viewer if present, else no access
		if _, hasViewer := r.Roles["viewer"]; hasViewer {
			roleName = "viewer"
		} else {
			return false
		}
	}
	role, ok := r.Roles[roleName]
	if !ok {
		return false
	}
	for _, p := range role.Permissions {
		if p == "*" || p == action {
			return true
		}
	}
	return false
}

// Enforce returns error when denied, nil when allowed.
func Enforce(r RBAC, user, action, resource string) error {
	if Can(r, user, action) {
		return nil
	}
	return fmt.Errorf("user %q denied for action %q on %q", user, action, resource)
}

// ListRoles returns sorted role names.
func ListRoles(r RBAC) []string {
	var out []string
	for k := range r.Roles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Load loads RBAC from a file. Missing file returns DefaultRBAC.
func Load(path string) (RBAC, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultRBAC(), nil
		}
		return RBAC{}, err
	}
	var r RBAC
	if err := yaml.Unmarshal(data, &r); err != nil {
		return RBAC{}, fmt.Errorf("parse rbac: %w", err)
	}
	if err := Validate(r); err != nil {
		return RBAC{}, err
	}
	return r, nil
}

// Save writes RBAC to a file with 0600.
func Save(path string, r RBAC) error {
	if err := Validate(r); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	data, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// DefaultPath returns the default RBAC path for a workspace root.
func DefaultPath(root string) string {
	return filepath.Join(root, ".reqly", "rbac.yaml")
}
