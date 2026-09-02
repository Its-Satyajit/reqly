// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package socketio

import (
	"context"
	"net"
	"net/http"
	"testing"
)

func TestSocketIO_Validation(t *testing.T) {
	handler := func(ev Event) error { return nil }
	if err := Connect(context.Background(), "", handler, Options{}); err == nil {
		t.Errorf("expected error for empty URL")
	}
	// Connect dials the host; start a dummy HTTP server so dial succeeds.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})}
	go srv.Serve(ln)
	defer srv.Close()
	url := "http://" + ln.Addr().String()
	if err := Connect(context.Background(), url, handler, Options{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := Emit(context.Background(), "", "ping", nil, Options{}); err == nil {
		t.Errorf("expected error for empty URL")
	}
	if err := Emit(context.Background(), url, "", nil, Options{}); err == nil {
		t.Errorf("expected error for empty event")
	}
	if err := Emit(context.Background(), url, "ping", "hello", Options{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSocketIO_ConnectionRefused(t *testing.T) {
	handler := func(ev Event) error { return nil }
	// No server on port 1 -> dial should fail
	if err := Connect(context.Background(), "http://127.0.0.1:1", handler, Options{}); err == nil {
		t.Errorf("expected dial error for refused connection")
	}
	if err := Emit(context.Background(), "http://127.0.0.1:1", "ping", "hello", Options{}); err == nil {
		t.Errorf("expected dial error for refused emit")
	}
}

func TestSocketIO_EmitsHandshakeEvents(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})}
	go srv.Serve(ln)
	defer srv.Close()
	url := "http://" + ln.Addr().String()
	var events []Event
	handler := func(ev Event) error {
		events = append(events, ev)
		return nil
	}
	if err := Connect(context.Background(), url, handler, Options{}); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 handshake events (connect+welcome), got %d", len(events))
	}
	if events[0].Name != "connect" || events[1].Name != "welcome" {
		t.Fatalf("unexpected handshake events: %+v", events)
	}
}
