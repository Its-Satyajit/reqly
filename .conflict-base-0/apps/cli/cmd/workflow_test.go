package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowCLICommand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token": "xyz"}`))
	}))
	defer srv.Close()

	wfYaml := `
name: Test Flow
steps:
  - id: step1
    name: Step 1
    request:
      method: GET
      url: ` + srv.URL + `
    extract:
      token: token
`
	dir := t.TempDir()
	wfFile := filepath.Join(dir, "flow.yaml")
	if err := os.WriteFile(wfFile, []byte(wfYaml), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"workflow", wfFile})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("workflow CLI failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Workflow: Test Flow") || !strings.Contains(out, "token = xyz") {
		t.Fatalf("unexpected workflow CLI output: %s", out)
	}
}

func TestWorkflowCLICommand_VerboseFailureDiagnostics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "database connection failure"}`))
	}))
	defer srv.Close()

	wfYaml := `
name: Failing Flow
steps:
  - id: failing_step
    name: Failing Step
    request:
      method: GET
      url: ` + srv.URL + `
`
	dir := t.TempDir()
	wfFile := filepath.Join(dir, "fail.yaml")
	if err := os.WriteFile(wfFile, []byte(wfYaml), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"workflow", wfFile})

	_ = rootCmd.Execute()
	out := buf.String()
	if !strings.Contains(out, "[FAILED]") || !strings.Contains(out, "500") || !strings.Contains(out, "database connection failure") {
		t.Fatalf("expected detailed failure diagnostics, got:\n%s", out)
	}
}
