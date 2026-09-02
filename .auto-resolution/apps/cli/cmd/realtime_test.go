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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestWsCommandDialFailure(t *testing.T) {
	resetWsFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"ws", "ws://127.0.0.1:1/socket"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for unreachable ws host")
	}
}

func TestWsCommandConnectAndExchange(t *testing.T) {
	resetWsFlags()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "hello" {
			t.Errorf("missing X-Test header")
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")
		typ, data, err := conn.Read(r.Context())
		if err != nil {
			return
		}
		if err := conn.Write(r.Context(), typ, data); err != nil {
			return
		}
	}))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	// Simulate stdin: send one message then EOF.
	origStdin := stdinReader
	done := make(chan struct{})
	stdinReader = &testStdin{data: []byte("hello\n"), onEOF: func() {
		close(done)
	}}
	defer func() { stdinReader = origStdin }()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"ws", "-H", "X-Test: hello", "-t", "3s", wsURL})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{"connected to", "> hello", "[", "closed"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stdin was not drained")
	}
}

func TestSseCommandStreamsEvents(t *testing.T) {
	resetSseFlags()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("no flusher")
			return
		}
		w.Write([]byte("id: 1\nevent: update\ndata: first\n\ndata: second\n\n"))
		flusher.Flush()
	}))
	defer srv.Close()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"sse", "-c", "2", srv.URL})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{"streaming from", "event update (1)", "first", "second", "received 2 event(s)"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestSseCommandConnectFailure(t *testing.T) {
	resetSseFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"sse", "http://127.0.0.1:1/stream"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for unreachable sse host")
	}
}

// testStdin implements io.Reader to emulate piped stdin in tests.
type testStdin struct {
	data   []byte
	pos    int
	onEOF  func()
	called bool
}

func (t *testStdin) Read(p []byte) (int, error) {
	if t.pos >= len(t.data) {
		if !t.called && t.onEOF != nil {
			t.called = true
			t.onEOF()
		}
		return 0, io.EOF
	}
	n := copy(p, t.data[t.pos:])
	t.pos += n
	return n, nil
}

func resetWsFlags() {
	wsHeaderFlags = nil
	wsTimeout = 30 * time.Second
	for _, name := range []string{"header", "timeout"} {
		if flag := wsCmd.Flags().Lookup(name); flag != nil {
			flag.Changed = false
		}
	}
}

func resetSseFlags() {
	sseHeaderFlags = nil
	sseTimeout = 0
	sseCount = 0
	for _, name := range []string{"header", "timeout", "count"} {
		if flag := sseCmd.Flags().Lookup(name); flag != nil {
			flag.Changed = false
		}
	}
}
