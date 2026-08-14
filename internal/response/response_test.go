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

package response

import "testing"

func TestResponseOK(t *testing.T) {
	for _, tc := range []struct {
		code int
		ok   bool
	}{
		{199, false},
		{200, true},
		{204, true},
		{299, true},
		{300, false},
		{404, false},
		{500, false},
	} {
		if got := (&Response{StatusCode: tc.code}).OK(); got != tc.ok {
			t.Fatalf("expected OK()=%v for %d, got %v", tc.ok, tc.code, got)
		}
	}
}

func TestResponseText(t *testing.T) {
	resp := &Response{Body: []byte("hello world")}
	if got := resp.Text(); got != "hello world" {
		t.Fatalf("expected 'hello world', got %q", got)
	}
}

func TestResponseEmptyText(t *testing.T) {
	resp := &Response{}
	if got := resp.Text(); got != "" {
		t.Fatalf("expected empty text, got %q", got)
	}
}
