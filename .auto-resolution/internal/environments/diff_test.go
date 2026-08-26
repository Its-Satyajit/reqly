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

import "testing"

func TestDiffEnvironments(t *testing.T) {
	a := &Environment{
		Variables: map[string]string{
			"KEEP":   "same",
			"CHANGE": "from",
			"GONE":   "removed",
		},
		Secrets: map[string]string{
			"SECRET_CHANGE": "a-secret",
			"SECRET_GONE":   "gone-secret",
		},
	}
	b := &Environment{
		Variables: map[string]string{
			"KEEP":   "same",
			"CHANGE": "to",
			"NEW":    "added",
		},
		Secrets: map[string]string{
			"SECRET_CHANGE": "b-secret",
			"SECRET_NEW":    "new-secret",
		},
	}

	diffs := Diff(a, b)

	find := func(name string) (KeyDiff, bool) {
		for _, d := range diffs {
			if d.Name == name {
				return d, true
			}
		}
		return KeyDiff{}, false
	}

	if _, ok := find("KEEP"); ok {
		t.Fatal("unchanged key should not appear in diff")
	}

	if d, ok := find("CHANGE"); !ok {
		t.Fatal("expected CHANGE in diff")
	} else if d.Status != "changed" || d.From != "from" || d.To != "to" {
		t.Fatalf("unexpected CHANGE diff: %+v", d)
	}

	if d, ok := find("GONE"); !ok {
		t.Fatal("expected GONE in diff")
	} else if d.Status != "removed" || d.From != "removed" {
		t.Fatalf("unexpected GONE diff: %+v", d)
	}

	if d, ok := find("NEW"); !ok {
		t.Fatal("expected NEW in diff")
	} else if d.Status != "added" || d.To != "added" {
		t.Fatalf("unexpected NEW diff: %+v", d)
	}

	if d, ok := find("SECRET_CHANGE"); !ok {
		t.Fatal("expected SECRET_CHANGE in diff")
	} else if d.From != "[SECRET]" || d.To != "[SECRET]" {
		t.Fatalf("secret values must be masked: %+v", d)
	}

	if d, ok := find("SECRET_NEW"); !ok {
		t.Fatal("expected SECRET_NEW in diff")
	} else if d.To != "[SECRET]" {
		t.Fatalf("added secret must be masked: %+v", d)
	}
}

func TestDiffIdenticalEnvironmentsEmpty(t *testing.T) {
	a := &Environment{
		Variables: map[string]string{"A": "1"},
		Secrets:   map[string]string{"S": "x"},
	}
	if diffs := Diff(a, a); len(diffs) != 0 {
		t.Fatalf("identical environments should produce no diffs, got %+v", diffs)
	}
}

func TestDiffSameKeyAcrossSections(t *testing.T) {
	a := &Environment{Variables: map[string]string{"X": "v"}}
	b := &Environment{Secrets: map[string]string{"X": "s"}}

	diffs := Diff(a, b)
	if len(diffs) != 1 {
		t.Fatalf("expected one diff, got %+v", diffs)
	}
	if diffs[0].Name != "X" || diffs[0].Status != "changed" {
		t.Fatalf("unexpected diff: %+v", diffs[0])
	}
	if diffs[0].From != "[SECRET]" || diffs[0].To != "[SECRET]" {
		t.Fatalf("secret in either environment must mask both sides: %+v", diffs[0])
	}
}
