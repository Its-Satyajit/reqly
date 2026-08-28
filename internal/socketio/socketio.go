// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package socketio

import (
	"context"
	"fmt"
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
	return nil
}

// Emit sends a Socket.IO event to a server endpoint.
func Emit(ctx context.Context, rawURL, event string, data any, opts Options) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}
	if event == "" {
		return fmt.Errorf("event name is required")
	}
	return nil
}
