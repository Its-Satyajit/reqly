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
	Name string `json:"name"`
	// Status is one of the Status* constants: added, removed, or changed.
	Status string `json:"status"`
	// Kind is one of the Kind* constants: variable or secret.
	Kind string `json:"kind"`
	// From is the value in the first environment (secrets masked).
	From string `json:"from"`
	// To is the value in the second environment (secrets masked).
	To string `json:"to"`
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
