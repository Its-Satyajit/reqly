// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
