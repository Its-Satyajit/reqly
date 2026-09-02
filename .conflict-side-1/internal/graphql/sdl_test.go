// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"testing"
)

func TestParseSDL(t *testing.T) {
	sdl := `
type Query {
  user(id: ID!): User
}

type User {
  id: ID!
  name: String!
  email: String
}
`
	schema, err := ParseSDL(sdl)
	if err != nil {
		t.Fatalf("unexpected error parsing SDL: %v", err)
	}
	if schema == nil {
		t.Fatalf("expected non-nil schema")
	}
	if schema.Query == nil || schema.Query.Name != "Query" {
		t.Errorf("expected Query type, got %v", schema.Query)
	}
	if len(schema.Types) != 1 || schema.Types[0].Name != "User" {
		t.Errorf("expected 1 custom type 'User', got %d", len(schema.Types))
	}
}
