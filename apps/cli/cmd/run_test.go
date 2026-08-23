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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/environments"
)

func TestRunCommandExecutesRequest(t *testing.T) {
	resetRunFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", srv.URL})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{"200 OK", "Content-Type: application/json", `{"ok":true}`} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestRunCommandSendsHeadersAndBody(t *testing.T) {
	resetRunFlags()
	var gotHeader, gotBody, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Test")
		gotMethod = r.Method
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		gotBody = string(buf)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{
		"run", "-m", "POST", "-H", "X-Test: hello", "-d", `{"a":1}`, srv.URL,
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if gotMethod != "POST" {
		t.Fatalf("expected POST, got %q", gotMethod)
	}
	if gotHeader != "hello" {
		t.Fatalf("expected header 'hello', got %q", gotHeader)
	}
	if gotBody != `{"a":1}` {
		t.Fatalf("expected body %q, got %q", `{"a":1}`, gotBody)
	}
}

func TestRunCommandInvalidHeader(t *testing.T) {
	resetRunFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", "-H", "BadHeader", "https://example.com"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for invalid header")
	}
}

func TestRunCommandNetworkError(t *testing.T) {
	resetRunFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", "http://127.0.0.1:1"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("unexpected error %q", err)
	}
}

func TestRunCommandMasksSecretsInRequestError(t *testing.T) {
	resetRunFlags()
	secret := "sup3r-s3cr3t-key"
	dir := t.TempDir()
	writeEnv(t, dir, "dev", "secrets:\n  API_KEY: "+secret+"\n")
	t.Chdir(dir)
	requestPath := filepath.Join(dir, "req.yaml")
	// The secret is interpolated into the URL; the connection to the host
	// fails, so the resulting *url.Error embeds the URL.
	if err := os.WriteFile(requestPath, []byte(`
environment: dev
request:
  method: GET
  url: http://127.0.0.1:1/{{API_KEY}}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", requestPath})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for unreachable host")
	} else if strings.Contains(err.Error(), secret) {
		t.Fatalf("secret leaked in request error: %q", err)
	}
}

func TestRunCommandMasksAuthConfigInOutput(t *testing.T) {
	resetRunFlags()
	token := "s3cr3t-bearer-token"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo the Authorization header back so a leaked token would surface.
		fmt.Fprint(w, r.Header.Get("Authorization"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	t.Chdir(dir)
	requestPath := filepath.Join(dir, "req.yaml")
	content := `request:
  method: GET
  url: ` + srv.URL + `
  auth:
    type: bearer
    config:
      token: ` + token + `
`
	if err := os.WriteFile(requestPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", requestPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	if strings.Contains(output, token) {
		t.Fatalf("auth token leaked in output: %q", output)
	}
	if !strings.Contains(output, environments.MaskedSecret) {
		t.Fatalf("expected [SECRET] in output, got %q", output)
	}
}

func TestRunCommandMasksAcquiredOAuthToken(t *testing.T) {
	resetRunFlags()
	accessToken := "oauth-access-token-abc123"
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"` + accessToken + `","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo the Authorization header back so a leaked token would surface.
		fmt.Fprint(w, r.Header.Get("Authorization"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	t.Chdir(dir)
	requestPath := filepath.Join(dir, "req.yaml")
	content := `request:
  method: GET
  url: ` + srv.URL + `
  auth:
    type: oauth2
    config:
      token_url: ` + tokenSrv.URL + `
      client_id: client-123
      client_secret: client-secret-value
`
	if err := os.WriteFile(requestPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", requestPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	if strings.Contains(output, accessToken) {
		t.Fatalf("acquired oauth token leaked in output: %q", output)
	}
	if strings.Contains(output, "client-secret-value") {
		t.Fatalf("client secret leaked in output: %q", output)
	}
	if !strings.Contains(output, environments.MaskedSecret) {
		t.Fatalf("expected [SECRET] in output, got %q", output)
	}
}

func TestRunCommandFromJSONFile(t *testing.T) {
	resetRunFlags()
	var gotHeader, gotBody, gotURL, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		gotMethod = r.Method
		gotURL = r.URL.String()
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		gotBody = string(buf)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "req.json")
	content := `{
		"name": "create",
		"variables": {"token": "abc123"},
		"request": {
			"method": "POST",
			"url": "` + srv.URL + `/items?page=1",
			"headers": [{"key": "Authorization", "value": "Bearer {{token}}"}],
			"body": "{\"name\":\"reqly\"}"
		}
	}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if gotMethod != "POST" {
		t.Fatalf("method: got %q", gotMethod)
	}
	if gotHeader != "Bearer abc123" {
		t.Fatalf("authorization (interpolated): got %q", gotHeader)
	}
	if gotBody != `{"name":"reqly"}` {
		t.Fatalf("body: got %q", gotBody)
	}
	if gotURL != "/items?page=1" {
		t.Fatalf("url: got %q", gotURL)
	}
}

func TestRunCommandFromYAMLFile(t *testing.T) {
	resetRunFlags()
	var gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		gotBody = string(buf)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "req.yaml")
	content := "name: ping\nrequest:\n  method: PUT\n  url: " + srv.URL + "/ping\n  body: '{\"p\":1}'\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if gotMethod != "PUT" || gotBody != `{"p":1}` {
		t.Fatalf("got method %q body %q", gotMethod, gotBody)
	}
}

