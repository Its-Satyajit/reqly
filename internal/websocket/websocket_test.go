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

package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// newEchoServer returns a WebSocket server that echoes text messages back.
func newEchoServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			typ, data, err := conn.Read(r.Context())
			if err != nil {
				return
			}
			if err := conn.Write(r.Context(), typ, data); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestDialSendReceive(t *testing.T) {
	srv := newEchoServer(t)
	wsURL := "ws" + srv.URL[len("http"):]

	var statuses []Status
	var mu sync.Mutex
	c := NewClient(wsURL, WithStatusHandler(func(s Status) {
		mu.Lock()
		defer mu.Unlock()
		statuses = append(statuses, s)
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Dial(ctx); err != nil {
		t.Fatal(err)
	}
	if c.Status() != StatusConnected {
		t.Fatalf("status: got %q", c.Status())
	}

	if err := c.SendText(ctx, "hello"); err != nil {
		t.Fatal(err)
	}
	msg, err := c.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != MessageText {
		t.Fatalf("type: got %q", msg.Type)
	}
	if string(msg.Data) != "hello" {
		t.Fatalf("echo: got %q", msg.Data)
	}
	if msg.Sent {
		t.Fatal("received message should not be marked sent")
	}

	if err := c.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if c.Status() != StatusDisconnected {
		t.Fatalf("after close status: got %q", c.Status())
	}

	mu.Lock()
	defer mu.Unlock()
	if len(statuses) < 2 {
		t.Fatalf("expected status transitions, got %v", statuses)
	}
}

func TestReceiveAfterCloseErrors(t *testing.T) {
	srv := newEchoServer(t)
	wsURL := "ws" + srv.URL[len("http"):]

	c := NewClient(wsURL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Dial(ctx); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(ctx); err != nil {
		t.Fatal(err)
	}

	// The server also closes; Receive should error out.
	if _, err := c.Receive(ctx); err == nil {
		t.Fatal("expected error receiving after close")
	}
}

func TestSendWithoutConnect(t *testing.T) {
	c := NewClient("ws://example.com")
	if err := c.SendText(context.Background(), "x"); err == nil {
		t.Fatal("expected error sending before dial")
	}
}

func TestDialError(t *testing.T) {
	c := NewClient("ws://127.0.0.1:1/ws")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Dial(ctx); err == nil {
		t.Fatal("expected dial error for unreachable host")
	}
	if c.Status() != StatusDisconnected {
		t.Fatalf("status after failed dial: got %q", c.Status())
	}
}

func TestSubprotocolNegotiated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{"graphql-ws"}})
		if err != nil {
			return
		}
		conn.Close(websocket.StatusNormalClosure, "")
	}))
	t.Cleanup(srv.Close)

	c := NewClient("ws"+srv.URL[len("http"):], WithSubprotocols([]string{"graphql-ws"}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Dial(ctx); err != nil {
		t.Fatal(err)
	}
	if got := c.Subprotocol(); got != "graphql-ws" {
		t.Fatalf("subprotocol: got %q", got)
	}
}
