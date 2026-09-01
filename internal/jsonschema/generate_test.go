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

package jsonschema

import (
	"encoding/json"
	"strings"
	"testing"
)

const inspectSchema = `
$schema: "https://json-schema.org/draft/2020-12/schema"
type: object
required: [id]
properties:
  id:
    type: integer
  role:
    type: string
    enum: [admin, user]
    maxLength: 20
  profile:
    $ref: "#/$defs/profile"
$defs:
  profile:
    type: object
    properties:
      bio:
        type: string
`

func TestInspectTree(t *testing.T) {
	out, err := Inspect([]byte(inspectSchema))
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	for _, want := range []string{"object", "id", "required", "integer", "role", "enum", "maxLength"} {
		if !strings.Contains(out, want) {
			t.Errorf("inspect output missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "$defs/profile") {
		t.Errorf("resolved $ref target not shown:\n%s", out)
	}
}

func TestInspectCycleSafe(t *testing.T) {
	cycle := `{
	  "$defs": {"node": {"type": "object", "properties": {"child": {"$ref": "#/$defs/node"}}}},
	  "$ref": "#/$defs/node"
	}`
	out, err := Inspect([]byte(cycle))
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if strings.Count(out, "child") > 2 {
		t.Errorf("cycle expanded too deep:\n%s", out)
	}
}

const genSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["name", "age", "kind", "email", "refused", "self", "code"],
  "properties": {
    "name": {"type": "string", "example": "Ada"},
    "age": {"type": "integer", "minimum": 18, "maximum": 99},
    "kind": {"enum": ["a", "b"]},
    "fixed": {"const": 42},
    "email": {"type": "string", "format": "email"},
    "nickname": {"type": "string"},
    "pets": {"type": "array", "items": {"type": "string"}, "minItems": 2},
    "code": {"type": "string", "pattern": "^XX-\\d+$"},
    "profile": {"allOf": [{"type": "object"}, {"properties": {"bio": {"type": "string"}, "level": {"type": "integer"}}}]},
    "either": {"anyOf": [{"type": "string"}, {"type": "integer"}]},
    "refused": {"not": {}},
    "self": {"$ref": "#/$defs/node"}
  },
  "$defs": {
    "node": {"type": "object", "required": ["child"], "properties": {"child": {"$ref": "#/$defs/node"}}}
  }
}`

func decodeGen(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestGenerateDeterministicAndPrecedence(t *testing.T) {
	a, _, err := Generate([]byte(genSchema), GenerateOptions{IncludeOptional: true})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	b, _, err := Generate([]byte(genSchema), GenerateOptions{IncludeOptional: true})
	if err != nil {
		t.Fatalf("Generate() second run error = %v", err)
	}
	if string(a) != string(b) {
		t.Errorf("generation is not deterministic:\n%s\n---\n%s", a, b)
	}
	inst := decodeGen(t, a)
	if inst["fixed"] != float64(42) {
		t.Errorf("const precedence broken: fixed = %v", inst["fixed"])
	}
	if inst["name"] != "Ada" {
		t.Errorf("example precedence broken: name = %v", inst["name"])
	}
	if inst["kind"] != "a" {
		t.Errorf("enum precedence broken: kind = %v", inst["kind"])
	}
	if n, ok := inst["age"].(float64); !ok || n < 18 || n > 99 {
		t.Errorf("minimum/maximum violated: age = %v", inst["age"])
	}
	if !strings.Contains(inst["email"].(string), "@") {
		t.Errorf("format email not honored: %v", inst["email"])
	}
	pets := inst["pets"].([]any)
	if len(pets) < 2 {
		t.Errorf("minItems violated: pets = %v", pets)
	}
	if inst["code"] != "string" {
		t.Errorf("pattern fallback = %v, want \"string\"", inst["code"])
	}
	profile := inst["profile"].(map[string]any)
	if _, ok := profile["bio"]; !ok {
		t.Error("allOf properties not merged")
	}
	if _, ok := inst["either"].(string); !ok {
		t.Errorf("anyOf first branch = %v, want string", inst["either"])
	}
}

func TestGenerateWarnings(t *testing.T) {
	_, warnings, err := Generate([]byte(genSchema), GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	joined := strings.Join(warnings, "\n")
	for _, want := range []string{"pattern", "not", "depth"} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings missing %q:\n%s", want, joined)
		}
	}
}

func TestGenerateOptionalFlag(t *testing.T) {
	data, _, err := Generate([]byte(genSchema), GenerateOptions{IncludeOptional: true})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	inst := decodeGen(t, data)
	if nick, ok := inst["nickname"].(string); !ok || nick == "" {
		t.Errorf("optional nickname = %v, want synthesized string", inst["nickname"])
	}
}

func TestGenerateStringConstraintsAndSeed(t *testing.T) {
	schema := `{
      "type": "object",
      "properties": {
        "s": {"type": "string", "minLength": 12, "maxLength": 20},
        "n": {"type": "number", "multipleOf": 5}
      },
      "required": ["s", "n"]
    }`
	data, _, err := Generate([]byte(schema), GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	inst := decodeGen(t, data)
	s := inst["s"].(string)
	if len(s) < 12 || len(s) > 20 {
		t.Errorf("length constraints violated: s = %q (%d)", s, len(s))
	}
	n := inst["n"].(float64)
	if int64(n)%5 != 0 {
		t.Errorf("multipleOf violated: n = %v", n)
	}

	seeded, _, err := Generate([]byte(schema), GenerateOptions{Seed: 7})
	if err != nil {
		t.Fatalf("Generate(seed) error = %v", err)
	}
	if len(seeded) == 0 {
		t.Fatal("seeded generation produced nothing")
	}
}

func TestGenerateKnownFormats(t *testing.T) {
	schema := `{
      "type": "object",
      "required": ["dt", "u", "ip", "host", "link"],
      "properties": {
        "dt": {"type": "string", "format": "date-time"},
        "u": {"type": "string", "format": "uuid"},
        "ip": {"type": "string", "format": "ipv4"},
        "host": {"type": "string", "format": "hostname"},
        "link": {"type": "string", "format": "uri"}
      }
    }`
	data, _, err := Generate([]byte(schema), GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	inst := decodeGen(t, data)
	checks := map[string]func(string) bool{
		"dt":   func(s string) bool { return strings.Contains(s, "T") && strings.Contains(s, "Z") },
		"u":    func(s string) bool { return len(strings.Split(s, "-")) == 5 },
		"ip":   func(s string) bool { return strings.Count(s, ".") == 3 },
		"host": func(s string) bool { return strings.Contains(s, ".") },
		"link": func(s string) bool { return strings.HasPrefix(s, "https://") },
	}
	for key, ok := range checks {
		got := inst[key].(string)
		if !ok(got) {
			t.Errorf("format %s generated %q, which fails its check", key, got)
		}
	}
}

func TestGenerateInvalidSchemaErrors(t *testing.T) {
	if _, _, err := Generate([]byte(`nope`), GenerateOptions{}); err == nil {
		t.Fatal("expected error for unparseable schema")
	}
}

func TestGenerateYAMLSchema(t *testing.T) {
	yaml := `
type: object
required: [x]
properties:
  x:
    type: string
`
	data, _, err := Generate([]byte(yaml), GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() YAML error = %v", err)
	}
	if !strings.Contains(string(data), `"x"`) {
		t.Errorf("unexpected YAML-generated instance: %s", data)
	}
}
