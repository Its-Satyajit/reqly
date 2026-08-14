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

package variables

import (
	"fmt"
	"strings"
)

// Scope represents a named layer of variables with an explicit precedence.
type Scope string

const (
	ScopeGlobal      Scope = "global"
	ScopeEnvironment Scope = "environment"
	ScopeCollection  Scope = "collection"
	ScopeFolder      Scope = "folder"
	ScopeRequest     Scope = "request"
	ScopeRuntime     Scope = "runtime"
)

// Precedence returns the scopes ordered from lowest to highest priority.
func Precedence() []Scope {
	return []Scope{ScopeGlobal, ScopeEnvironment, ScopeCollection, ScopeFolder, ScopeRequest, ScopeRuntime}
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
func (s *Set) Interpolate(input string) (string, error) {
	var result strings.Builder
	for {
		start := strings.Index(input, "{{")
		if start == -1 {
			result.WriteString(input)
			break
		}
		result.WriteString(input[:start])
		rest := input[start+2:]
		end := strings.Index(rest, "}}")
		if end == -1 {
			return "", fmt.Errorf("unclosed variable reference in %q", input)
		}
		key := rest[:end]
		value, ok := s.Resolve(key)
		if !ok {
			return "", fmt.Errorf("undefined variable %q", key)
		}
		result.WriteString(value)
		input = rest[end+2:]
	}
	return result.String(), nil
}
