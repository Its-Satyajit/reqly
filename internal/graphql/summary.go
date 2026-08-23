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

package graphql

import (
	"sort"
	"strings"
)

func sortTypes(types []*Type) {
	sort.Slice(types, func(i, j int) bool { return types[i].Name < types[j].Name })
}

// Signature renders a field GraphQL-style: field(arg: T, arg2: U): Return.
func (f Field) Signature() string {
	var sb strings.Builder
	sb.WriteString(f.Name)
	if len(f.Args) > 0 {
		sb.WriteString("(")
		for i, a := range f.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(a.Name)
			sb.WriteString(": ")
			sb.WriteString(a.Type.String())
			if a.Def != "" {
				sb.WriteString(" = ")
				sb.WriteString(a.Def)
			}
		}
		sb.WriteString(")")
	}
	sb.WriteString(": ")
	sb.WriteString(f.Type.String())
	return sb.String()
}

// Summary renders the schema as text: root fields first, then remaining
// object/interface/input types alphabetically; enums show values inline.
// When typeName is non-empty, only that type is rendered.
func (s *Schema) Summary(typeName string) string {
	var sb strings.Builder
	renderType := func(t *Type, header string) {
		sb.WriteString(header)
		sb.WriteString("\n")
		for _, f := range t.Fields {
			sb.WriteString("  ")
			sb.WriteString(f.Signature())
			sb.WriteString("\n")
		}
	}
	roots := [][2]string{
		{"query", "query"},
		{"mutation", "mutation"},
		{"subscription", "subscription"},
	}
	if typeName == "" {
		for _, r := range roots {
			var t *Type
			switch r[0] {
			case "query":
				t = s.Query
			case "mutation":
				t = s.Mutation
			default:
				t = s.Subscription
			}
			if t != nil {
				renderType(t, r[1])
			}
		}
	}
	for _, t := range s.Types {
		if typeName != "" && !strings.EqualFold(t.Name, typeName) {
			continue
		}
		if t.Kind == "SCALAR" {
			if typeName != "" {
				sb.WriteString("scalar " + t.Name + "\n")
			}
			continue
		}
		if len(t.Fields) == 0 && len(t.EnumValues) == 0 {
			if typeName != "" {
				sb.WriteString(strings.ToLower(t.Kind) + " " + t.Name + "\n")
			}
			continue
		}
		if len(t.EnumValues) > 0 {
			sb.WriteString("enum " + t.Name + " { " + strings.Join(t.EnumValues, " | ") + " }")
			sb.WriteString("\n")
			continue
		}
		sb.WriteString("type " + t.Name + "\n")
		for _, f := range t.Fields {
			sb.WriteString("  ")
			sb.WriteString(f.Signature())
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
