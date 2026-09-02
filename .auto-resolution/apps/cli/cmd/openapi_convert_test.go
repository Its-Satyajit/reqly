// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIConvertV2Cmd(t *testing.T) {
	dir := t.TempDir()
	swaggerPath := filepath.Join(dir, "legacy.json")
	content := `{
  "swagger": "2.0",
  "info": {"title": "Legacy API", "version": "1.0.0"},
  "host": "api.test.com",
  "basePath": "/v1",
  "paths": {"/ping": {"get": {"summary": "ping"}}}
}`
	if err := os.WriteFile(swaggerPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write swagger spec: %v", err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"openapi", "convert-v2", swaggerPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing openapi convert-v2: %v", err)
	}
	if !strings.Contains(buf.String(), "openapi: 3.0.3") {
		t.Errorf("expected openapi: 3.0.3 header, got:\n%s", buf.String())
	}
}
