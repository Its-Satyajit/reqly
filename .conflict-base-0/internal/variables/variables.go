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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"regexp"
	"strings"
	"time"
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

// tagGenerator produces dynamic values for template tags.
type tagGenerator interface {
	Generate(tag string, args []string) (string, bool)
}

var globalTagGen tagGenerator = defaultTagGenerator{}

// SetTagGeneratorForTest sets the global tag generator for tests. Pass nil to restore default.
func SetTagGeneratorForTest(g tagGenerator) {
	if g == nil {
		globalTagGen = defaultTagGenerator{}
	} else {
		globalTagGen = g
	}
}

type defaultTagGenerator struct{}

func (d defaultTagGenerator) Generate(tag string, args []string) (string, bool) {
	switch tag {
	case "uuid", "guid":
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return "", false
		}
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		hexStr := hex.EncodeToString(b)
		uuid := fmt.Sprintf("%s-%s-%s-%s-%s", hexStr[0:8], hexStr[8:12], hexStr[12:16], hexStr[16:20], hexStr[20:32])
		return uuid, true
	case "timestamp":
		return fmt.Sprintf("%d", time.Now().Unix()), true
	case "isoTimestamp", "iso-timestamp":
		return time.Now().UTC().Format(time.RFC3339Nano), true
	case "randomInt":
		// ignore args in M23, default 0-1000
		return fmt.Sprintf("%d", mrand.Intn(1001)), true
	case "randomString":
		const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
		n := 8
		// args ignored in M23
		b := make([]byte, n)
		for i := range b {
			b[i] = chars[mrand.Intn(len(chars))]
		}
		return string(b), true
	case "now":
		return fmt.Sprintf("%d", time.Now().UnixMilli()), true
	default:
		return "", false
	}
}

var tagRegexp = regexp.MustCompile(`\{\{\$([^}\s]+)(?:\s+([^}]*))?\}\}`)

func expandDynamicTags(s string) string {
	return tagRegexp.ReplaceAllStringFunc(s, func(m string) string {
		sub := tagRegexp.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		tag := sub[1]
		var args []string
		if len(sub) >= 3 && sub[2] != "" {
			args = strings.Fields(sub[2])
		}
		if val, ok := globalTagGen.Generate(tag, args); ok {
			return val
		}
		return m
	})
}

// UnknownDynamicTags returns the list of unknown {{$tag}} names found in s (for saveWarnings).
func UnknownDynamicTags(s string) []string {
	matches := tagRegexp.FindAllStringSubmatch(s, -1)
	var out []string
	seen := map[string]bool{}
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		tag := m[1]
		if _, ok := globalTagGen.Generate(tag, nil); !ok && !seen[tag] {
			seen[tag] = true
			out = append(out, tag)
		}
	}
	return out
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

// Interpolate replaces {{key}} placeholders using resolved scope precedence and expands {{$tag}} dynamic values.
// Placeholders inside substituted values are also resolved, so variables can
// reference other variables ({{a}} → "Bearer {{b}}" → "Bearer tok"). Dynamic tags generate per occurrence fresh.
func (s *Set) Interpolate(input string) (string, error) {
	const maxPasses = 16
	for pass := 0; pass < maxPasses; pass++ {
		expanded := expandDynamicTags(input)
		tagsChanged := expanded != input
		input = expanded
		output, varChanged, err := s.interpolateOnce(input)
		if err != nil {
			return "", err
		}
		input = output
		if !tagsChanged && !varChanged {
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
		if strings.HasPrefix(key, "$") {
			// dynamic tag left literal (unknown) — not a variable, keep as-is
			result.WriteString("{{" + key + "}}")
			input = rest[end+2:]
			continue
		}
		value, ok := s.Resolve(key)
		if !ok {
			return "", changed, fmt.Errorf("undefined variable %q", key)
		}
		result.WriteString(value)
		changed = true
		input = rest[end+2:]
	}
}
