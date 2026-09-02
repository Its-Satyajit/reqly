// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package importer

import (
	"testing"
)

func TestParseSwagger2(t *testing.T) {
	swaggerJSON := []byte(`{
  "swagger": "2.0",
  "info": {
    "title": "Sample Legacy API",
    "version": "1.0.0"
  },
  "host": "api.legacy.com",
  "basePath": "/v1",
  "schemes": ["https"],
  "paths": {
    "/users": {
      "get": {
        "summary": "Get all users",
        "operationId": "getUsers",
        "tags": ["Users"]
      }
    }
  }
}`)

	doc, err := ParseSwagger2(swaggerJSON)
	if err != nil {
		t.Fatalf("unexpected error parsing Swagger 2.0: %v", err)
	}
	if doc == nil || doc.Swagger != "2.0" {
		t.Fatalf("expected Swagger 2.0 document")
	}

	res := doc.ToOpenAPIResult()
	if res.Title != "Sample Legacy API" {
		t.Errorf("expected title 'Sample Legacy API', got %q", res.Title)
	}
	if res.BaseURL != "https://api.legacy.com/v1" {
		t.Errorf("expected base URL 'https://api.legacy.com/v1', got %q", res.BaseURL)
	}
	if len(res.Collections) != 1 || res.Collections[0].Name != "Users" {
		t.Errorf("expected 1 collection 'Users'")
	}
}
