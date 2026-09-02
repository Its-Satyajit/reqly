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

func TestMaskerAdd(t *testing.T) {
	m := NewMasker("env-secret")
	m.Add("auth-secret")

	if got := m.Mask("prefix env-secret middle auth-secret suffix"); got !=
		"prefix [SECRET] middle [SECRET] suffix" {
		t.Fatalf("got %q", got)
	}
}

func TestMaskerAddEmptyIgnored(t *testing.T) {
	m := NewMasker("keep")
	m.Add("", "")
	if got := m.Mask("keep"); got != "[SECRET]" {
		t.Fatalf("got %q", got)
	}
}

func TestMaskerAddAfterMask(t *testing.T) {
	m := NewMasker("first")
	_ = m.Mask("first")
	m.Add("second")
	if got := m.Mask("first and second"); got != "[SECRET] and [SECRET]" {
		t.Fatalf("got %q", got)
	}
}
