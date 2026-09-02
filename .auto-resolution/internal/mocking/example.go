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

package mocking

import (
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// maxExampleDepth guards against recursive schemas that reference themselves
// (directly or through allOf/oneOf chains) when generating example values.
const maxExampleDepth = 6

// generateExample produces a deterministic mock value for an OpenAPI schema.
// generateExample generates a deterministic example value from an OpenAPI schema.
// It prioritizes an explicit example, default, constant, or first enum value
// before generating a value from the schema structure.
func generateExample(schema *openapi3.Schema) any {
	return generateExampleDepth(schema, 0)
}

// generateExampleDepth generates a deterministic example value from an OpenAPI schema,
// honoring explicit values and limiting recursive traversal to the maximum example depth.
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

// generateObjectProperties creates deterministic example values for a schema's properties, using an additional property example when no regular properties are available. Unresolved property references receive nil.
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

// stringExample returns a deterministic example string based on the schema format.
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

// integerExample returns the schema minimum as an integer, or zero when no minimum is defined.
func integerExample(schema *openapi3.Schema) any {
	if schema.Min != nil {
		return int(*schema.Min)
	}
	return 0
}

// numberExample returns the schema's minimum value when specified, or 0.0 otherwise.
func numberExample(schema *openapi3.Schema) any {
	if schema.Min != nil {
		return *schema.Min
	}
	return 0.0
}
