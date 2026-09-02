// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGraphQLParseCmd(t *testing.T) {
	dir := t.TempDir()
	sdlPath := filepath.Join(dir, "schema.graphql")
	content := `
type Query {
  user: User
}
type User {
  name: String
}
`
	if err := os.WriteFile(sdlPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write sdl: %v", err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"graphql", "parse", sdlPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected failure executing reqly graphql parse: %v", err)
	}
}
