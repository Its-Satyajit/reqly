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
	"strings"
	"testing"
)

const userSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["id", "email"],
  "properties": {
    "id": {"type": "integer"},
    "email": {"type": "string", "format": "email"},
    "tags": {"type": "array", "items": {"type": "string"}}
  }
}`

func TestCompileAndValidInstance(t *testing.T) {
	sch, err := Compile([]byte(userSchema), "")
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	violations, err := Validate(sch, []byte(`{"id": 1, "email": "a@b.co", "tags": ["x"]}`))
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %v", violations)
	}
}

func TestValidateViolationsWithPaths(t *testing.T) {
	sch, err := Compile([]byte(userSchema), "")
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	violations, err := Validate(sch, []byte(`{"id": "nope", "tags": ["ok", 3]}`))
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(violations) < 3 {
		t.Fatalf("expected at least 3 violations (bad id, missing email, bad tag), got %d: %v", len(violations), violations)
	}
	paths := map[string]bool{}
	missingMentioned := false
	for _, v := range violations {
		paths[v.Path] = true
		if strings.Contains(v.Message, "email") {
			missingMentioned = true
		}
	}
	for _, want := range []string{"$.id", "$.tags[1]"} {
		if !paths[want] {
			t.Errorf("missing violation path %q; got %v", want, paths)
		}
	}
	if !missingMentioned {
		t.Errorf("no violation mentions the missing email property: %v", violations)
	}
	for _, v := range violations {
		if v.Message == "" {
			t.Errorf("violation at %q has empty message", v.Path)
		}
	}
}

func TestCompileYAMLSchema(t *testing.T) {
	yamlSchema := `
$schema: "https://json-schema.org/draft/2020-12/schema"
type: object
required: [name]
properties:
  name:
    type: string
`
	sch, err := Compile([]byte(yamlSchema), "")
	if err != nil {
		t.Fatalf("Compile() YAML error = %v", err)
	}
	violations, err := Validate(sch, []byte(`{"name": "Ada"}`))
	if err != nil || len(violations) != 0 {
		t.Fatalf("YAML round-trip failed: %v %v", violations, err)
	}
}

func TestCompileInvalidSchemaErrors(t *testing.T) {
	if _, err := Compile([]byte(`{"type": "not-a-real-type"}`), ""); err == nil {
		t.Fatal("expected compile error for invalid schema")
	}
	if _, err := Compile([]byte(`{broken`), ""); err == nil {
		t.Fatal("expected parse error for malformed schema")
	}
}

func TestValidateUnparseableInstance(t *testing.T) {
	sch, err := Compile([]byte(userSchema), "")
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if _, err := Validate(sch, []byte(`{broken`)); err == nil {
		t.Fatal("expected parse error for malformed instance")
	}
}

func TestDraftOverride(t *testing.T) {
	draft7 := `{"type": "object", "required": ["x"], "properties": {"x": {"type": "string"}}}`
	sch, err := Compile([]byte(draft7), "7")
	if err != nil {
		t.Fatalf("Compile() with --draft 7 error = %v", err)
	}
	violations, err := Validate(sch, []byte(`{}`))
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0].Message+violations[0].Path, "x") {
		t.Fatalf("expected one violation about required x, got %v", violations)
	}
}

func TestDefaultDraftWithoutSchemaKeyword(t *testing.T) {
	noDraft := `{"type": "object"}`
	if _, err := Compile([]byte(noDraft), ""); err != nil {
		t.Fatalf("Compile() without $schema should default to 2020-12, got %v", err)
	}
}
