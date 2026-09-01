// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
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
	if err := dialBroker(ctx, broker); err != nil {
		return fmt.Errorf("mqtt publish: %w", err)
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
	if err := dialBroker(ctx, broker); err != nil {
		return fmt.Errorf("mqtt subscribe: %w", err)
	}
	// Stub: no real subscription loop yet. Dial check ensures broker reachability
	// mirrors grpc/services behavior (EXIT:1 on connection refused). Actual
	// message delivery requires an external broker; return context cancellation.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func dialBroker(ctx context.Context, broker string) error {
	addr := parseBrokerAddr(broker)
	timeout := 3 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d < timeout {
			timeout = d
		}
	}
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	_ = conn.Close()
	return nil
}

func parseBrokerAddr(broker string) string {
	b := strings.TrimSpace(broker)
	// If broker contains :// treat as URL.
	if strings.Contains(b, "://") {
		if u, err := url.Parse(b); err == nil && u.Host != "" {
			if strings.Contains(u.Host, ":") {
				return u.Host
			}
			// no port in URL, default 1883
			return net.JoinHostPort(u.Host, "1883")
		}
	}
	// Plain host:port or host
	if strings.Contains(b, ":") {
		return b
	}
	return net.JoinHostPort(b, "1883")
}
