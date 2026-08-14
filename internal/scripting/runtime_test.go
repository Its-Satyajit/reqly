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

package scripting

import "testing"

func TestRuntimeIsLazy(t *testing.T) {
	r := NewRuntime()
	if r.vm != nil {
		t.Fatal("expected Goja runtime to remain uninitialized until first use")
	}
}

func TestRunScriptReturnsValue(t *testing.T) {
	r := NewRuntime()
	value, err := r.RunScript("1 + 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := value.ToInteger(); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestRunScriptReportsSyntaxError(t *testing.T) {
	r := NewRuntime()
	if _, err := r.RunScript("const =="); err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestRuntimeRetainsStateAcrossRuns(t *testing.T) {
	r := NewRuntime()
	if _, err := r.RunScript("var counter = 0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.RunScript("counter = counter + 1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	value, err := r.RunScript("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := value.ToInteger(); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}
