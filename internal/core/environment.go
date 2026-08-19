// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package core

import (
	"fmt"
	"sort"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/environments"
)

// Environment is the bridge-friendly view of an environment: its variables
// and the *names* of its secrets. Secret values never cross the service
// boundary — they stay on disk and are only written, never read back.
type Environment struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Variables   map[string]string `json:"variables"`
	Secrets     []string          `json:"secrets"`
}

// EnvListResponse is the result of listing a workspace's environments plus
// the currently active one.
type EnvListResponse struct {
	Active       string        `json:"active"`
	Environments []Environment `json:"environments"`
}

// EnvironmentService exposes workspace environment management (list, read,
// set active) to front-ends on the same seams the CLI uses: the
// internal/environments package for files and the workspace descriptor for
// the active selection. The service is UI-agnostic and never exposes secret
// values to callers.
type EnvironmentService struct {
	root string
}

// NewEnvironmentService returns an EnvironmentService rooted at the given
// workspace root ("" means no workspace; listing then returns empty).
func NewEnvironmentService(root string) *EnvironmentService {
	return &EnvironmentService{root: root}
}

// List returns the workspace's environments (name, description, variables,
// secret names) in lexical order plus the active environment name from the
// workspace descriptor. An empty root or a missing environments/ directory
// yields an empty list without error.
func (s *EnvironmentService) List() (*EnvListResponse, error) {
	names, err := environments.List(s.root)
	if err != nil {
		// A missing environments/ directory is valid for an env-less project:
		// report an empty list (but still surface the active descriptor name).
		if s.root == "" {
			return &EnvListResponse{}, nil
		}
		envDir, _ := environments.Discover(s.root)
		if envDir == "" {
			return &EnvListResponse{Active: collections.WorkspaceEnvironment(s.root)}, nil
		}
		return nil, err
	}

	out := make([]Environment, 0, len(names))
	for _, name := range names {
		env, err := environments.Read(name, s.root)
		if err != nil {
			return nil, err
		}
		out = append(out, environmentDTO(env))
	}
	return &EnvListResponse{
		Active:       collections.WorkspaceEnvironment(s.root),
		Environments: out,
	}, nil
}

// Read returns a single environment by name, without its secret values.
func (s *EnvironmentService) Read(name string) (*Environment, error) {
	if s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to manage environments")
	}
	env, err := environments.Read(name, s.root)
	if err != nil {
		return nil, err
	}
	out := environmentDTO(env)
	return &out, nil
}

// Create writes a new environment (name + optional description + variables)
// to environments/<name>.yaml. An empty or unsafe name, or an environment
// that already exists, is an error. Without a workspace, this is an error.
func (s *EnvironmentService) Create(name, description string, variables map[string]string) error {
	if s.root == "" {
		return fmt.Errorf("no workspace found: open a reqly workspace to create an environment")
	}
	if _, err := environments.Read(name, s.root); err == nil {
		return fmt.Errorf("environment %q already exists", name)
	}
	env := &environments.Environment{
		Name:        name,
		Description: description,
		Variables:   variables,
	}
	return environments.Save(env, s.root)
}

// SetActive persists name as the workspace's active environment in the
// descriptor. An empty name clears the selection. Without a workspace, this
// is an error.
func (s *EnvironmentService) SetActive(name string) error {
	if s.root == "" {
		return fmt.Errorf("no workspace found: open a reqly workspace to set an active environment")
	}
	if name != "" {
		if _, err := environments.Read(name, s.root); err != nil {
			return err
		}
		if _, err := collections.LoadWorkspace(s.root); err != nil {
			return fmt.Errorf("no workspace descriptor found: %w", err)
		}
	}
	return collections.SetWorkspaceEnvironment(s.root, name)
}

// environmentDTO maps an internal environment to its bridge-friendly view,
// exposing secret *names* only (sorted) and never secret values.
func environmentDTO(env *environments.Environment) Environment {
	secrets := make([]string, 0, len(env.Secrets))
	for name := range env.Secrets {
		secrets = append(secrets, name)
	}
	sort.Strings(secrets)
	return Environment{
		Name:        env.Name,
		Description: env.Description,
		Variables:   env.Variables,
		Secrets:     secrets,
	}
}
