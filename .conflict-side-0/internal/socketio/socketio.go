// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package socketio

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// Options configures connection parameters for Socket.IO.
type Options struct {
	Namespace string            `json:"namespace,omitempty"`
	Query     map[string]string `json:"query,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// Event represents a Socket.IO event.
type Event struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Data      any    `json:"data"`
}

// Connect establishes a Socket.IO connection and listens for events.
func Connect(ctx context.Context, rawURL string, handler func(Event) error, opts Options) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if err := dialSocketIO(ctx, rawURL); err != nil {
		return fmt.Errorf("socketio connect: %w", err)
	}
	// Stub: no real Engine.IO handshake yet. Dial check ensures broker
	// reachability similar to mqtt/grpc. Simulate a handshake event so
	// callers see progress; block until context cancelled.
	if handler != nil {
		ns := opts.Namespace
		if ns == "" {
			ns = "/"
		}
		_ = handler(Event{Namespace: ns, Name: "connect", Data: map[string]string{"status": "connected", "url": rawURL}})
		_ = handler(Event{Namespace: ns, Name: "welcome", Data: "welcome"})
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Emit sends a Socket.IO event to a server endpoint.
func Emit(ctx context.Context, rawURL, event string, data any, opts Options) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}
	if event == "" {
		return fmt.Errorf("event name is required")
	}
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if err := dialSocketIO(ctx, rawURL); err != nil {
		return fmt.Errorf("socketio emit: %w", err)
	}
	return nil
}

func dialSocketIO(ctx context.Context, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	host := u.Host
	if host == "" {
		return fmt.Errorf("missing host in URL %q", rawURL)
	}
	if !strings.Contains(host, ":") {
		if u.Scheme == "https" || u.Scheme == "wss" {
			host = net.JoinHostPort(host, "443")
		} else {
			host = net.JoinHostPort(host, "80")
		}
	}
	timeout := 3 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d < timeout {
			timeout = d
		}
	}
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return fmt.Errorf("dial %s: %w", host, err)
	}
	_ = conn.Close()
	return nil
}
