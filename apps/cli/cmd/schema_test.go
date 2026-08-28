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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetSchemaFlags() {
	schemaValidateDraft = ""
	schemaValidateType = ""
	schemaValidateJSON = false
	schemaInspectJSON = false
	schemaGenerateSeed = 0
	schemaGenerateOptIncl = false
}

const cliSchemaJSON = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["id"],
  "properties": {
    "id": {"type": "integer"},
    "role": {"enum": ["admin", "user"]}
  }
}`

func writeSchemaFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSchemaValidatePass(t *testing.T) {
	resetSchemaFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	schemaPath := writeSchemaFile(t, cliSchemaJSON)
	instancePath := filepath.Join(filepath.Dir(schemaPath), "ok.json")
	if err := os.WriteFile(instancePath, []byte(`{"id": 1, "role": "user"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"schema", "validate", schemaPath, instancePath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("valid instance should pass: %v", err)
	}
	if !strings.Contains(out.String(), "0 violation(s)") {
		t.Errorf("unexpected output:\n%s", out.String())
	}
}

func TestSchemaValidateViolationsExitNonZero(t *testing.T) {
	resetSchemaFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	schemaPath := writeSchemaFile(t, cliSchemaJSON)
	instancePath := filepath.Join(filepath.Dir(schemaPath), "bad.json")
	if err := os.WriteFile(instancePath, []byte(`{"id": "nope"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"schema", "validate", schemaPath, instancePath})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit for violations")
	}
	for _, want := range []string{"$.id", "violation"} {
		if !strings.Contains(out.String()+err.Error(), want) {
			t.Errorf("output missing %q:\n%s", want, out.String())
		}
	}
}

func TestSchemaValidateJSONOutput(t *testing.T) {
	resetSchemaFlags()
	schemaValidateJSON = true
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	schemaPath := writeSchemaFile(t, cliSchemaJSON)
	instancePath := filepath.Join(filepath.Dir(schemaPath), "bad.json")
	if err := os.WriteFile(instancePath, []byte(`{"id": "nope"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"schema", "validate", schemaPath, instancePath})
	rootCmd.Execute() // expected to fail; output is what matters

	var violations []map[string]any
	if err := json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&violations); err != nil {
		t.Fatalf("--json output is not JSON: %v\n%s", err, out.String())
	}
	if len(violations) == 0 || violations[0]["path"] != "$.id" {
		t.Fatalf("unexpected violations: %s", out.String())
	}
}

func TestSchemaValidateStdin(t *testing.T) {
	resetSchemaFlags()
	schemaPath := writeSchemaFile(t, cliSchemaJSON)

	defer func(orig *os.File) { os.Stdin = orig }(os.Stdin)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	if _, err := w.WriteString(`{"id": 5}`); err != nil {
		t.Fatal(err)
	}
	w.Close()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"schema", "validate", schemaPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("stdin validation should pass: %v", err)
	}
}

func TestSchemaInspectOutput(t *testing.T) {
	resetSchemaFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"schema", "inspect", writeSchemaFile(t, cliSchemaJSON)})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"object", "id", "required", "enum"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("inspect output missing %q:\n%s", want, out.String())
		}
	}
}

func TestSchemaGenerateOutput(t *testing.T) {
	resetSchemaFlags()
	schemaGenerateOptIncl = true
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"schema", "generate", writeSchemaFile(t, cliSchemaJSON), "--seed", "3"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var inst map[string]any
	if err := json.Unmarshal(out.Bytes(), &inst); err != nil {
		t.Fatalf("generate output is not JSON: %v\n%s", err, out.String())
	}
	if inst["id"] == nil || inst["role"] != "admin" {
		t.Fatalf("unexpected generated instance: %s", out.String())
	}
}
