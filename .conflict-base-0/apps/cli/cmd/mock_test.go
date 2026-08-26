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
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/mocking"
	"github.com/Its-Satyajit/reqly/internal/openapi"
)

// syncBuffer is a concurrency-safe bytes.Buffer for capturing command output
// written from a background goroutine.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

const mockSpec = `openapi: 3.0.3
info:
  title: Mock Test API
  version: 1.0.0
paths:
  /ping:
    get:
      operationId: ping
      responses:
        "200":
          description: pong
          content:
            application/json:
              example: {message: pong}
`

func writeMockSpec(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(path, []byte(mockSpec), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func TestMockServe(t *testing.T) {
	spec := writeMockSpec(t)
	doc, err := openapi.LoadFile(spec)
	if err != nil {
		t.Fatalf("openapi.LoadFile() error = %v", err)
	}
	handler, err := mocking.NewServer(doc)
	if err != nil {
		t.Fatalf("mocking.NewServer() error = %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()

	var out syncBuffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serveMock(ctx, rootCmd, handler, listener) }()

	url := "http://" + listener.Addr().String()
	resp, err := http.Get(url + "/ping")
	if err != nil {
		cancel()
		t.Fatalf("http.Get() error = %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		cancel()
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var parsed map[string]string
	if err := json.Unmarshal(body, &parsed); err != nil {
		cancel()
		t.Fatalf("body not JSON: %v (%s)", err, body)
	}
	if parsed["message"] != "pong" {
		cancel()
		t.Fatalf("message = %q, want pong", parsed["message"])
	}
	if !strings.Contains(out.String(), "mock server listening") {
		t.Fatalf("startup line missing from output: %q", out.String())
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serveMock error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serveMock did not stop after cancel")
	}
}

func TestMockCommandLoadError(t *testing.T) {
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"mock", "/nonexistent/spec.yaml", "--port", "0"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for missing spec file")
	}
}
