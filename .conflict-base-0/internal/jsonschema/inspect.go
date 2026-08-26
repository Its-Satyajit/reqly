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
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const maxInspectDepth = 12

// Inspect renders a schema document as a text tree: one line per node with
// its type, required marker, and inline constraint summary. $refs resolve to
// their target with the pointer shown; cycles stop expansion.
func Inspect(data []byte) (string, error) {
	doc, err := decode(data)
	if err != nil {
		return "", fmt.Errorf("schema: %w", err)
	}
	var sb strings.Builder
	walkInspect(doc, doc, "", 0, map[string]bool{}, &sb)
	return sb.String(), nil
}

func walkInspect(root, node any, required string, depth int, refChain map[string]bool, sb *strings.Builder) {
	if depth > maxInspectDepth {
		sb.WriteString("  \n")
		return
	}
	obj, ok := node.(map[string]any)
	if !ok {
		return
	}

	target := ""
	if ref, ok := obj["$ref"].(string); ok && strings.HasPrefix(ref, "#/") {
		resolved := resolvePointer(root, ref)
		target = strings.TrimPrefix(ref, "#/")
		if resolved == nil || refChain[ref] {
			sb.WriteString(fmt.Sprintf("%s%s → %s (recursive)\n", indent(depth), required, target))
			return
		}
		refChain[ref] = true
		defer delete(refChain, ref)
		sb.WriteString(fmt.Sprintf("%s%s → %s\n", indent(depth), required, target))
		walkInspect(root, resolved, "", depth+1, refChain, sb)
		return
	}

	sb.WriteString(indent(depth) + describeType(obj) + required + describeConstraints(obj) + "\n")

	props, _ := obj["properties"].(map[string]any)
	reqSet := map[string]bool{}
	if reqs, ok := obj["required"].([]any); ok {
		for _, r := range reqs {
			if s, ok := r.(string); ok {
				reqSet[s] = true
			}
		}
	}
	names := sortedKeys(props)
	for _, name := range names {
		marker := ""
		if reqSet[name] {
			marker = "! (required)"
		} else {
			marker = ""
		}
		child := props[name]
		sb.WriteString(indent(depth+1) + name)
		if childObj, ok := child.(map[string]any); ok && isRefOrNested(childObj) {
			sb.WriteString(":\n")
			walkInspect(root, child, marker, depth+2, refChain, sb)
		} else {
			walkLeaf(child, marker, sb)
			sb.WriteString("\n")
		}
	}
	if items, ok := obj["items"]; ok {
		sb.WriteString(indent(depth+1) + "items:\n")
		walkInspect(root, items, "", depth+2, refChain, sb)
	}
}

// walkLeaf renders a scalar/inline property on a single line.
func walkLeaf(node any, suffix string, sb *strings.Builder) {
	obj, ok := node.(map[string]any)
	if !ok {
		sb.WriteString(suffix)
		return
	}
	if ref, ok := obj["$ref"].(string); ok {
		fmt.Fprintf(sb, " → %s%s", strings.TrimPrefix(ref, "#/"), suffix)
		return
	}
	sb.WriteString(" " + describeType(obj) + suffix + describeConstraints(obj))
}

func isRefOrNested(obj map[string]any) bool {
	if _, ok := obj["$ref"]; ok {
		return false // refs render inline with target pointer
	}
	if t, _ := obj["type"].(string); t == "object" || t == "array" {
		return true
	}
	if _, ok := obj["properties"]; ok {
		return true
	}
	if _, ok := obj["allOf"]; ok {
		return true
	}
	if _, ok := obj["oneOf"]; ok {
		return true
	}
	if _, ok := obj["anyOf"]; ok {
		return true
	}
	return false
}

func describeType(obj map[string]any) string {
	switch t := obj["type"].(type) {
	case string:
		return t
	case []any:
		parts := make([]string, 0, len(t))
		for _, v := range t {
			parts = append(parts, fmt.Sprint(v))
		}
		return strings.Join(parts, "|")
	}
	if _, ok := obj["properties"]; ok {
		return "object"
	}
	if _, ok := obj["items"]; ok {
		return "array"
	}
	if _, ok := obj["enum"]; ok {
		return ""
	}
	return "any"
}

func describeConstraints(obj map[string]any) string {
	var parts []string
	if c, ok := obj["const"]; ok {
		parts = append(parts, "const:"+fmt.Sprint(c))
	}
	if e, ok := obj["enum"].([]any); ok && len(e) > 0 {
		vals := make([]string, 0, len(e))
		for _, v := range e {
			vals = append(vals, fmt.Sprint(v))
		}
		parts = append(parts, "enum:"+strings.Join(vals, "|"))
	}
	if f, ok := obj["format"].(string); ok {
		parts = append(parts, "format:"+f)
	}
	if p, ok := obj["pattern"].(string); ok {
		parts = append(parts, "pattern:"+p)
	}
	if v, ok := numberVal(obj["minimum"]); ok {
		parts = append(parts, "min:"+strconv.FormatFloat(v, 'f', -1, 64))
	}
	if v, ok := numberVal(obj["maximum"]); ok {
		parts = append(parts, "max:"+strconv.FormatFloat(v, 'f', -1, 64))
	}
	if v, ok := numberVal(obj["minLength"]); ok {
		parts = append(parts, "minLength:"+strconv.FormatFloat(v, 'f', -1, 64))
	}
	if v, ok := numberVal(obj["maxLength"]); ok && v != 0 {
		parts = append(parts, "maxLength:"+strconv.FormatFloat(v, 'f', -1, 64))
	}
	if v, ok := numberVal(obj["minItems"]); ok {
		parts = append(parts, "minItems:"+strconv.FormatFloat(v, 'f', -1, 64))
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

func numberVal(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func indent(depth int) string {
	if depth <= 0 {
		return ""
	}
	return strings.Repeat("  ", depth)
}
