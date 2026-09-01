// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"net"
	"testing"
)

func TestMQTT_PublishValidation(t *testing.T) {
	if err := Publish(context.Background(), "", "test/topic", []byte("hello"), MQTTOptions{}); err == nil {
		t.Errorf("expected error for empty broker")
	}
	if err := Publish(context.Background(), "tcp://localhost:1883", "", []byte("hello"), MQTTOptions{}); err == nil {
		t.Errorf("expected error for empty topic")
	}
	// Publish dials the broker; start a dummy listener so dial succeeds.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	broker := ln.Addr().String()
	if err := Publish(context.Background(), broker, "test/topic", []byte("hello"), MQTTOptions{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Also test tcp:// scheme form.
	if err := Publish(context.Background(), "tcp://"+broker, "test/topic", []byte("hello"), MQTTOptions{}); err != nil {
		t.Errorf("unexpected error with scheme: %v", err)
	}
}

func TestMQTT_PublishConnectionRefused(t *testing.T) {
	// No broker listening on this port -> should return dial error (R2).
	if err := Publish(context.Background(), "127.0.0.1:1", "test/topic", []byte("hello"), MQTTOptions{}); err == nil {
		t.Errorf("expected dial error for refused connection")
	}
}

func TestMQTT_SubscribeValidation(t *testing.T) {
	handler := func(msg Message) error { return nil }
	if err := Subscribe(context.Background(), "", "test/topic", handler, MQTTOptions{}); err == nil {
		t.Errorf("expected error for empty broker")
	}
	if err := Subscribe(context.Background(), "tcp://localhost:1883", "", handler, MQTTOptions{}); err == nil {
		t.Errorf("expected error for empty topic")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	broker := ln.Addr().String()
	if err := Subscribe(context.Background(), broker, "test/topic", handler, MQTTOptions{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMQTT_SubscribeConnectionRefused(t *testing.T) {
	handler := func(msg Message) error { return nil }
	if err := Subscribe(context.Background(), "127.0.0.1:1", "test/topic", handler, MQTTOptions{}); err == nil {
		t.Errorf("expected dial error for refused connection")
	}
}
