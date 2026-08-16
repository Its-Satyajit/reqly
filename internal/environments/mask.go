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
	var kept []string
	for _, v := range values {
		if v != "" {
			kept = append(kept, v)
		}
	}
	return &Masker{values: kept}
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
