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
