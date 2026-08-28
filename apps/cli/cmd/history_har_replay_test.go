// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryHARReplayCmd(t *testing.T) {
	dir := t.TempDir()
	harPath := filepath.Join(dir, "network.har")
	harContent := `{
  "log": {
    "version": "1.2",
    "entries": [
      {
        "request": {
          "method": "GET",
          "url": "https://httpbin.org/get",
          "headers": []
        },
        "response": {
          "status": 200,
          "content": {"text": "{}"}
        }
      }
    ]
  }
}`
	if err := os.WriteFile(harPath, []byte(harContent), 0644); err != nil {
		t.Fatalf("failed to write har: %v", err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"history", "replay", "--har", harPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error running reqly history replay --har: %v", err)
	}
}
