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
