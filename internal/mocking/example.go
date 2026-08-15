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
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// maxExampleDepth guards against recursive schemas that reference themselves
// (directly or through allOf/oneOf chains) when generating example values.
const maxExampleDepth = 6

// generateExample produces a deterministic mock value for an OpenAPI schema.
// Precedence: explicit example > default > const > first enum > structural shape.
func generateExample(schema *openapi3.Schema) any {
	return generateExampleDepth(schema, 0)
}

func generateExampleDepth(schema *openapi3.Schema, depth int) any {
	if schema == nil {
		return nil
	}
	if schema.Example != nil {
		return schema.Example
	}
	if schema.Default != nil {
		return schema.Default
	}
	if schema.Const != nil {
		return schema.Const
	}
	if len(schema.Enum) > 0 {
		return schema.Enum[0]
	}
	if depth >= maxExampleDepth {
		return nil
	}

	if len(schema.AllOf) > 0 {
		merged := map[string]any{}
		for _, ref := range schema.AllOf {
			for k, v := range generateObjectProperties(ref.Value, depth) {
				merged[k] = v
			}
		}
		return merged
	}
	if len(schema.OneOf) > 0 {
		return generateExampleDepth(schema.OneOf[0].Value, depth+1)
	}
	if len(schema.AnyOf) > 0 {
		return generateExampleDepth(schema.AnyOf[0].Value, depth+1)
	}
	if schema.Type == nil {
		return nil
	}

	switch {
	case schema.Type.Is("object"):
		return generateObjectProperties(schema, depth)
	case schema.Type.Is("array"):
		if schema.Items != nil && schema.Items.Value != nil && !schema.Items.Value.Type.IsEmpty() {
			item := generateExampleDepth(schema.Items.Value, depth+1)
			return []any{item}
		}
		return []any{}
	case schema.Type.Is("string"):
		return stringExample(schema)
	case schema.Type.Is("integer"):
		return integerExample(schema)
	case schema.Type.Is("number"):
		return numberExample(schema)
	case schema.Type.Is("boolean"):
		return false
	case schema.Type.Is("null"):
		return nil
	default:
		return nil
	}
}

func generateObjectProperties(schema *openapi3.Schema, depth int) map[string]any {
	out := map[string]any{}
	keys := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		ref := schema.Properties[name]
		if ref == nil || ref.Value == nil {
			out[name] = nil
			continue
		}
		out[name] = generateExampleDepth(ref.Value, depth+1)
	}
	if additional := schema.AdditionalProperties.Schema; additional != nil && additional.Value != nil && len(out) == 0 {
		out["additionalProperty"] = generateExampleDepth(additional.Value, depth+1)
	}
	return out
}

func stringExample(schema *openapi3.Schema) any {
	switch schema.Format {
	case "date":
		return "2026-08-15"
	case "date-time":
		return "2026-08-15T12:00:00Z"
	case "uuid":
		return "00000000-0000-0000-0000-000000000000"
	case "email":
		return "user@example.com"
	case "uri", "url":
		return "https://example.com"
	default:
		return "string"
	}
}

func integerExample(schema *openapi3.Schema) any {
	if schema.Min != nil {
		return int(*schema.Min)
	}
	return 0
}

func numberExample(schema *openapi3.Schema) any {
	if schema.Min != nil {
		return *schema.Min
	}
	return 0.0
}
