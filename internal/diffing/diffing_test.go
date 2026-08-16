// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package diffing_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/diffing"
)

func TestDiffJSON(t *testing.T) {
	json1 := `{"name": "Alice", "age": 30}`
	json2 := `{"name": "Alice", "age": 31, "role": "admin"}`

	diff, err := diffing.JSON([]byte(json1), []byte(json2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !diff.HasChanges {
		t.Fatalf("expected diff changes")
	}
}

func TestDiffOpenAPI(t *testing.T) {
	tmpDir := t.TempDir()

	spec1 := `
openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
`
	spec2 := `
openapi: 3.0.0
info:
  title: API
  version: 1.1.0
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
  /posts:
    get:
      responses:
        '200':
          description: OK
`

	path1 := filepath.Join(tmpDir, "spec1.yaml")
	path2 := filepath.Join(tmpDir, "spec2.yaml")

	os.WriteFile(path1, []byte(spec1), 0644)
	os.WriteFile(path2, []byte(spec2), 0644)

	diff, err := diffing.OpenAPIFiles(path1, path2)
	if err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}

	if !diff.HasChanges {
		t.Fatalf("expected spec diff changes")
	}
}
