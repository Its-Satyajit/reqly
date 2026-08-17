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

package auth

import (
	"testing"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestMaskValuesBearer(t *testing.T) {
	v := variables.NewSet()
	v.Set(variables.ScopeRequest, "token", "secret-tok")
	got := MaskValues("bearer", map[string]string{"token": "{{token}}"}, v)
	if len(got) != 1 || got[0] != "secret-tok" {
		t.Fatalf("got %v, want [secret-tok]", got)
	}
}

func TestMaskValuesBasicPasswordOnly(t *testing.T) {
	v := variables.NewSet()
	got := MaskValues("basic", map[string]string{"username": "alice", "password": "hunter2"}, v)
	if len(got) != 1 || got[0] != "hunter2" {
		t.Fatalf("got %v, want [hunter2] (username is not a secret)", got)
	}
}

func TestMaskValuesAPIKey(t *testing.T) {
	got := MaskValues("apikey", map[string]string{"key": "X-Key", "value": "k-999"}, variables.NewSet())
	if len(got) != 1 || got[0] != "k-999" {
		t.Fatalf("got %v, want [k-999]", got)
	}
}

func TestMaskValuesJWTSecret(t *testing.T) {
	got := MaskValues("jwt", map[string]string{"secret": "jwt-secret"}, variables.NewSet())
	if len(got) != 1 || got[0] != "jwt-secret" {
		t.Fatalf("got %v, want [jwt-secret]", got)
	}
}

func TestMaskValuesDigest(t *testing.T) {
	got := MaskValues("digest", map[string]string{"username": "u", "password": "pw"}, variables.NewSet())
	if len(got) != 1 || got[0] != "pw" {
		t.Fatalf("got %v, want [pw]", got)
	}
}

func TestMaskValuesNone(t *testing.T) {
	if got := MaskValues("none", nil, variables.NewSet()); len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestMaskValuesUnknownType(t *testing.T) {
	if got := MaskValues("ntlm", nil, variables.NewSet()); len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}
