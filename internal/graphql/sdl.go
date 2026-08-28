// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"strings"
)

// ParseSDL parses raw GraphQL Schema Definition Language (SDL) text into a Schema model.
func ParseSDL(sdlText string) (*Schema, error) {
	lines := strings.Split(sdlText, "\n")
	s := &Schema{}

	var currentType *Type
	var inTypeBlock bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "type ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				name := strings.TrimSuffix(parts[1], "{")
				name = strings.TrimSpace(name)
				currentType = &Type{
					Kind: "OBJECT",
					Name: name,
				}
				inTypeBlock = true

				if name == "Query" {
					s.Query = currentType
				} else if name == "Mutation" {
					s.Mutation = currentType
				} else if name == "Subscription" {
					s.Subscription = currentType
				} else {
					s.Types = append(s.Types, currentType)
				}
			}
			continue
		}

		if trimmed == "}" {
			inTypeBlock = false
			currentType = nil
			continue
		}

		if inTypeBlock && currentType != nil {
			fieldParts := strings.SplitN(trimmed, ":", 2)
			if len(fieldParts) == 2 {
				fieldName := strings.TrimSpace(fieldParts[0])
				fieldTypeRaw := strings.TrimSpace(fieldParts[1])
				currentType.Fields = append(currentType.Fields, Field{
					Name: fieldName,
					Type: &TypeRef{Name: fieldTypeRaw, Kind: "NAMED"},
				})
			}
		}
	}

	return s, nil
}
