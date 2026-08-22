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

package variables

import (
	"fmt"
	"strings"
)

// Scope represents a named layer of variables with an explicit precedence.
type Scope string

const (
	ScopeProcessEnv  Scope = "process-env"
	ScopeGlobal      Scope = "global"
	ScopeEnvironment Scope = "environment"
	ScopeCollection  Scope = "collection"
	ScopeFolder      Scope = "folder"
	ScopeRequest     Scope = "request"
	ScopeRuntime     Scope = "runtime"
)

// Precedence returns the scopes ordered from lowest to highest priority.
func Precedence() []Scope {
	return []Scope{ScopeProcessEnv, ScopeGlobal, ScopeEnvironment, ScopeCollection, ScopeFolder, ScopeRequest, ScopeRuntime}
}

// Set holds variables grouped by scope.
type Set struct {
	values map[Scope]map[string]string
}

// NewSet returns an empty variable set.
func NewSet() *Set {
	return &Set{values: make(map[Scope]map[string]string)}
}

// Set stores value for key in the given scope.
func (s *Set) Set(scope Scope, key, value string) {
	if s.values[scope] == nil {
		s.values[scope] = make(map[string]string)
	}
	s.values[scope][key] = value
}

// Get returns the raw value for key in scope without precedence resolution.
func (s *Set) Get(scope Scope, key string) (string, bool) {
	value, ok := s.values[scope][key]
	return value, ok
}

// Range calls fn for every key in scope. Iteration order is unspecified.
func (s *Set) Range(scope Scope, fn func(key, value string)) {
	for key, value := range s.values[scope] {
		fn(key, value)
	}
}

// Clone returns a deep copy of the set.
func (s *Set) Clone() *Set {
	clone := NewSet()
	for scope, values := range s.values {
		for key, value := range values {
			clone.Set(scope, key, value)
		}
	}
	return clone
}

// Resolve looks up key respecting scope precedence (highest priority wins).
func (s *Set) Resolve(key string) (string, bool) {
	precedence := Precedence()
	for i := len(precedence) - 1; i >= 0; i-- {
		scope := precedence[i]
		if value, ok := s.values[scope][key]; ok {
			return value, true
		}
	}
	return "", false
}

// Interpolate replaces {{key}} placeholders using resolved scope precedence.
// Placeholders inside substituted values are also resolved, so variables can
// reference other variables ({{a}} → "Bearer {{b}}" → "Bearer tok").
func (s *Set) Interpolate(input string) (string, error) {
	const maxPasses = 16
	for pass := 0; pass < maxPasses; pass++ {
		output, changed, err := s.interpolateOnce(input)
		if err != nil {
			return "", err
		}
		input = output
		if !changed {
			return input, nil
		}
	}
	return input, nil
}

// interpolateOnce performs a single replacement pass, reporting whether any
// placeholder was substituted.
func (s *Set) interpolateOnce(input string) (string, bool, error) {
	var result strings.Builder
	changed := false
	for {
		start := strings.Index(input, "{{")
		if start == -1 {
			result.WriteString(input)
			return result.String(), changed, nil
		}
		result.WriteString(input[:start])
		rest := input[start+2:]
		end := strings.Index(rest, "}}")
		if end == -1 {
			return "", changed, fmt.Errorf("unclosed variable reference in %q", input)
		}
		key := rest[:end]
		value, ok := s.Resolve(key)
		if !ok {
			return "", changed, fmt.Errorf("undefined variable %q", key)
		}
		result.WriteString(value)
		changed = true
		input = rest[end+2:]
	}
}
