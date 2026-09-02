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

import (
	"reflect"
	"testing"
)

func jsonResp(body string) *Response {
	return &Response{Body: []byte(body)}
}

func TestJSONValue(t *testing.T) {
	resp := jsonResp(`{
		"user": {"name": "reqly", "age": 30, "active": true},
		"tags": ["a", "b", "c"],
		"meta": {"nested": {"deep": 1.5}}
	}`)

	cases := []struct {
		path string
		want any
		ok   bool
	}{
		{"$.user.name", "reqly", true},
		{"$.user.age", float64(30), true},
		{"$.user.active", true, true},
		{"$.tags[0]", "a", true},
		{"$.tags[2]", "c", true},
		{"$.meta.nested.deep", 1.5, true},
		{"$['user']['name']", "reqly", true},
		{"user.name", "reqly", true},
		{".user.name", "reqly", true},
		{"$.missing", nil, false},
		{"$.user.missing", nil, false},
		{"$.tags[5]", nil, false},
		{"$.tags[abc]", nil, false},
		{"$.tags[-1]", nil, false},
		{"$.user.name.extra", nil, false},
	}

	for _, tc := range cases {
		got := resp.JSONValue(tc.path)
		if tc.ok {
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("JSONValue(%q) = %#v, want %#v", tc.path, got, tc.want)
			}
		} else if got != nil {
			t.Errorf("JSONValue(%q) = %#v, want nil", tc.path, got)
		}
	}
}

func TestJSONValueNonJSONBody(t *testing.T) {
	resp := jsonResp("not json")
	if got := resp.JSONValue("$.anything"); got != nil {
		t.Errorf("expected nil for non-JSON body, got %#v", got)
	}
	if got := resp.JSON(); got != nil {
		t.Errorf("expected nil JSON() for non-JSON body, got %#v", got)
	}
}

func TestJSONValueEmptyBody(t *testing.T) {
	resp := jsonResp("")
	if got := resp.JSONValue("$.x"); got != nil {
		t.Errorf("expected nil for empty body, got %#v", got)
	}
}

func TestJSONObject(t *testing.T) {
	resp := jsonResp(`{"a":1}`)
	got := resp.JSON()
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", got)
	}
	if m["a"] != float64(1) {
		t.Fatalf("expected a=1, got %v", m["a"])
	}
}

func TestParseJSONPath(t *testing.T) {
	cases := []struct {
		path string
		want []string
	}{
		{"$.user.name", []string{"user", "name"}},
		{"$['users'][0]['name']", []string{"users", "0", "name"}},
		{"$.a.b[2].c", []string{"a", "b", "2", "c"}},
		{"user", []string{"user"}},
		{"", []string{}},
		{"$.", []string{}},
		{"$[0]", []string{"0"}},
	}

	for _, tc := range cases {
		steps := parseJSONPath(tc.path)
		keys := make([]string, len(steps))
		for i, s := range steps {
			keys[i] = s.key
		}
		if !reflect.DeepEqual(keys, tc.want) {
			t.Errorf("parseJSONPath(%q) = %v, want %v", tc.path, keys, tc.want)
		}
	}
}
