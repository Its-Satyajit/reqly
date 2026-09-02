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

package sse

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newStreamServer returns a server that writes the given raw SSE payload.
func newStreamServer(t *testing.T, payload string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("response writer does not support flushing")
			return
		}
		fmt.Fprint(w, payload)
		flusher.Flush()
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSSENextParsesEvent(t *testing.T) {
	payload := "data: hello world\n\n"
	srv := newStreamServer(t, payload)

	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	if c.Status() != StatusConnected {
		t.Fatalf("status: got %q", c.Status())
	}

	ev, err := c.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Data != "hello world" {
		t.Fatalf("data: got %q", ev.Data)
	}
	if ev.Name != "" {
		t.Fatalf("name: got %q", ev.Name)
	}
}

func TestSSENextNamedEventWithFields(t *testing.T) {
	payload := "id: 42\nevent: stock\ndata: {\"symbol\":\"AAPL\"}\n\n"
	srv := newStreamServer(t, payload)

	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	ev, err := c.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ev.ID != "42" {
		t.Fatalf("id: got %q", ev.ID)
	}
	if ev.Name != "stock" {
		t.Fatalf("name: got %q", ev.Name)
	}
	if ev.Data != `{"symbol":"AAPL"}` {
		t.Fatalf("data: got %q", ev.Data)
	}
}

func TestSSEMultiLineData(t *testing.T) {
	payload := "data: line1\ndata: line2\n\n"
	srv := newStreamServer(t, payload)

	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	ev, err := c.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Data != "line1\nline2" {
		t.Fatalf("data: got %q", ev.Data)
	}
}

func TestSSESkipsCommentsAndBlankFields(t *testing.T) {
	payload := ": heartbeat\n\nid: 1\n: comment\ndata: value\n\n"
	srv := newStreamServer(t, payload)

	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	ev, err := c.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ev.ID != "1" || ev.Data != "value" {
		t.Fatalf("event: got %+v", ev)
	}
}

func TestSSERetryField(t *testing.T) {
	payload := "retry: 5000\ndata: x\n\n"
	srv := newStreamServer(t, payload)

	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	ev, err := c.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Retry != 5*time.Second {
		t.Fatalf("retry: got %v", ev.Retry)
	}
}

func TestSSENotEventStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{}")
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err == nil {
		t.Fatal("expected error for non-event-stream response")
	}
	if c.Status() != StatusDisconnected {
		t.Fatalf("status: got %q", c.Status())
	}
}

func TestSSEBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestSSEConnectError(t *testing.T) {
	c := NewClient("http://127.0.0.1:1/stream")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Connect(ctx); err == nil {
		t.Fatal("expected connect error")
	}
}

func TestSSENextBeforeConnect(t *testing.T) {
	c := NewClient("http://example.com/stream")
	if _, err := c.Next(context.Background()); err == nil {
		t.Fatal("expected error calling Next before Connect")
	}
}

func TestSSEStreamEnds(t *testing.T) {
	srv := newStreamServer(t, "data: one\n\ndata: two\n\n")
	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	ev, err := c.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ev.Data, "one") {
		t.Fatalf("data: got %q", ev.Data)
	}
	// Stream has ended; the next read should hit EOF (an error).
	if ev, err := c.Next(ctx); err != nil {
		t.Fatal(err)
	} else if !strings.Contains(ev.Data, "two") {
		t.Fatalf("data: got %q", ev.Data)
	}
	if _, err := c.Next(ctx); err == nil {
		t.Fatal("expected EOF error after stream ends")
	}
}
