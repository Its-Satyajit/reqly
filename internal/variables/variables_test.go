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

import "testing"

// TestResolveEnvironmentBeforeGlobal mirrors the TDD example in the testing
// strategy: environment variables must override global variables.
func TestResolveEnvironmentBeforeGlobal(t *testing.T) {
	set := NewSet()
	set.Set(ScopeGlobal, "API_URL", "https://global.example.com")
	set.Set(ScopeEnvironment, "API_URL", "https://dev.example.com")

	got, ok := set.Resolve("API_URL")
	if !ok {
		t.Fatal("expected API_URL to resolve")
	}
	if got != "https://dev.example.com" {
		t.Fatalf("expected environment value, got %q", got)
	}
}

func TestResolveRequestOverridesEnvironment(t *testing.T) {
	set := NewSet()
	set.Set(ScopeEnvironment, "TOKEN", "env-token")
	set.Set(ScopeRequest, "TOKEN", "request-token")

	got, ok := set.Resolve("TOKEN")
	if !ok {
		t.Fatal("expected TOKEN to resolve")
	}
	if got != "request-token" {
		t.Fatalf("expected request value, got %q", got)
	}
}

func TestResolveMissingVariable(t *testing.T) {
	set := NewSet()

	if _, ok := set.Resolve("MISSING"); ok {
		t.Fatal("expected MISSING not to resolve")
	}
}

func TestInterpolateReplacesMultipleVariables(t *testing.T) {
	set := NewSet()
	set.Set(ScopeEnvironment, "BASE_URL", "https://api.example.com")
	set.Set(ScopeRequest, "ID", "42")

	got, err := set.Interpolate("{{BASE_URL}}/users/{{ID}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://api.example.com/users/42"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestInterpolateUndefinedVariable(t *testing.T) {
	set := NewSet()

	if _, err := set.Interpolate("{{NOPE}}"); err == nil {
		t.Fatal("expected error for undefined variable")
	}
}

func TestInterpolateUnclosedReference(t *testing.T) {
	set := NewSet()
	set.Set(ScopeGlobal, "A", "x")

	if _, err := set.Interpolate("{{A"); err == nil {
		t.Fatal("expected error for unclosed reference")
	}
}

func TestInterpolateEmptyValue(t *testing.T) {
	set := NewSet()
	set.Set(ScopeEnvironment, "EMPTY", "")

	got, err := set.Interpolate("value={{EMPTY}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "value=" {
		t.Fatalf("expected empty interpolation, got %q", got)
	}
}
