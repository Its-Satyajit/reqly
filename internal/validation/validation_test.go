// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/validation"
)

func TestValidateOpenAPI(t *testing.T) {
	tmpDir := t.TempDir()

	validSpec := `
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
`
	validPath := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validPath, []byte(validSpec), 0644); err != nil {
		t.Fatalf("failed writing file: %v", err)
	}

	res, err := validation.ValidateOpenAPIFile(validPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected valid spec, got errors: %v", res.Errors)
	}

	invalidSpec := `
openapi: 3.0.0
info:
  title: Sample API
paths:
  /users:
    get:
      responses:
        '200':
          $ref: '#/components/responses/NotFound'
`
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(invalidPath, []byte(invalidSpec), 0644); err != nil {
		t.Fatalf("failed writing file: %v", err)
	}

	res, err = validation.ValidateOpenAPIFile(invalidPath)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if res.Valid {
		t.Fatalf("expected invalid spec due to broken $ref")
	}
}

func TestValidateProject(t *testing.T) {
	tmpDir := t.TempDir()

	reqFile := `
name: Get User
request:
  method: GET
  url: https://api.example.com/user
`
	reqPath := filepath.Join(tmpDir, "request.json")
	if err := os.WriteFile(reqPath, []byte(reqFile), 0644); err != nil {
		t.Fatalf("failed writing file: %v", err)
	}

	res, err := validation.ValidateProject(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected valid project, got errors: %v", res.Errors)
	}
}
