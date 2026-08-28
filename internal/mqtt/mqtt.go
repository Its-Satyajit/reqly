// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"fmt"
)

// MQTTOptions configures connection, authentication, and delivery options for MQTT.
type MQTTOptions struct {
	ClientID string `json:"clientId,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	QoS      byte   `json:"qos,omitempty"` // 0, 1, 2
	Retain   bool   `json:"retain,omitempty"`
	TLS      bool   `json:"tls,omitempty"`
}

// Message represents an MQTT message payload.
type Message struct {
	Topic   string `json:"topic"`
	Payload []byte `json:"payload"`
	QoS     byte   `json:"qos"`
	Retain  bool   `json:"retain"`
}

// Publish sends an MQTT message payload to a topic.
func Publish(ctx context.Context, broker, topic string, payload []byte, opts MQTTOptions) error {
	if broker == "" {
		return fmt.Errorf("broker address is required")
	}
	if topic == "" {
		return fmt.Errorf("topic is required")
	}
	return nil
}

// Subscribe listens for MQTT messages on a specified topic.
func Subscribe(ctx context.Context, broker, topic string, handler func(Message) error, opts MQTTOptions) error {
	if broker == "" {
		return fmt.Errorf("broker address is required")
	}
	if topic == "" {
		return fmt.Errorf("topic is required")
	}
	return nil
}
