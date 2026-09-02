// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Its-Satyajit/reqly/internal/mqtt"
)

func mqttEventName(sessionID string) string {
	return "reqly.mqtt." + sessionID
}

// MqttPublishRequest mirrors CLI `mqtt pub <broker> --topic --message --qos --retain --username --password`.
type MqttPublishRequest struct {
	Broker   string `json:"broker"`
	Topic    string `json:"topic"`
	Message  string `json:"message"`
	QoS      byte   `json:"qos"`
	Retain   bool   `json:"retain"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	ClientID string `json:"clientId,omitempty"`
}

// MqttSubscribeRequest mirrors CLI `mqtt sub <broker> --topic --qos --username --password` with a tab-owned SessionID.
type MqttSubscribeRequest struct {
	SessionID string `json:"sessionId"`
	Broker    string `json:"broker"`
	Topic     string `json:"topic"`
	QoS       byte   `json:"qos"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	ClientID  string `json:"clientId,omitempty"`
}

// MqttFrame streams over `reqly.mqtt.<sessionID>`.
type MqttFrame struct {
	SessionID string `json:"sessionId"`
	Type      string `json:"type"` // "message" | "status" | "error" | "closed"
	Topic     string `json:"topic,omitempty"`
	Payload   string `json:"payload,omitempty"`
	QoS       byte   `json:"qos,omitempty"`
	Retain    bool   `json:"retain,omitempty"`
	Data      string `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type mqttSession struct {
	cancel context.CancelFunc
	mu     sync.Mutex
	closed bool
}

var (
	mqttMu       sync.Mutex
	mqttSessions = make(map[string]*mqttSession)
)

var emitMqttFrame = func(frame *MqttFrame) {
	emitEvent(mqttEventName(frame.SessionID), frame)
}

func (s *mqttSession) emit(f *MqttFrame) {
	f.Timestamp = time.Now().UnixMilli()
	emitMqttFrame(f)
}

// MqttPublish sends one message to the broker. It dials to verify reachability via internal/mqtt.
func (s *AppService) MqttPublish(req MqttPublishRequest) error {
	broker := strings.TrimSpace(req.Broker)
	topic := strings.TrimSpace(req.Topic)
	if broker == "" {
		return fmt.Errorf("broker is required")
	}
	if topic == "" {
		return fmt.Errorf("topic is required")
	}
	opts := mqtt.MQTTOptions{
		ClientID: req.ClientID,
		Username: req.Username,
		Password: req.Password,
		QoS:      req.QoS,
		Retain:   req.Retain,
	}
	if err := mqtt.Publish(context.Background(), broker, topic, []byte(req.Message), opts); err != nil {
		return err
	}
	return nil
}

// MqttSubscribe opens a subscription streaming messages to `reqly.mqtt.<sessionID>`.
func (s *AppService) MqttSubscribe(req MqttSubscribeRequest) error {
	sessionID := strings.TrimSpace(req.SessionID)
	broker := strings.TrimSpace(req.Broker)
	topic := strings.TrimSpace(req.Topic)
	if sessionID == "" {
		return fmt.Errorf("sessionId is required")
	}
	if broker == "" {
		return fmt.Errorf("broker is required")
	}
	if topic == "" {
		return fmt.Errorf("topic is required")
	}
	if err := s.MqttCancel(sessionID); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	sess := &mqttSession{cancel: cancel}
	mqttMu.Lock()
	mqttSessions[sessionID] = sess
	mqttMu.Unlock()

	opts := mqtt.MQTTOptions{
		ClientID: req.ClientID,
		Username: req.Username,
		Password: req.Password,
		QoS:      req.QoS,
	}

	sess.emit(&MqttFrame{SessionID: sessionID, Type: "status", Data: "subscribing"})
	go func() {
		defer func() {
			mqttMu.Lock()
			if mqttSessions[sessionID] == sess {
				delete(mqttSessions, sessionID)
			}
			mqttMu.Unlock()
			sess.mu.Lock()
			alreadyClosed := sess.closed
			sess.closed = true
			sess.mu.Unlock()
			if !alreadyClosed {
				sess.emit(&MqttFrame{SessionID: sessionID, Type: "closed"})
			}
		}()
		handler := func(msg mqtt.Message) error {
			sess.emit(&MqttFrame{
				SessionID: sessionID,
				Type:      "message",
				Topic:     msg.Topic,
				Payload:   string(msg.Payload),
				QoS:       msg.QoS,
				Retain:    msg.Retain,
			})
			return nil
		}
		if err := mqtt.Subscribe(ctx, broker, topic, handler, opts); err != nil {
			if ctx.Err() == nil {
				sess.emit(&MqttFrame{SessionID: sessionID, Type: "error", Data: err.Error()})
			} else {
				sess.emit(&MqttFrame{SessionID: sessionID, Type: "status", Data: "cancelled"})
			}
			return
		}
		sess.emit(&MqttFrame{SessionID: sessionID, Type: "status", Data: "subscribed"})
		<-ctx.Done()
	}()
	return nil
}

// MqttCancel tears down one subscription by sessionID.
func (s *AppService) MqttCancel(sessionID string) error {
	mqttMu.Lock()
	sess, ok := mqttSessions[sessionID]
	if ok {
		delete(mqttSessions, sessionID)
	}
	mqttMu.Unlock()
	if !ok {
		return nil
	}
	sess.mu.Lock()
	if sess.closed {
		sess.mu.Unlock()
		return nil
	}
	sess.closed = true
	sess.mu.Unlock()
	sess.cancel()
	return nil
}
