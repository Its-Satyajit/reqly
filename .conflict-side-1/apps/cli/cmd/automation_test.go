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
)

func TestAutomationCLIOnce(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	yamlStr := fmt.Sprintf(`
name: test-auto
workflow:
  name: wf
  steps:
    - id: s1
      name: S1
      request:
        method: GET
        url: %s
interval: "0"
`, srv.URL)

	dir := t.TempDir()
	f := filepath.Join(dir, "auto.yaml")
	if err := os.WriteFile(f, []byte(yamlStr), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"automation", "run", f, "--once"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("automation run: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "test-auto") || !strings.Contains(out, "PASSED") {
		t.Fatalf("unexpected output: %s", out)
	}
}
