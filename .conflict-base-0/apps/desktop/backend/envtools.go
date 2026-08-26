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

package main

import (
	"fmt"
	"sort"

	"github.com/Its-Satyajit/reqly/internal/environments"
)

// EnvDiffResult is the key-level diff of two named environments, with secret
// values masked exactly like `reqly env diff`.
type EnvDiffResult struct {
	EnvA  string                 `json:"envA"`
	EnvB  string                 `json:"envB"`
	Diffs []environments.KeyDiff `json:"diffs"`
}

// EnvValidateResult carries one environment's validation findings.
type EnvValidateResult struct {
	Env    string               `json:"env"`
	Issues []environments.Issue `json:"issues"`
}

// CrossEnvGap reports one variable key whose presence is inconsistent across
// environments — defined in some, missing in others.
type CrossEnvGap struct {
	Key       string   `json:"key"`
	PresentIn []string `json:"presentIn"`
	MissingIn []string `json:"missingIn"`
}

// resolveWorkspaceEnv loads a named environment rooted at the workspace.
func (s *AppService) resolveWorkspaceEnv(name string) (*environments.Environment, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to inspect environments")
	}
	env, err := environments.Read(name, s.root)
	if err != nil {
		return nil, fmt.Errorf("load environment %q: %w", name, err)
	}
	return env, nil
}

// EnvDiff compares two environments by name (masked secrets).
func (s *AppService) EnvDiff(envA string, envB string) (*EnvDiffResult, error) {
	a, err := s.resolveWorkspaceEnv(envA)
	if err != nil {
		return nil, err
	}
	b, err := s.resolveWorkspaceEnv(envB)
	if err != nil {
		return nil, err
	}
	return &EnvDiffResult{
		EnvA:  envA,
		EnvB:  envB,
		Diffs: environments.Diff(a, b),
	}, nil
}

// EnvValidate runs `reqly env validate` for one environment, scanning the
// workspace for undefined {{variable}} references.
func (s *AppService) EnvValidate(name string) (*EnvValidateResult, error) {
	env, err := s.resolveWorkspaceEnv(name)
	if err != nil {
		return nil, err
	}
	issues, err := environments.Validate(env, s.root)
	if err != nil {
		return nil, fmt.Errorf("validate %q: %w", name, err)
	}
	return &EnvValidateResult{Env: name, Issues: issues}, nil
}

// EnvCrossValidate checks variable-key consistency across every environment
// in the workspace: a key present in some but missing in others is a gap.
func (s *AppService) EnvCrossValidate() ([]CrossEnvGap, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to validate environments")
	}
	names, err := environments.List(s.root)
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}
	presence := map[string]map[string]bool{} // key -> env -> defined
	for _, name := range names {
		env, rerr := environments.Read(name, s.root)
		if rerr != nil {
			continue // unreadable envs surface through EnvValidate instead
		}
		for key := range env.Variables {
			if presence[key] == nil {
				presence[key] = map[string]bool{}
			}
			presence[key][name] = true
		}
	}
	var gaps []CrossEnvGap
	for key, envs := range presence {
		var presentIn, missingIn []string
		for _, name := range names {
			if envs[name] {
				presentIn = append(presentIn, name)
			} else {
				missingIn = append(missingIn, name)
			}
		}
		if len(missingIn) > 0 {
			sort.Strings(missingIn)
			gaps = append(gaps, CrossEnvGap{Key: key, PresentIn: presentIn, MissingIn: missingIn})
		}
	}
	sort.Slice(gaps, func(i, j int) bool { return gaps[i].Key < gaps[j].Key })
	return gaps, nil
}
