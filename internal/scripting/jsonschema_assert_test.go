// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package scripting

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSandbox_AssertJSONSchema(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "user.schema.json")
	schemaContent := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["id", "name"],
  "properties": {
    "id": {"type": "integer"},
    "name": {"type": "string"}
  }
}`

	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	sb := NewSandbox(SandboxOptions{})
	sb.BindResponse(&responseView{
		Body: `{"id": 1, "name": "Satyajit"}`,
	})

	script := `reqly.test("valid json schema", function() { return reqly.assertJSONSchema("` + schemaPath + `"); });`
	if err := sb.Run(script); err != nil {
		t.Fatalf("script run failed: %v", err)
	}

	tests := sb.Tests()
	if len(tests) != 1 || !tests[0].Fn() {
		t.Errorf("expected json schema assertion test to pass for valid response")
	}
}
