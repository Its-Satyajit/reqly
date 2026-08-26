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
	"fmt"
	"math/rand"
	"strings"
)

const maxGenerateDepth = 8

// GenerateOptions tunes instance synthesis.
type GenerateOptions struct {
	// Seed varies randomized choices; zero means fully deterministic defaults.
	Seed int64
	// IncludeOptional adds non-required object properties deterministically.
	IncludeOptional bool
}

// Generate synthesizes a sample instance for a schema document (JSON or YAML)
// and returns it as pretty-printed JSON. Explicit values win over synthesized
// ones; unresolvable constraints degrade to warnings, never errors.
func Generate(schemaData []byte, opts GenerateOptions) ([]byte, []string, error) {
	rootDoc, err := decode(schemaData)
	if err != nil {
		return nil, nil, fmt.Errorf("schema: %w", err)
	}
	if _, ok := rootDoc.(map[string]any); !ok {
		return nil, nil, fmt.Errorf("schema: not a JSON Schema object")
	}
	g := &generator{
		root: rootDoc,
		opts: opts,
	}
	if opts.Seed != 0 {
		g.rng = rand.New(rand.NewSource(opts.Seed))
	}
	value := g.synth(rootAsSchema(rootDoc), "$", map[string]bool{}, 0)
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, g.warnings, fmt.Errorf("serialize instance: %w", err)
	}
	return out, g.warnings, nil
}

type generator struct {
	root     any
	opts     GenerateOptions
	warnings []string
	rng      *rand.Rand
}

func (g *generator) warnf(format string, args ...any) {
	g.warnings = append(g.warnings, fmt.Sprintf(format, args...))
}

func rootAsSchema(doc any) any {
	if m, ok := doc.(map[string]any); ok {
		if _, hasRef := m["$ref"]; !hasRef {
			return m
		}
		return m
	}
	return map[string]any{}
}

// synth renders one node of the schema into a concrete value.
func (g *generator) synth(node any, path string, refChain map[string]bool, depth int) any {
	if depth > maxGenerateDepth {
		g.warnf("%s: exceeded maximum recursion depth (%d); emitted null", path, maxGenerateDepth)
		return nil
	}

	obj, ok := node.(map[string]any)
	if !ok {
		return nil
	}

	if ref, isRef := obj["$ref"].(string); isRef && strings.HasPrefix(ref, "#/") {
		if refChain[ref] {
			g.warnf("%s: recursive $ref %s hit depth cap; emitted null", path, strings.TrimPrefix(ref, "#/"))
			return nil
		}
		resolved := resolvePointer(g.root, ref)
		if resolved == nil {
			g.warnf("%s: unresolvable $ref %s; emitted null", path, ref)
			return nil
		}
		next := make(map[string]bool, len(refChain)+1)
		for k := range refChain {
			next[k] = true
		}
		next[ref] = true
		return g.synth(resolved, path+"→"+ref, next, depth+1)
	}

	if merged, changed := mergeAllOf(obj); changed {
		obj = merged
	}
	if branch := firstCompositeBranch(obj); branch != nil {
		switch branch.keyword {
		case "oneOf", "anyOf":
			options, _ := branch.value.([]any)
			if len(options) == 0 {
				g.warnf("%s: empty %s; emitted null", path, branch.keyword)
				return nil
			}
			return g.synth(options[0], path, refChain, depth+1)
		case "not":
			g.warnf("%s: \"not\" ignored; emitted null", path)
			return nil
		}
	}

	if c, ok := obj["const"]; ok {
		return normalizeScalar(c)
	}
	if e, ok := obj["enum"].([]any); ok && len(e) > 0 {
		return normalizeScalar(e[0])
	}
	if ex, ok := obj["example"]; ok {
		return g.expandExample(ex)
	}
	if d, ok := obj["default"]; ok {
		return g.expandExample(d)
	}

	schemaType := typeOf(obj)
	switch schemaType {
	case "object":
		return g.synthObject(obj, path, refChain, depth)
	case "array":
		return g.synthArray(obj, path, refChain, depth)
	case "string":
		return g.synthString(obj, path)
	case "integer":
		return int64(g.synthNumber(obj, true))
	case "number":
		return g.synthNumber(obj, false)
	case "boolean":
		return true
	case "null":
		return nil
	default:
		return ""
	}
}

// expandExample deep-copies an example/default and synthesizes any missing
// required structure inside objects.
func (g *generator) expandExample(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = g.expandExample(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = g.expandExample(val)
		}
		return out
	default:
		return normalizeScalar(v)
	}
}

func (g *generator) synthObject(obj map[string]any, path string, refChain map[string]bool, depth int) any {
	props, _ := obj["properties"].(map[string]any)
	reqSet := map[string]bool{}
	if reqs, ok := obj["required"].([]any); ok {
		for _, r := range reqs {
			if s, ok := r.(string); ok {
				reqSet[s] = true
			}
		}
	}
	out := map[string]any{}
	names := sortedKeys(props)
	for _, name := range names {
		required := reqSet[name]
		if !required && !g.opts.IncludeOptional {
			continue
		}
		childPath := path + "." + name
		out[name] = g.synth(props[name], childPath, refChain, depth+1)
	}
	return out
}

func (g *generator) synthArray(obj map[string]any, path string, refChain map[string]bool, depth int) any {
	minItems := numOr(obj["minItems"], 1)
	count := int(minItems)
	if count < 1 {
		count = 1
	}
	items := obj["items"]
	out := make([]any, 0, count)
	for i := 0; i < count; i++ {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		out = append(out, g.synth(items, itemPath, refChain, depth+1))
	}
	return out
}