func TestRunCommandFileWithFlagOverrides(t *testing.T) {
	resetRunFlags()
	var gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		gotBody = string(buf)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "req.json")
	content := `{"request": {"method": "GET", "url": "` + srv.URL + `"}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"run", "-m", "DELETE", "-d", `{"x":1}`, path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if gotMethod != "DELETE" || gotBody != `{"x":1}` {
		t.Fatalf("overrides not applied: method %q body %q", gotMethod, gotBody)
	}
}

func TestRunCommandFileMissing(t *testing.T) {
	resetRunFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", filepath.Join(t.TempDir(), "nope.json")})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "read request file") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestParseHeaders(t *testing.T) {
	headers, err := parseHeaders([]string{"Content-Type: application/json", " X-A : b "})
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(headers))
	}
	if headers[0].Key != "Content-Type" || headers[0].Value != "application/json" {
		t.Fatalf("unexpected header %+v", headers[0])
	}
	if headers[1].Key != "X-A" || headers[1].Value != "b" {
		t.Fatalf("unexpected header %+v", headers[1])
	}
}

func TestParseHeadersInvalid(t *testing.T) {
	if _, err := parseHeaders([]string{"NoColon"}); err == nil {
		t.Fatal("expected error for header without colon")
	}
}

// resetRunFlags clears flag state that persists across tests on the shared command.
func resetRunFlags() {
	runMethod = "GET"
	runHeaders = nil
	runBody = ""
	runTimeout = 30 * time.Second
	runRetries = 0
	runRetryDelay = 0
	envFlag = ""
	for _, name := range []string{"method", "header", "data", "timeout", "env", "retries", "retry-delay"} {
		if flag := runCmd.Flags().Lookup(name); flag != nil {
			flag.Changed = false
		}
	}
}

func TestRunCommandResolvesEnvironmentViaREQLYEnv(t *testing.T) {
	resetRunFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer dev-secret" {
			t.Errorf("expected environment-interpolated header, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dir := t.TempDir()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, "dev.yaml"), []byte(`
variables:
  API_URL: `+srv.URL+`
secrets:
  API_KEY: dev-secret
`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("REQLY_ENV", "dev")
	t.Chdir(dir)

	requestPath := filepath.Join(dir, "req.yaml")
	if err := os.WriteFile(requestPath, []byte(`
request:
  method: GET
  url: `+srv.URL+`
  headers:
    - key: Authorization
      value: Bearer {{API_KEY}}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"run", requestPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommandErrorsOnMissingSelectedEnvironment(t *testing.T) {
	resetRunFlags()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "environments"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REQLY_ENV", "staging")
	t.Chdir(dir)

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"run", "https://example.com"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for selected-but-missing environment")
	}
}

func TestRunCommandUsesFileEnvironmentField(t *testing.T) {
	resetRunFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Env"); got != "from-file-env" {
			t.Errorf("expected file-selected environment, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dir := t.TempDir()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, "prod.yaml"), []byte(`
variables:
  REGION: from-file-env
`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)
	requestPath := filepath.Join(dir, "req.yaml")
	if err := os.WriteFile(requestPath, []byte(`
environment: prod
request:
  method: GET
  url: `+srv.URL+`
  headers:
    - key: X-Env
      value: "{{REGION}}"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"run", requestPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommandEnvFlagOverridesFileEnvironment(t *testing.T) {
	resetRunFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Env"); got != "flag-wins" {
			t.Errorf("expected --env to override file field, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dir := t.TempDir()
	envDir := filepath.Join(dir, "environments")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, "prod.yaml"), []byte("variables:\n  REGION: file-field\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, "staging.yaml"), []byte("variables:\n  REGION: flag-wins\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)
	requestPath := filepath.Join(dir, "req.yaml")
	if err := os.WriteFile(requestPath, []byte(`
environment: prod
request:
  method: GET
  url: `+srv.URL+`
  headers:
    - key: X-Env
      value: "{{REGION}}"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"run", "--env", "staging", requestPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	envFlag = ""
}

func TestRunCommandMasksSecretsInResponseBody(t *testing.T) {
	resetRunFlags()
	secret := "super-secret-token-value"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"api_key": %q}`, secret)
	}))
	defer srv.Close()

	dir := t.TempDir()
	writeEnv(t, dir, "dev", "secrets:\n  API_KEY: "+secret+"\n")
	t.Chdir(dir)
	requestPath := filepath.Join(dir, "req.yaml")
	if err := os.WriteFile(requestPath, []byte(`
environment: dev
request:
  method: GET
  url: `+srv.URL+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", requestPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), secret) {
		t.Fatalf("secret leaked in run output:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "[SECRET]") {
		t.Fatalf("expected [SECRET] masking:\n%s", out.String())
	}
}

func TestRunCommandWithoutEnvironmentLeavesBodyUnmasked(t *testing.T) {
	resetRunFlags()
	body := "no-secret-here"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"value": %q}`, body)
	}))
	defer srv.Close()

	dir := t.TempDir()
	t.Chdir(dir)
	requestPath := filepath.Join(dir, "req.yaml")
	if err := os.WriteFile(requestPath, []byte(`
request:
  method: GET
  url: `+srv.URL+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"run", requestPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), body) {
		t.Fatalf("body should be printed unchanged when no environment is active:\n%s", out.String())
	}
}
