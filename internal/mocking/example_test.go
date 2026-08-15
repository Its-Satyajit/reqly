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

package mocking

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func schemaFixture(yamlSpec string) *openapi3.Schema {
	var raw any
	if err := yaml.Unmarshal([]byte(yamlSpec), &raw); err != nil {
		panic(err)
	}
	data, err := json.Marshal(raw)
	if err != nil {
		panic(err)
	}
	schema := &openapi3.Schema{}
	if err := json.Unmarshal(data, schema); err != nil {
		panic(err)
	}
	return schema
}

func TestGenerateExamplePrimitives(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want any
	}{
		{"string", `type: string`, "string"},
		{"date", `type: string
format: date`, "2026-08-15"},
		{"date-time", `type: string
format: date-time`, "2026-08-15T12:00:00Z"},
		{"uuid", `type: string
format: uuid`, "00000000-0000-0000-0000-000000000000"},
		{"integer", `type: integer`, 0},
		{"integer-with-min", `type: integer
minimum: 1`, 1},
		{"number", `type: number`, 0.0},
		{"boolean", `type: boolean`, false},
		{"null", `type: "null"`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateExample(schemaFixture(tt.yaml))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("generateExample() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGenerateExampleExplicit(t *testing.T) {
	schema := schemaFixture(`type: string
example: hello`)
	if got := generateExample(schema); got != "hello" {
		t.Fatalf("example precedence: got %#v, want hello", got)
	}

	schema = schemaFixture(`type: integer
default: 42`)
	if got := generateExample(schema); got != float64(42) && got != 42 {
		t.Fatalf("default precedence: got %#v, want 42", got)
	}

	schema = schemaFixture(`type: string
enum: [red, green]`)
	if got := generateExample(schema); got != "red" {
		t.Fatalf("enum precedence: got %#v, want red", got)
	}
}

func TestGenerateExampleObject(t *testing.T) {
	schema := schemaFixture(`type: object
properties:
  id:
    type: integer
  name:
    type: string
`)
	got := generateExample(schema).(map[string]any)
	if got["id"] != 0 {
		t.Fatalf("id = %#v, want 0", got["id"])
	}
	if got["name"] != "string" {
		t.Fatalf("name = %#v, want string", got["name"])
	}
}

func TestGenerateExampleArray(t *testing.T) {
	schema := schemaFixture(`type: array
items:
  type: integer
`)
	got := generateExample(schema).([]any)
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("array = %#v, want [0]", got)
	}
}

func TestGenerateExampleAllOf(t *testing.T) {
	schema := schemaFixture(`allOf:
  - type: object
    properties:
      a:
        type: integer
  - type: object
    properties:
      b:
        type: string
`)
	got := generateExample(schema).(map[string]any)
	if got["a"] != 0 || got["b"] != "string" {
		t.Fatalf("allOf merge = %#v", got)
	}
}

func TestGenerateExampleOneOf(t *testing.T) {
	schema := schemaFixture(`oneOf:
  - type: object
    properties:
      kind:
        type: string
        example: cat
  - type: object
    properties:
      kind:
        type: string
        example: dog
`)
	got := generateExample(schema).(map[string]any)
	if got["kind"] != "cat" {
		t.Fatalf("oneOf first = %#v, want kind=cat", got)
	}
}

func TestGenerateExampleRecursiveTerminates(t *testing.T) {
	node := &openapi3.Schema{Type: &openapi3.Types{"object"}}
	node.Properties = openapi3.Schemas{"child": {Value: node}}
	got := generateExample(node)
	// Must not stack overflow; yields something within depth bounds.
	if got == nil {
		t.Fatal("expected a non-nil value for recursive schema")
	}
}