var stringFillers = []string{"string", "alpha", "bravo", "charlie", "delta"}

func (g *generator) synthString(obj map[string]any, path string) string {
	var s string
	format, _ := obj["format"].(string)
	s = formatValue(format, g.rng)

	minLen := int(numOr(obj["minLength"], 0))
	maxLen := int(numOr(obj["maxLength"], 0))
	filler := g.pickFiller()
	if s == "" {
		s = filler
	} else if minLen > 0 && len(s) < minLen {
		s = pad(s, filler, minLen)
	}
	for len(s) < minLen {
		s += filler
	}
	if maxLen > 0 && len(s) > maxLen {
		s = s[:maxLen]
	}
	if _, hasPattern := obj["pattern"]; hasPattern && format == "" {
		if _, explicit := obj["example"]; !explicit {
			if _, explicit := obj["default"]; !explicit {
				g.warnf("%s: pattern cannot be synthesized; used %q", path, s)
			}
		}
	}
	return s
}

func (g *generator) pickFiller() string {
	if g.rng != nil {
		return stringFillers[g.rng.Intn(len(stringFillers))]
	}
	return stringFillers[0]
}

func (g *generator) synthNumber(obj map[string]any, integer bool) float64 {
	n := 1.0
	if min, ok := numberVal(obj["minimum"]); ok {
		n = min
	} else if minEx, ok := numberVal(obj["exclusiveMinimum"]); ok {
		n = minEx + 1
	} else if max, ok := numberVal(obj["maximum"]); ok {
		n = max - 1
	}
	if mult, ok := numberVal(obj["multipleOf"]); ok && mult > 0 {
		steps := n / mult
		ceil := float64(int(steps))
		if ceil < steps {
			ceil++
		}
		if ceil < 1 {
			ceil = 1
		}
		n = ceil * mult
	}
	if integer {
		n = float64(int64(n))
	}
	return n
}

// composite groups a combinator keyword with its value.
type composite struct {
	keyword string
	value   any
}

// firstCompositeBranch finds allOf/oneOf/anyOf/not on a node.
func firstCompositeBranch(obj map[string]any) *composite {
	for _, kw := range []string{"oneOf", "anyOf", "not", "allOf"} {
		if v, ok := obj[kw]; ok {
			return &composite{keyword: kw, value: v}
		}
	}
	return nil
}

// mergeAllOf shallow-merges allOf branches: first non-empty type wins,
// properties and required lists concatenate.
func mergeAllOf(obj map[string]any) (map[string]any, bool) {
	branches, ok := obj["allOf"].([]any)
	if !ok || len(branches) == 0 {
		return obj, false
	}
	merged := map[string]any{}
	for k, v := range obj {
		if k != "allOf" {
			merged[k] = v
		}
	}
	for _, b := range branches {
		bm, ok := b.(map[string]any)
		if !ok {
			continue
		}
		if _, exists := merged["type"]; !exists {
			if t, ok := bm["type"]; ok {
				merged["type"] = t
			}
		}
		if props, ok := bm["properties"].(map[string]any); ok {
			existing, _ := merged["properties"].(map[string]any)
			if existing == nil {
				existing = map[string]any{}
			}
			for pk, pv := range props {
				existing[pk] = pv
			}
			merged["properties"] = existing
		}
		if reqs, ok := bm["required"].([]any); ok {
			existing, _ := merged["required"].([]any)
			merged["required"] = append(existing, reqs...)
		}
	}
	return merged, true
}

func typeOf(obj map[string]any) string {
	switch t := obj["type"].(type) {
	case string:
		return t
	case []any:
		if len(t) > 0 {
			if s, ok := t[0].(string); ok {
				return s
			}
		}
	}
	if _, ok := obj["properties"]; ok {
		return "object"
	}
	if _, ok := obj["items"]; ok {
		return "array"
	}
	return ""
}

func numOr(v any, fallback float64) float64 {
	if n, ok := numberVal(v); ok {
		return n
	}
	return fallback
}

// normalizeScalar converts decoded numbers to JSON-friendly values.
func normalizeScalar(v any) any {
	if f, ok := v.(float64); ok {
		if f == float64(int64(f)) {
			return int64(f)
		}
	}
	return v
}

// resolvePointer walks a "#/a/b" JSON pointer within the root document.
func resolvePointer(root any, ref string) any {
	current := root
	for _, seg := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		if seg == "" {
			continue
		}
		seg = strings.ReplaceAll(seg, "~1", "/")
		seg = strings.ReplaceAll(seg, "~0", "~")
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		next, ok := m[seg]
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

// formatValue renders a realistic value for known formats; empty for unknown.
func formatValue(format string, rng *rand.Rand) string {
	switch format {
	case "email":
		return "user@example.com"
	case "date-time":
		return "2026-01-01T00:00:00Z"
	case "date":
		return "2026-01-01"
	case "time":
		return "12:00:00Z"
	case "uri", "uri-reference":
		return "https://example.com/resource"
	case "uuid":
		if rng != nil {
			b := make([]byte, 16)
			rng.Read(b)
			return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
		}
		return "00000000-0000-4000-8000-000000000000"
	case "ipv4":
		return "192.0.2.1"
	case "ipv6":
		return "2001:db8::1"
	case "hostname":
		return "host.example.com"
	default:
		return ""
	}
}

// pad extends s by repeating filler until it reaches at least min characters.
func pad(s, filler string, min int) string {
	for len(s) < min {
		s += filler
	}
	return s
}
