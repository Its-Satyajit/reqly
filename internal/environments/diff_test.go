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
