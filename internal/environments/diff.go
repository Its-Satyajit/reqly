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

import "sort"

// KeyDiff status values.
const (
	StatusAdded   = "added"
	StatusRemoved = "removed"
	StatusChanged = "changed"
)

// KeyDiff kind values.
const (
	KindVariable = "variable"
	KindSecret   = "secret"
)

// KeyDiff describes a single key-level change between two environments.
type KeyDiff struct {
	// Name is the variable or secret key.
	Name string
	// Status is one of the Status* constants: added, removed, or changed.
	Status string
	// Kind is one of the Kind* constants: variable or secret.
	Kind string
	// From is the value in the first environment (secrets masked).
	From string
	// To is the value in the second environment (secrets masked).
	To string
}

// maskSecret replaces a secret value with the masking sentinel. A blank value
// is left blank so empty-vs-empty never renders as a [SECRET] pair.
func maskSecret(v string) string {
	if v == "" {
		return ""
	}
	return MaskedSecret
}

// Diff compares two environments key by key and returns the differences in
// sorted key order. Secret values are always rendered as [SECRET]. Keys moving
// between the variables and secrets sections are reported as changed.
func Diff(a, b *Environment) []KeyDiff {
	merged := make(map[string]struct {
		aVal, aKind string
		bVal, bKind string
	})
	for k, v := range a.Variables {
		s := merged[k]
		s.aVal, s.aKind = v, KindVariable
		merged[k] = s
	}
	for k, v := range a.Secrets {
		s := merged[k]
		s.aVal, s.aKind = v, KindSecret
		merged[k] = s
	}
	for k, v := range b.Variables {
		s := merged[k]
		s.bVal, s.bKind = v, KindVariable
		merged[k] = s
	}
	for k, v := range b.Secrets {
		s := merged[k]
		s.bVal, s.bKind = v, KindSecret
		merged[k] = s
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var diffs []KeyDiff
	for _, k := range keys {
		s := merged[k]
		secret := s.aKind == KindSecret || s.bKind == KindSecret
		from := s.aVal
		if secret {
			from = maskSecret(from)
		}
		to := s.bVal
		if secret {
			to = maskSecret(to)
		}
		switch {
		case s.aKind == "":
			diffs = append(diffs, KeyDiff{Name: k, Status: StatusAdded, Kind: s.bKind, From: "", To: to})
		case s.bKind == "":
			diffs = append(diffs, KeyDiff{Name: k, Status: StatusRemoved, Kind: s.aKind, From: from, To: ""})
		case s.aKind != s.bKind || s.aVal != s.bVal:
			diffs = append(diffs, KeyDiff{Name: k, Status: StatusChanged, Kind: s.bKind, From: from, To: to})
		}
	}
	return diffs
}
