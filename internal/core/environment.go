// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"os"
	"path/filepath"
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
	if s == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to list environments")
	}
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
	if s == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to manage environments")
	}
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
	if s == nil {
		return fmt.Errorf("no workspace found: open a reqly workspace to create an environment")
	}
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

// Update rewrites an existing environment's description and variables,
// preserving its secrets on disk (secret values are never read back through
// the service). A missing environment or workspace is an error, as is a
// variable key that collides with an existing secret name.
func (s *EnvironmentService) Update(name, description string, variables map[string]string) error {
	if s == nil {
		return fmt.Errorf("no workspace found: open a reqly workspace to update an environment")
	}
	if s.root == "" {
		return fmt.Errorf("no workspace found: open a reqly workspace to update an environment")
	}
	existing, err := environments.Read(name, s.root)
	if err != nil {
		return err
	}
	for key := range variables {
		if _, dup := existing.Secrets[key]; dup {
			return fmt.Errorf("key %q is defined in both variables and secrets", key)
		}
	}
	env := &environments.Environment{
		Name:        name,
		Description: description,
		Variables:   variables,
		Secrets:     existing.Secrets,
	}
	return environments.Save(env, s.root)
}

// UpdateSecrets changes an environment's secrets without ever reading their
// values back to the caller. `values` holds only the secrets the user changed
// (existing ones keep their on-disk values); `remove` lists secret names to
// delete. A missing environment or workspace is an error.
func (s *EnvironmentService) UpdateSecrets(name string, values map[string]string, remove []string) error {
	if s == nil {
		return fmt.Errorf("no workspace found: open a reqly workspace to edit secrets")
	}
	if s.root == "" {
		return fmt.Errorf("no workspace found: open a reqly workspace to edit secrets")
	}
	existing, err := environments.Read(name, s.root)
	if err != nil {
		return err
	}
	for key := range values {
		if _, dup := existing.Variables[key]; dup {
			return fmt.Errorf("key %q is defined in both variables and secrets", key)
		}
	}
	secrets := existing.Secrets
	if secrets == nil {
		secrets = make(map[string]string, len(values)+len(remove))
	}
	for key, value := range values {
		secrets[key] = value
	}
	for _, key := range remove {
		delete(secrets, key)
	}
	env := &environments.Environment{
		Name:        existing.Name,
		Description: existing.Description,
		Variables:   existing.Variables,
		Secrets:     secrets,
	}
	return environments.Save(env, s.root)
}

// Delete removes an environment's file. If the deleted environment is the
// workspace's active one, the descriptor's selection is cleared. A missing
// environment or workspace is an error.
func (s *EnvironmentService) Delete(name string) error {
	if s == nil {
		return fmt.Errorf("no workspace found: open a reqly workspace to delete an environment")
	}
	if s.root == "" {
		return fmt.Errorf("no workspace found: open a reqly workspace to delete an environment")
	}
	if _, err := environments.Read(name, s.root); err != nil {
		return err
	}
	envDir, err := environments.Discover(s.root)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(envDir, name+".yaml")); err != nil {
		return fmt.Errorf("remove environment file %q: %w", name, err)
	}
	if collections.WorkspaceEnvironment(s.root) == name {
		return collections.SetWorkspaceEnvironment(s.root, "")
	}
	return nil
}

// SetActive persists name as the workspace's active environment in the
// descriptor. An empty name clears the selection. Without a workspace, this
// is an error.
func (s *EnvironmentService) SetActive(name string) error {
	if s == nil {
		return fmt.Errorf("no workspace found: open a reqly workspace to set an active environment")
	}
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
