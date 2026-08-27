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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func resetOpenAPIFlags() {
	openapiExploreTags = nil
	openapiExploreJSON = false
	openapiGenerateOps = nil
	openapiGenerateTags = nil
	openapiGenerateMethod = ""
	openapiGeneratePath = ""
	openapiGenerateAll = false
	openapiGenerateOut = ""
}

const cliOpenAPISpec = `openapi: 3.0.3
info:
  title: CLI API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
security:
  - bearerAuth: []
paths:
  /users/{id}:
    get:
      operationId: getUser
      summary: Get a user
      tags: [users]
      parameters:
        - name: id
          in: path
          required: true
          example: "7"
          schema:
            type: string
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
`

func writeCliSpec(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spec.yaml")
	if err := os.WriteFile(path, []byte(cliOpenAPISpec), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOpenAPIExploreTable(t *testing.T) {
	resetOpenAPIFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"openapi", "explore", writeCliSpec(t)})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"METHOD", "GET", "/users/{id}", "getUser", "Get a user"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("explore output missing %q:\n%s", want, out.String())
		}
	}
}

func TestOpenAPIExploreJSON(t *testing.T) {
	resetOpenAPIFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"openapi", "explore", writeCliSpec(t), "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var endpoints []map[string]any
	if err := json.Unmarshal(out.Bytes(), &endpoints); err != nil {
		t.Fatalf("explore --json is not JSON: %v\n%s", err, out.String())
	}
	if len(endpoints) != 1 || endpoints[0]["operationId"] != "getUser" {
		t.Fatalf("unexpected JSON endpoints: %s", out.String())
	}
}

func TestOpenAPIGenerateWritesFiles(t *testing.T) {
	resetOpenAPIFlags()
	outDir := filepath.Join(t.TempDir(), "requests")
	specPath := filepath.Join(t.TempDir(), "spec.yaml")
	if err := os.WriteFile(specPath, []byte(strings.Replace(cliOpenAPISpec, `example: "7"`, "", 1)), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs([]string{"openapi", "generate", specPath, "--operation", "getUser", "--output", outDir})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "getUser.yaml"))
	if err != nil {
		t.Fatalf("generated file missing: %v\nstdout:\n%s", err, out.String())
	}
	text := string(data)
	for _, want := range []string{"{{baseUrl}}/users/{id}", "bearer", "{{token}}", "baseUrl"} {
		if !strings.Contains(text, want) {
			t.Errorf("generated file missing %q:\n%s", want, text)
		}
	}
	if !strings.Contains(errOut.String(), "unfilled required parameters") {
		t.Errorf("stderr missing unfilled-parameter warning:\n%s", errOut.String())
	}
}

func TestOpenAPIGenerateNoSelectorErrors(t *testing.T) {
	resetOpenAPIFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"openapi", "generate", writeCliSpec(t), "--output", filepath.Join(t.TempDir(), "r")})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no selector is given")
	}
	if !strings.Contains(err.Error(), "getUser") {
		t.Errorf("error should list available operations, got: %v", err)
	}
}

func TestOpenAPIGenerateUnknownOperationErrors(t *testing.T) {
	resetOpenAPIFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"openapi", "generate", writeCliSpec(t), "--operation", "nope"})
	if err := rootCmd.Execute(); err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected unknown-operation error, got %v", err)
	}
}

func TestOpenAPIExploreAgainstImportSuiteFixture(t *testing.T) {
	resetOpenAPIFlags()
	fixture := filepath.Join("..", "..", "..", "internal", "importer", "testdata", "import-suite", "openapi", "fixtures", "openapi-with-examples.yaml")
	if _, err := os.Stat(fixture); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"openapi", "explore", fixture})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "PATH") {
		t.Fatalf("unexpected explore output for vendored fixture:\n%s", out.String())
	}
}

func TestOpenAPIValidateCommand(t *testing.T) {
	resetOpenAPIFlags()
	specFile := writeCliSpec(t)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"openapi", "validate", specFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("openapi validate failed: %v", err)
	}
	if !strings.Contains(out.String(), "cleanly") {
		t.Fatalf("expected validation success message, got:\n%s", out.String())
	}
}

