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

package exporter

import (
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
)

func openapiFixtureRequests() []request.Request {
	return []request.Request{
		{
			Name:    "list users",
			Method:  request.MethodGet,
			URL:     "{{baseUrl}}/users?page=1",
			Query:   []request.Parameter{{Key: "page", Value: "1"}},
			Headers: []request.Header{{Key: "Accept", Value: "application/json"}},
		},
		{
			Name:    "create user",
			Method:  request.MethodPost,
			URL:     "https://api.example.com/users",
			Headers: []request.Header{{Key: "Content-Type", Value: "application/json"}},
			Body:    `{"name":"ada"}`,
		},
		{
			Name:   "basic ping",
			Method: request.MethodGet,
			URL:    "https://api.example.com/ping",
			Auth:   request.Auth{Type: "basic", Config: map[string]string{"username": "u", "password": "p"}},
		},
		{
			Name:   "bearer me",
			Method: request.MethodGet,
			URL:    "https://api.example.com/me",
			Auth:   request.Auth{Type: "bearer", Config: map[string]string{"token": "t"}},
		},
	}
}

func TestExportOpenAPIBasics(t *testing.T) {
	out, err := ExportOpenAPI("Demo API", "https://api.example.com", openapiFixtureRequests())
	if err != nil {
		t.Fatal(err)
	}
	doc := string(out)
	for _, want := range []string{
		`openapi: 3.0.3`,
		`title: Demo API`,
		`servers:`,
		`url: https://api.example.com`,
		`/users:`,
		`operationId: list_users`,
		`name: page`,
		`in: query`,
		`application/json`,
		`requestBody:`,
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("spec missing %q:\n%s", want, doc)
		}
	}
}

func TestExportOpenAPISecuritySchemes(t *testing.T) {
	out, err := ExportOpenAPI("Demo", "", openapiFixtureRequests())
	if err != nil {
		t.Fatal(err)
	}
	doc := string(out)
	for _, want := range []string{
		`components:`,
		`securitySchemes:`,
		`scheme: basic`,
		`scheme: bearer`,
		`security:`,
		`basicAuth: []`,
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("spec missing %q:\n%s", want, doc)
		}
	}
	// Templated URL without base must be skipped.
	if strings.Contains(doc, "%7B%7BbaseUrl%7D%7D") || strings.Contains(doc, "{{baseUrl}}/") {
		t.Fatalf("templated URL leaked into paths:\n%s", doc)
	}
}

func TestExportOpenAPIAPIKeyScheme(t *testing.T) {
	reqs := []request.Request{{
		Name: "keyed",
		URL:  "https://api.example.com/x",
		Auth: request.Auth{Type: "apikey", Config: map[string]string{"key": "X-Key", "value": "v", "in": "header"}},
	}}
	out, err := ExportOpenAPI("K", "", reqs)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(out)
	for _, want := range []string{`type: apiKey`, `name: X-Key`, `in: header`} {
		if !strings.Contains(doc, want) {
			t.Errorf("spec missing %q:\n%s", want, doc)
		}
	}
}

func TestExportOpenAPIEmpty(t *testing.T) {
	out, err := ExportOpenAPI("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(out)
	if !strings.Contains(doc, "Reqly Collection") || !strings.Contains(doc, "paths:") {
		t.Fatalf("empty export =\n%s", doc)
	}
	if strings.Contains(doc, "servers:") {
		t.Fatalf("no requests → no servers:\n%s", doc)
	}
}
