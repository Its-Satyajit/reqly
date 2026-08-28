// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"testing"
)

func TestMQTT_PublishValidation(t *testing.T) {
	if err := Publish(context.Background(), "", "test/topic", []byte("hello"), MQTTOptions{}); err == nil {
		t.Errorf("expected error for empty broker")
	}
	if err := Publish(context.Background(), "tcp://localhost:1883", "", []byte("hello"), MQTTOptions{}); err == nil {
		t.Errorf("expected error for empty topic")
	}
	if err := Publish(context.Background(), "tcp://localhost:1883", "test/topic", []byte("hello"), MQTTOptions{}); err != nil {
		t.Errorf("unexpected error: %v", err)
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
	if err := Subscribe(context.Background(), "tcp://localhost:1883", "test/topic", handler, MQTTOptions{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
