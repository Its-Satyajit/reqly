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

package environments

import "strings"

// MaskedSecret is the sentinel that replaces secret values in output.
const MaskedSecret = "[SECRET]"

// Masker redacts sensitive values from output. It combines environment secret
// values with values loaded from a .env file, so both are masked even when a
// key name does not look secret.
type Masker struct {
	values []string
}

// NewMasker returns a Masker for the given values. Empty values are dropped.
func NewMasker(values ...string) *Masker {
	m := &Masker{}
	m.Add(values...)
	return m
}

// Add includes extra sensitive values (e.g. resolved auth credentials) in the
// mask. Empty values are dropped.
func (m *Masker) Add(values ...string) {
	for _, v := range values {
		if v != "" {
			m.values = append(m.values, v)
		}
	}
}

// Mask replaces every sensitive value in text with [SECRET]. Values are
// matched longest-first so a value that is a prefix of another never gets
// shadowed. Non-sensitive text is left untouched.
func (m *Masker) Mask(text string) string {
	if m == nil || len(m.values) == 0 {
		return text
	}
	values := append([]string(nil), m.values...)
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if len(values[j]) > len(values[i]) {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
	masked := text
	for _, value := range values {
		masked = strings.ReplaceAll(masked, value, MaskedSecret)
	}
	return masked
}

// SecretValues returns the unique secret values of the environment.
func (e *Environment) SecretValues() []string {
	seen := make(map[string]struct{}, len(e.Secrets))
	for _, value := range e.Secrets {
		if value == "" {
			continue
		}
		seen[value] = struct{}{}
	}
	values := make([]string, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	return values
}
