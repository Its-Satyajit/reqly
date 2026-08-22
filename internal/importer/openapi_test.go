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

package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

const petstoreYAML = `openapi: 3.0.3
info:
  title: Petstore
  version: "1.0.0"
servers:
  - url: https://api.petstore.example.com/v1
paths:
  /pets:
    get:
      operationId: listPets
      summary: List all pets
      tags: [pets]
      parameters:
        - name: limit
          in: query
          schema: {type: integer}
    post:
      operationId: createPet
      summary: Create a pet
      tags: [pets]
      requestBody:
        content:
          application/json:
            schema: {type: object}
  /pets/{petId}:
    get:
      operationId: getPet
      summary: Get a pet
      tags: [pets]
      parameters:
        - name: petId
          in: path
          required: true
          schema: {type: integer}
        - name: X-Trace
          in: header
          schema: {type: string}
  /health:
    get:
      summary: Health check
`

func TestParseOpenAPIYAML(t *testing.T) {
	doc, err := ParseOpenAPI([]byte(petstoreYAML))
	if err != nil {
		t.Fatal(err)
	}
	if doc.OpenAPI != "3.0.3" {
		t.Fatalf("openapi version: got %q", doc.OpenAPI)
	}
	result := doc.ToOpenAPIResult()
	if result.Title != "Petstore" {
		t.Fatalf("title: got %q", result.Title)
	}
	if result.BaseURL != "https://api.petstore.example.com/v1" {
		t.Fatalf("baseURL: got %q", result.BaseURL)
	}
	if len(result.Collections) != 2 {
		t.Fatalf("collections: got %d", len(result.Collections))
	}
	// "pets" sorted before "default" (health has no tags).
	if result.Collections[0].Name != "default" || result.Collections[1].Name != "pets" {
		t.Fatalf("collection order: got %q, %q", result.Collections[0].Name, result.Collections[1].Name)
	}
}

func TestParseOpenAPIJSON(t *testing.T) {
	doc, err := ParseOpenAPI([]byte(`{"openapi":"3.1.0","info":{"title":"J","version":"1"},"paths":{"/a":{"get":{"operationId":"getA"}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if doc.OpenAPI != "3.1.0" {
		t.Fatalf("version: got %q", doc.OpenAPI)
	}
}

func TestParseOpenAPINotASpec(t *testing.T) {
	if _, err := ParseOpenAPI([]byte(`{"foo": 1}`)); err == nil {
		t.Fatal("expected error for non-OpenAPI document")
	}
}

func TestToOpenAPIResult(t *testing.T) {
	doc, err := ParseOpenAPI([]byte(petstoreYAML))
	if err != nil {
		t.Fatal(err)
	}
	result := doc.ToOpenAPIResult()

	// pets collection has 3 operations.
	pets := result.Collections[1]
	if len(pets.Request) != 3 {
		t.Fatalf("pets requests: got %d, want 3", len(pets.Request))
	}

	// listPets: GET, query param limit.
	var listPets *requestfile.File
	for _, f := range pets.Request {
		if f.Name == "listPets" {
			listPets = f
		}
	}
	if listPets == nil {
		t.Fatal("listPets not found")
	}
	if listPets.Request.Method != request.MethodGet {
		t.Fatalf("listPets method: got %q", listPets.Request.Method)
	}
	if listPets.Request.URL != "/pets" {
		t.Fatalf("listPets url: got %q", listPets.Request.URL)
	}
	if len(listPets.Request.Query) != 1 || listPets.Request.Query[0].Key != "limit" {
		t.Fatalf("listPets query: got %+v", listPets.Request.Query)
	}

	// createPet: POST, JSON body + Content-Type header.
	var createPet *requestfile.File
	for _, f := range pets.Request {
		if f.Name == "createPet" {
			createPet = f
		}
	}
	if createPet == nil {
		t.Fatal("createPet not found")
	}
	if createPet.Request.Method != request.MethodPost {
		t.Fatalf("createPet method: got %q", createPet.Request.Method)
	}
	if createPet.Request.Body != "{}" {
		t.Fatalf("createPet body: got %q", createPet.Request.Body)
	}
	hasCT := false
	for _, h := range createPet.Request.Headers {
		if h.Key == "Content-Type" {
			hasCT = true
		}
	}
	if !hasCT {
		t.Fatal("createPet missing Content-Type header")
	}

	// getPet: path param kept in URL, header param as header.
	var getPet *requestfile.File
	for _, f := range pets.Request {
		if f.Name == "getPet" {
			getPet = f
		}
	}
	if getPet == nil {
		t.Fatal("getPet not found")
	}
	if getPet.Request.URL != "/pets/{petId}" {
		t.Fatalf("getPet url: got %q", getPet.Request.URL)
	}
	hasTrace := false
	for _, h := range getPet.Request.Headers {
		if h.Key == "X-Trace" {
			hasTrace = true
		}
	}
	if !hasTrace {
		t.Fatal("getPet missing X-Trace header")
	}

	// health (no tags) → default collection, filename from method+path.
	def := result.Collections[0]
	if len(def.Request) != 1 {
		t.Fatalf("default requests: got %d", len(def.Request))
	}
	if def.Request[0].Name != "get-health" {
		t.Fatalf("health filename: got %q", def.Request[0].Name)
	}
}

func TestOpenAPIResultWrite(t *testing.T) {
	doc, err := ParseOpenAPI([]byte(petstoreYAML))
	if err != nil {
		t.Fatal(err)
	}
	result := doc.ToOpenAPIResult()

	dir := t.TempDir()
	if err := result.Write(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "reqly.yaml")); err != nil {
		t.Fatalf("workspace descriptor missing: %v", err)
	}
	petsDir := filepath.Join(dir, "collections", "pets")
	if _, err := os.Stat(filepath.Join(petsDir, "reqly.yaml")); err != nil {
		t.Fatalf("collection descriptor missing: %v", err)
	}
	for _, name := range []string{"listPets.yaml", "createPet.yaml", "getPet.yaml"} {
		if _, err := os.Stat(filepath.Join(petsDir, name)); err != nil {
			t.Fatalf("request file %s missing: %v", name, err)
		}
	}

	// Round-trip: load the written workspace.
	data, err := os.ReadFile(filepath.Join(dir, "reqly.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "baseURL") {
		t.Fatalf("descriptor missing baseURL:\n%s", data)
	}
}

func TestOpenAPIResultURL(t *testing.T) {
	r := &OpenAPIResult{BaseURL: "https://api.example.com/v1"}
	got, err := r.URL("/pets")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://api.example.com/v1/pets" {
		t.Fatalf("url: got %q", got)
	}
}

func TestSanitizeName(t *testing.T) {
	cases := map[string]string{
		"List All Pets": "List-All-Pets",
		"a/b":           "a-b",
		"  spaced  ":    "spaced",
	}
	for in, want := range cases {
		if got := sanitizeName(in); got != want {
			t.Fatalf("sanitize(%q): got %q, want %q", in, got, want)
		}
	}
}

func TestOperationFilename(t *testing.T) {
	if got := operationFilename("get", "", "/pets/{petId}"); got != "get-pets-petId" {
		t.Fatalf("got %q", got)
	}
	if got := operationFilename("post", "createPet", "/pets"); got != "createPet" {
		t.Fatalf("got %q", got)
	}
	if got := operationFilename("get", "", "/"); got != "get-root" {
		t.Fatalf("got %q", got)
	}
}
