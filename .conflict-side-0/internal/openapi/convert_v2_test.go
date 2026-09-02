// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"strings"
	"testing"
)

func TestConvertSwagger2ToOpenAPI3(t *testing.T) {
	swaggerJSON := []byte(`{
  "swagger": "2.0",
  "info": {
    "title": "Legacy API",
    "version": "1.0.0"
  },
  "host": "api.example.com",
  "basePath": "/v1",
  "schemes": ["https"],
  "paths": {
    "/ping": {
      "get": {
        "summary": "Ping endpoint"
      }
    }
  }
}`)

	converted, err := ConvertSwagger2ToOpenAPI3(swaggerJSON)
	if err != nil {
		t.Fatalf("unexpected conversion error: %v", err)
	}
	if !strings.Contains(string(converted), "openapi: 3.0.3") {
		t.Errorf("expected openapi: 3.0.3 header, got:\n%s", string(converted))
	}
	if !strings.Contains(string(converted), "https://api.example.com/v1") {
		t.Errorf("expected converted server URL, got:\n%s", string(converted))
	}
}
