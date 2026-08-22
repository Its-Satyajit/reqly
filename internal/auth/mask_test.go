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