func TestDiffCommandNoChanges(t *testing.T) {
	diffFailOnBreaking = false
	tmp := t.TempDir()
	f1 := filepath.Join(tmp, "a.json")
	f2 := filepath.Join(tmp, "b.json")
	_ = os.WriteFile(f1, []byte(`{"hello": "world"}`), 0644)
	_ = os.WriteFile(f2, []byte(`{"hello": "world"}`), 0644)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"diff", f1, f2})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("diff execution failed: %v", err)
	}
	if !strings.Contains(out.String(), "No structural changes found.") {
		t.Fatalf("expected no structural changes, got: %s", out.String())
	}
}

func TestDiffCommandOpenAPIBreakingChange(t *testing.T) {
	diffFailOnBreaking = false
	tmp := t.TempDir()
	specA := `openapi: 3.0.3
info: {title: API, version: "1"}
paths:
  /pets:
    get:
      responses:
        "200": {description: ok}
  /users:
    get:
      responses:
        "200": {description: ok}
`
	specB := `openapi: 3.0.3
info: {title: API, version: "1"}
paths:
  /pets:
    get:
      responses:
        "200": {description: ok}
`
	f1 := filepath.Join(tmp, "specA.yaml")
	f2 := filepath.Join(tmp, "specB.yaml")
	_ = os.WriteFile(f1, []byte(specA), 0644)
	_ = os.WriteFile(f2, []byte(specB), 0644)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"diff", f1, f2})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("diff execution failed: %v", err)
	}
	if !strings.Contains(out.String(), "breaking") {
		t.Fatalf("expected breaking change classification in output, got: %s", out.String())
	}
}

func TestDiffCommandFailOnBreakingFlag(t *testing.T) {
	diffFailOnBreaking = false
	tmp := t.TempDir()
	specA := `openapi: 3.0.3
info: {title: API, version: "1"}
paths:
  /pets:
    get:
      responses:
        "200": {description: ok}
  /users:
    get:
      responses:
        "200": {description: ok}
`
	specB := `openapi: 3.0.3
info: {title: API, version: "1"}
paths:
  /pets:
    get:
      responses:
        "200": {description: ok}
`
	f1 := filepath.Join(tmp, "specA.yaml")
	f2 := filepath.Join(tmp, "specB.yaml")
	_ = os.WriteFile(f1, []byte(specA), 0644)
	_ = os.WriteFile(f2, []byte(specB), 0644)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"diff", f1, f2, "--fail-on-breaking"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit error when breaking changes exist with --fail-on-breaking")
	}
	if !strings.Contains(err.Error(), "breaking change") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestPerfRunCommand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	perfRPS = 2
	perfDuration = 300 * time.Millisecond
	perfConcurrency = 1
	perfJSON = true

	reqFile := filepath.Join(t.TempDir(), "get.yaml")
	reqContent := fmt.Sprintf(`name: Ping
request:
  method: GET
  url: %s
`, srv.URL)
	if err := os.WriteFile(reqFile, []byte(reqContent), 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"perf", "run", reqFile, "--rps", "2", "--duration", "300ms", "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("perf run failed: %v", err)
	}
	if !strings.Contains(out.String(), `"rps"`) || !strings.Contains(out.String(), `"statusCounts"`) {
		t.Fatalf("expected JSON perf summary in stdout, got: %s", out.String())
	}
}

func TestMonitorRunCommand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	monitorInterval = 0
	monitorJSON = true

	reqFile := filepath.Join(t.TempDir(), "health.yaml")
	reqContent := fmt.Sprintf(`name: HealthCheck
request:
  method: GET
  url: %s
`, srv.URL)
	if err := os.WriteFile(reqFile, []byte(reqContent), 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"monitor", "run", reqFile, "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("monitor run failed: %v", err)
	}
	if !strings.Contains(out.String(), `"ok":true`) || !strings.Contains(out.String(), `"status":200`) {
		t.Fatalf("expected JSON monitor output, got: %s", out.String())
	}
}
