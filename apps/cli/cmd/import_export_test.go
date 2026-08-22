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

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const testOpenAPISpec = `openapi: 3.0.3
info:
  title: Petstore
  version: "1.0.0"
servers:
  - url: https://api.petstore.example.com/v1
paths:
  /pets:
    get:
      operationId: listPets
      tags: [pets]
    post:
      operationId: createPet
      tags: [pets]
      requestBody:
        content:
          application/json: {}
`

func TestImportCurlPrints(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"import", "curl", "curl -X POST https://api.example.com/users -H 'Content-Type: application/json' -d '{}'"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	for _, want := range []string{"POST https://api.example.com/users", "Content-Type: application/json", "{}"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestImportCurlWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "req.yaml")
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"import", "curl", "curl https://api.example.com/users", "--output", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "users") {
		t.Fatalf("request file missing URL:\n%s", data)
	}
}

func TestImportCurlInvalid(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"import", "curl", "curl -F 'file=@a' https://api.example.com/"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for unsupported flag")
	}
}

func TestImportOpenAPI(t *testing.T) {
	spec := filepath.Join(t.TempDir(), "spec.yaml")
	if err := os.WriteFile(spec, []byte(testOpenAPISpec), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(t.TempDir(), "ws")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"import", "openapi", spec, "--output", outDir})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{
		"reqly.yaml",
		filepath.Join("collections", "pets", "reqly.yaml"),
		filepath.Join("collections", "pets", "listPets.yaml"),
		filepath.Join("collections", "pets", "createPet.yaml"),
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
}

func TestExportPostman(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	outPath := filepath.Join(t.TempDir(), "collection.json")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"export", "postman", root, "--output", outPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{`"schema"`, `"method": "GET"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in export, got:\n%s", want, text)
		}
	}
	if !strings.Contains(out.String(), "2 requests") {
		t.Fatalf("expected count in output, got:\n%s", out.String())
	}
}

func TestExportPostmanInvalidWorkspace(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"export", "postman", t.TempDir()})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for non-workspace dir")
	}
}

// resetImportExportFlags clears package-level flag state across tests.
func resetImportExportFlags() {
	importOutput = ""
	exportOutput = ""
	for _, cmd := range []*cobra.Command{importCurlCmd, importOpenAPICmd, exportPostmanCmd} {
		if flag := cmd.Flags().Lookup("output"); flag != nil {
			flag.Changed = false
		}
	}
}
