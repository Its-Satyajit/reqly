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
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

const generateSpec = `openapi: 3.0.3
info:
  title: Generate API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /users/{id}:
    get:
      operationId: getUser
      summary: Get a user
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          example: "42"
          schema:
            type: string
        - name: verbose
          in: query
          required: false
          schema:
            type: boolean
        - name: X-Trace
          in: header
          required: true
          schema:
            type: string
            default: trace-1
      responses:
        "200":
          description: ok
    post:
      operationId: updateUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        content:
          application/json:
            example:
              name: Ada
      responses:
        "200":
          description: ok
  /files:
    post:
      summary: Upload a file
      requestBody:
        content:
          multipart/form-data:
            schema:
              type: object
      responses:
        "201":
          description: created
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
`

func TestGenerateByOperation(t *testing.T) {
	doc := mustDoc(t, generateSpec)
	files, warnings, err := Generate(doc, GenerateOptions{Operations: []string{"getUser"}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(files) != 1 || files[0].Filename != "getUser" {
		t.Fatalf("got %+v, want single file getUser", files)
	}
	f := files[0].File
	if f.Request.Method != "GET" {
		t.Errorf("method = %q, want GET", f.Request.Method)
	}
	if f.Request.URL != "{{baseUrl}}/users/42" {
		t.Errorf("URL = %q, want {{baseUrl}}/users/42 (path param example substituted)", f.Request.URL)
	}
	if f.Variables["baseUrl"] != "https://api.example.com/v1" {
		t.Errorf("baseUrl variable = %q, want server URL", f.Variables["baseUrl"])
	}
	if f.Variables["id"] != "" {
		t.Errorf("id variable should not be declared once resolved, got %q", f.Variables["id"])
	}
	if f.Variables["X-Trace"] != "" {
		t.Errorf("X-Trace variable = %q, want empty declared variable for default-filled header param", f.Variables["X-Trace"])
	}
	var hasQuery, hasHeader bool
	for _, q := range f.Request.Query {
		if q.Key == "verbose" {
			hasQuery = true
		}
	}
	for _, h := range f.Request.Headers {
		if h.Key == "X-Trace" && h.Value == "trace-1" {
			hasHeader = true
		}
	}
	if hasQuery {
		t.Error("optional query param should be omitted")
	}
	if !hasHeader {
		t.Error("required header param with default should be filled with trace-1")
	}
	if f.Request.Auth.Type != "bearer" {
		t.Errorf("auth type = %q, want bearer", f.Request.Auth.Type)
	}
	if f.Request.Auth.Config["token"] != "{{token}}" {
		t.Errorf("bearer token = %q, want {{token}} placeholder", f.Request.Auth.Config["token"])
	}
	if f.Variables["token"] != "" {
		t.Errorf("token variable = %q, want empty declared variable", f.Variables["token"])
	}
	_ = warnings
}

func TestGenerateJSONBodyInline(t *testing.T) {
	doc := mustDoc(t, generateSpec)
	files, _, err := Generate(doc, GenerateOptions{Operations: []string{"updateUser"}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	f := files[0].File
	if !strings.Contains(f.Request.Body, `"Ada"`) {
		t.Errorf("body = %q, want inline example JSON containing \"Ada\"", f.Request.Body)
	}
	var ct string
	for _, h := range f.Request.Headers {
		if strings.EqualFold(h.Key, "Content-Type") {
			ct = h.Value
		}
	}
	if !strings.Contains(ct, "json") {
		t.Errorf("Content-Type = %q, want json media type", ct)
	}
}

func TestGenerateNonJSONBodyWarns(t *testing.T) {
	doc := mustDoc(t, generateSpec)
	files, warnings, err := Generate(doc, GenerateOptions{Method: "POST", Path: "/files"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "multipart/form-data") {
		t.Errorf("warnings = %v, want mention of skipped multipart/form-data body", warnings)
	}
}

func TestGenerateMethodPathSelector(t *testing.T) {
	doc := mustDoc(t, generateSpec)
	files, _, err := Generate(doc, GenerateOptions{Method: "GET", Path: "/users/{id}"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(files) != 1 || files[0].Filename != "getUser" {
		t.Fatalf("got %+v, want getUser via method+path selector", files)
	}
}

func TestGenerateTagAndAllSelectors(t *testing.T) {
	tagged := `openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /a:
    get:
      operationId: opA
      tags: [x]
      responses: {"200": {description: ok}}
  /b:
    get:
      operationId: opB
      responses: {"200": {description: ok}}
`
	doc := mustDoc(t, tagged)
	files, _, err := Generate(doc, GenerateOptions{Tags: []string{"x"}})
	if err != nil || len(files) != 1 || files[0].Filename != "opA" {
		t.Fatalf("tag selector got files=%+v err=%v, want [opA]", files, err)
	}
	files, _, err = Generate(doc, GenerateOptions{All: true})
	if err != nil || len(files) != 2 {
		t.Fatalf("--all got files=%+v err=%v, want 2 files", files, err)
	}
}

func TestGenerateFallbackFilenameAndCollisions(t *testing.T) {
	collide := `openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  "/x/y":
    get:
      operationId: same op
      responses: {"200": {description: ok}}
  "/x/z":
    get:
      operationId: same-op
      responses: {"200": {description: ok}}
  "/w":
    get:
      responses: {"200": {description: ok}}
`
	doc := mustDoc(t, collide)
	files, warnings, err := Generate(doc, GenerateOptions{All: true})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	names := map[string]bool{}
	for _, f := range files {
		names[f.Filename] = true
	}
	for _, want := range []string{"same-op", "same-op-2", "get-w"} {
		if !names[want] {
			t.Errorf("missing generated file %q; have %v", want, names)
		}
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "same-op-2") {
		t.Errorf("warnings = %v, want collision warning mentioning same-op-2", warnings)
	}
}

func TestGenerateUnresolvedPathParamStaysLiteral(t *testing.T) {
	unresolved := `openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  "/items/{itemId}":
    get:
      operationId: getItem
      parameters:
        - name: itemId
          in: path
          required: true
          schema:
            type: string
      responses: {"200": {description: ok}}
`
	doc := mustDoc(t, unresolved)
	files, warnings, err := Generate(doc, GenerateOptions{Operations: []string{"getItem"}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	f := files[0].File
	if f.Request.URL != "/items/{itemId}" {
		t.Errorf("URL = %q, want literal /items/{itemId}", f.Request.URL)
	}
	if v, ok := f.Variables["itemId"]; !ok || v != "" {
		t.Errorf("itemId variable = %q (present=%v), want empty declared variable", v, ok)
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "itemId") {
		t.Errorf("warnings = %v, want unfilled-param warning mentioning itemId", warnings)
	}
}

func TestGenerateNoServersOmitsBaseUrl(t *testing.T) {
	noServers := `openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /ping:
    get:
      operationId: ping
      responses: {"200": {description: ok}}
`
	doc := mustDoc(t, noServers)
	files, _, err := Generate(doc, GenerateOptions{Operations: []string{"ping"}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if files[0].File.Request.URL != "/ping" {
		t.Errorf("URL = %q, want /ping without baseUrl template", files[0].File.Request.URL)
	}
}

func TestGenerateOAuth2Warns(t *testing.T) {
	oauth := `openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /a:
    get:
      operationId: opA
      security:
        - oauthFlow: []
      responses: {"200": {description: ok}}
components:
  securitySchemes:
    oauthFlow:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://auth.example.com/token
          scopes: {}
`
	doc := mustDoc(t, oauth)
	files, warnings, err := Generate(doc, GenerateOptions{Operations: []string{"opA"}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1 (generation continues despite unmappable auth)", len(files))
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "oauth2") {
		t.Errorf("warnings = %v, want warning about unsupported oauth2 scheme", warnings)
	}
}

func TestGenerateAPIKeyHeaderScheme(t *testing.T) {
	apikey := `openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /a:
    get:
      operationId: opA
      security:
        - keyAuth: []
      responses: {"200": {description: ok}}
components:
  securitySchemes:
    keyAuth:
      type: apiKey
      in: header
      name: X-Api-Key
`
	doc := mustDoc(t, apikey)
	files, _, err := Generate(doc, GenerateOptions{Operations: []string{"opA"}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	auth := files[0].File.Request.Auth
	if auth.Type != "apikey" {
		t.Fatalf("auth type = %q, want apikey", auth.Type)
	}
	if auth.Config["key"] != "X-Api-Key" || auth.Config["value"] != "{{apiKey}}" || auth.Config["in"] != "header" {
		t.Errorf("apikey config = %v, want X-Api-Key/{{apiKey}}/header", auth.Config)
	}
	if v, ok := files[0].File.Variables["apiKey"]; !ok || v != "" {
		t.Errorf("apiKey variable = %q (present=%v), want empty declared variable", v, ok)
	}
}

func TestWriteGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	doc := mustDoc(t, generateSpec)
	files, _, err := Generate(doc, GenerateOptions{Operations: []string{"getUser"}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	paths, err := Write(dir, files)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("Write() returned %d paths, want 1", len(paths))
	}
	rf, err := requestfile.LoadFile(filepath.Join(dir, "getUser.yaml"))
	if err != nil {
		t.Fatalf("generated file does not parse as a request file: %v", err)
	}
	if rf.Request.URL != "{{baseUrl}}/users/42" {
		t.Errorf("round-trip URL = %q", rf.Request.URL)
	}
	if _, err := os.Stat(paths[0]); err != nil {
		t.Errorf("written file missing: %v", err)
	}
}
