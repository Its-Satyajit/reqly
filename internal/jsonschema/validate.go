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

// Package jsonschema validates JSON documents against JSON Schemas, renders
// schema summaries, and generates sample instances. The same compiled schema
// feeds the CLI (reqly schema) today and the test runner's planned contract
// assertions later; nothing here sends network traffic.
package jsonschema

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Violation is one validation failure, reported at its instance path.
type Violation struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (v Violation) String() string {
	return fmt.Sprintf("%s: %s", v.Path, v.Message)
}

var draftNames = map[string]*jsonschema.Draft{
	"2020": jsonschema.Draft2020,
	"2019": jsonschema.Draft2019,
	"7":    jsonschema.Draft7,
	"6":    jsonschema.Draft6,
	"4":    jsonschema.Draft4,
}

// decode accepts JSON first, then YAML — matching requestfile semantics.
func decode(data []byte) (any, error) {
	var v any
	if err := json.Unmarshal(data, &v); err == nil {
		return v, nil
	}
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("parse document: %w", err)
	}
	return normalizeYAML(v), nil
}

// normalizeYAML converts yaml.v3 map[any]any remnants and non-string scalars
// into the JSON-shaped values the compiler expects.
func normalizeYAML(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = normalizeYAML(val)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[fmt.Sprint(k)] = normalizeYAML(val)
		}
		return out
	case []any:
		for i := range t {
			t[i] = normalizeYAML(t[i])
		}
		return t
	default:
		return v
	}
}

// Compile parses a schema document (JSON or YAML). draft optionally overrides
// $schema detection ("2020", "2019", "7", "6", "4"); an empty string defaults
// to 2020-12 when the schema declares none.
func Compile(data []byte, draft string) (*jsonschema.Schema, error) {
	doc, err := decode(data)
	if err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	d := draftNames[draft]
	if d == nil && draft != "" {
		known := make([]string, 0, len(draftNames))
		for name := range draftNames {
			known = append(known, name)
		}
		return nil, fmt.Errorf("unknown --draft %q (want one of: %s)", draft, strings.Join(known, ", "))
	}
	if d != nil {
		compiler.DefaultDraft(d)
	}
	if err := compiler.AddResource("reqly:schema", doc); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	sch, err := compiler.Compile("reqly:schema")
	if err != nil {
		return nil, fmt.Errorf("compiling schema: %w", err)
	}
	return sch, nil
}

// Validate checks an instance document (JSON or YAML) against a compiled
// schema and returns every violation flattened to instance paths.
func Validate(sch *jsonschema.Schema, instance []byte) ([]Violation, error) {
	doc, err := decode(instance)
	if err != nil {
		return nil, fmt.Errorf("instance: %w", err)
	}
	validationErr := sch.Validate(doc)
	if validationErr == nil {
		return nil, nil
	}
	ve, ok := validationErr.(*jsonschema.ValidationError)
	if !ok {
		return nil, fmt.Errorf("validating instance: %w", validationErr)
	}
	printer := message.NewPrinter(language.English)
	var violations []Violation
	flatten(ve, printer, &violations)
	return violations, nil
}

// flatten walks the ValidationError tree collecting leaf causes as violations.
func flatten(err *jsonschema.ValidationError, printer *message.Printer, out *[]Violation) {
	if len(err.Causes) > 0 {
		for _, cause := range err.Causes {
			flatten(cause, printer, out)
		}
		return
	}
	*out = append(*out, Violation{
		Path:    joinInstancePath(err.InstanceLocation),
		Message: err.ErrorKind.LocalizedString(printer),
	})
}

// joinInstancePath renders []string{"items","2","price"} as $.items[2].price.
func joinInstancePath(loc []string) string {
	sb := strings.Builder{}
	sb.WriteString("$")
	for _, seg := range loc {
		if isNumeric(seg) {
			sb.WriteByte('[')
			sb.WriteString(seg)
			sb.WriteByte(']')
			continue
		}
		sb.WriteByte('.')
		sb.WriteString(seg)
	}
	return sb.String()
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
