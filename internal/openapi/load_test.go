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

package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

const yamlSpec = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: A user
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
`

const jsonSpec = `{
  "openapi": "3.0.3",
  "info": {"title": "Test API", "version": "1.0.0"},
  "paths": {
    "/ping": {
      "get": {
        "operationId": "ping",
        "responses": {"200": {"description": "pong"}}
      }
    }
  }
}
`

func TestLoadYAML(t *testing.T) {
	doc, err := Load([]byte(yamlSpec))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	op := doc.Paths.Value("/users/{id}").GetOperation("GET")
	if op == nil {
		t.Fatal("expected GET /users/{id} operation")
	}
	if op.OperationID != "getUser" {
		t.Fatalf("OperationID = %q, want getUser", op.OperationID)
	}
	if doc.Components.Schemas["User"] == nil {
		t.Fatal("expected User schema in components")
	}
}

func TestLoadJSON(t *testing.T) {
	doc, err := Load([]byte(jsonSpec))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if doc.Paths.Value("/ping").GetOperation("GET") == nil {
		t.Fatal("expected GET /ping operation")
	}
}

func TestLoadRejectsNonOpenAPI(t *testing.T) {
	_, err := Load([]byte("not: an openapi document"))
	if err == nil {
		t.Fatal("expected error for non-OpenAPI input")
	}
}

func TestLoadInvalidDocument(t *testing.T) {
	_, err := Load([]byte(`{"openapi": "3.0.3", "info": {"title": "x"}}`))
	if err == nil {
		t.Fatal("expected validation error for missing version/paths")
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(path, []byte(yamlSpec), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	doc, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if doc.Paths.Value("/users/{id}") == nil {
		t.Fatal("expected /users/{id} path")
	}
}

func TestLoadFileMissing(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadRefsResolved(t *testing.T) {
	doc, err := Load([]byte(yamlSpec))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	mt := doc.Paths.Value("/users/{id}").GetOperation("GET").Responses.Map()["200"].Value.Content["application/json"]
	if mt.Schema == nil || mt.Schema.Value == nil {
		t.Fatal("expected resolved schema ref on response media type")
	}
	if len(mt.Schema.Value.Type.Slice()) != 1 || !mt.Schema.Value.Type.Is("object") {
		t.Fatalf("resolved schema type = %v, want object", mt.Schema.Value.Type)
	}
}
