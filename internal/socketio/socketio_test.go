// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package socketio

import (
	"context"
	"testing"
)

func TestSocketIO_Validation(t *testing.T) {
	handler := func(ev Event) error { return nil }
	if err := Connect(context.Background(), "", handler, Options{}); err == nil {
		t.Errorf("expected error for empty URL")
	}
	if err := Connect(context.Background(), "http://localhost:3000", handler, Options{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := Emit(context.Background(), "", "ping", nil, Options{}); err == nil {
		t.Errorf("expected error for empty URL")
	}
	if err := Emit(context.Background(), "http://localhost:3000", "", nil, Options{}); err == nil {
		t.Errorf("expected error for empty event")
	}
	if err := Emit(context.Background(), "http://localhost:3000", "ping", "hello", Options{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
