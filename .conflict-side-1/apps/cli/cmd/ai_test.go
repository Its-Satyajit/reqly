package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAICLICommands(t *testing.T) {
	tempDir := t.TempDir()
	respFile := filepath.Join(tempDir, "response.json")
	reqFile := filepath.Join(tempDir, "request.json")

	_ = os.WriteFile(respFile, []byte(`{"statusCode": 200, "statusText": "OK", "body": "{\"id\": 1, \"name\": \"Reqly\"}"}`), 0600)
	_ = os.WriteFile(reqFile, []byte(`{"request": {"method": "GET", "url": "https://api.example.com/users"}}`), 0600)

	// 1. ai explain
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"ai", "explain", respFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("ai explain failed: %v", err)
	}
	if !strings.Contains(buf.String(), "response 200 OK") {
		t.Errorf("unexpected explain output: %s", buf.String())
	}

	// 2. ai test
	buf.Reset()
	rootCmd.SetArgs([]string{"ai", "test", respFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("ai test failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Status code is 200") {
		t.Errorf("unexpected test output: %s", buf.String())
	}

	// 3. ai docs
	buf.Reset()
	rootCmd.SetArgs([]string{"ai", "docs", reqFile, respFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("ai docs failed: %v", err)
	}
	if !strings.Contains(buf.String(), "GET") || !strings.Contains(buf.String(), "https://api.example.com/users") {
		t.Errorf("unexpected docs output: %s", buf.String())
	}

	// 4. ai diagnose
	buf.Reset()
	diagFile := filepath.Join(tempDir, "diag_401.json")
	_ = os.WriteFile(diagFile, []byte(`{"statusCode": 401, "statusText": "Unauthorized"}`), 0600)
	rootCmd.SetArgs([]string{"ai", "diagnose", diagFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("ai diagnose failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Unauthorized (401)") {
		t.Errorf("unexpected diagnose output: %s", buf.String())
	}
}
