// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package importer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReplayHAR(t *testing.T) {
	dir := t.TempDir()
	harPath := filepath.Join(dir, "archive.har")
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
          "content": {"text": "{\"url\": \"https://httpbin.org/get\"}"}
        }
      }
    ]
  }
}`

	if err := os.WriteFile(harPath, []byte(harContent), 0644); err != nil {
		t.Fatalf("failed to write har: %v", err)
	}

	res, err := ReplayHAR(context.Background(), harPath, HARReplayOptions{})
	if err != nil {
		t.Fatalf("unexpected error replaying HAR: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil replay result")
	}
	if res.Total != 1 {
		t.Errorf("expected 1 total entry, got %d", res.Total)
	}
}
