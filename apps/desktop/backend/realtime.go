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
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Its-Satyajit/reqly/internal/sse"
	"github.com/Its-Satyajit/reqly/internal/websocket"
)

// realtimeEventName is the Wails event channel for one realtime session;
// every frame, status change, and error streams over it.
func realtimeEventName(sessionID string) string {
	return "reqly.realtime." + sessionID
}

// RealtimeHeader is one connection header (key/value pair from the dialog).
type RealtimeHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RealtimeOpenRequest starts a WebSocket ("ws") or Server-Sent Events
// ("sse") session identified by SessionID — a frontend-generated id so the
// webview can route streamed events to the right tab.
type RealtimeOpenRequest struct {
	SessionID string           `json:"sessionId"`
	Kind      string           `json:"kind"` // "ws" | "sse"
	URL       string           `json:"url"`
	Headers   []RealtimeHeader `json:"headers,omitempty"`
}

// RealtimeFrame is one streamed payload delivered to the frontend. Binary
// frames carry base64 in Data with Encoding "base64"; everything else is
// UTF-8 text with Encoding "".
type RealtimeFrame struct {
	SessionID string `json:"sessionId"`
	Type      string `json:"type"` // "message" | "status" | "error" | "closed"
	Direction string `json:"direction,omitempty"`
	Data      string `json:"data,omitempty"`
	Encoding  string `json:"encoding,omitempty"`
	Name      string `json:"name,omitempty"`    // SSE event name
	ID        string `json:"id,omitempty"`      // SSE event id
	RetryMs   int64  `json:"retryMs,omitempty"` // SSE retry hint (G-7.2.3)
	Timestamp int64  `json:"timestamp"`
}

// realtimeSession owns one live client and its read loop.
type realtimeSession struct {
	kind      string
	cancel    context.CancelFunc
	wsClient  *websocket.Client
	sseClient *sse.Client
	mu        sync.Mutex
	closed    bool
}

// realtimeSessions tracks live sessions; the mutex guards the map only —
// each session's client is internally synchronized.
var (
	realtimeMu       sync.Mutex
	realtimeSessions = make(map[string]*realtimeSession)
)

// emitFrame is indirected so tests can capture frames without Wails.
var emitRealtimeFrame = func(frame *RealtimeFrame) {
 	emitEvent(realtimeEventName(frame.SessionID), frame)
}

func (s *realtimeSession) emit(f *RealtimeFrame) {
	f.Timestamp = time.Now().UnixMilli()
	emitRealtimeFrame(f)
}

func headersFrom(list []RealtimeHeader) http.Header {
	h := http.Header{}
	for _, kv := range list {
		if strings.TrimSpace(kv.Key) != "" {
			h.Add(kv.Key, kv.Value)
		}
	}
	return h
}

// RealtimeClose tears down one session by id. Closing an unknown or already
// closed session is a no-op.
func (s *AppService) RealtimeClose(sessionID string) error {
	realtimeMu.Lock()
	sess, ok := realtimeSessions[sessionID]
	if ok {
		delete(realtimeSessions, sessionID)
	}
	realtimeMu.Unlock()
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
	if sess.kind == "ws" && sess.wsClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = sess.wsClient.Close(ctx)
	}
	if sess.kind == "sse" && sess.sseClient != nil {
		_ = sess.sseClient.Close()
	}
	sess.emit(&RealtimeFrame{SessionID: sessionID, Type: "closed"})
	return nil
}

// RealtimeSend writes one text frame on an open WebSocket session.
func (s *AppService) RealtimeSend(sessionID string, data string) error {
	realtimeMu.Lock()
	sess := realtimeSessions[sessionID]
	realtimeMu.Unlock()
	if sess == nil || sess.kind != "ws" {
		return fmt.Errorf("no open websocket session %q", sessionID)
	}
	if err := sess.wsClient.SendText(context.Background(), data); err != nil {
		return fmt.Errorf("send failed: %w", err)
	}
	sess.emit(&RealtimeFrame{
		SessionID: sessionID,
		Type:      "message",
		Direction: "out",
		Data:      data,
	})
	return nil
}

// RealtimeSendBinary writes one binary frame on an open WebSocket session.
// The payload arrives base64-encoded from the webview and is echoed back into
// the inspector the same way inbound binary frames are rendered.
func (s *AppService) RealtimeSendBinary(sessionID string, data string) error {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("payload is not valid base64: %w", err)
	}
	realtimeMu.Lock()
	sess := realtimeSessions[sessionID]
	realtimeMu.Unlock()
	if sess == nil || sess.kind != "ws" {
		return fmt.Errorf("no open websocket session %q", sessionID)
	}
	if err := sess.wsClient.SendBinary(context.Background(), raw); err != nil {
		return fmt.Errorf("send failed: %w", err)
	}
	sess.emit(&RealtimeFrame{
		SessionID: sessionID,
		Type:      "message",
		Direction: "out",
		Data:      data,
		Encoding:  "base64",
	})
	return nil
}

// RealtimeOpen dials the endpoint and starts streaming frames to
// `reqly.realtime.<sessionID>`. It returns after the handshake resolves:
// connect errors surface here, stream errors arrive as events.
func (s *AppService) RealtimeOpen(req RealtimeOpenRequest) error {
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return fmt.Errorf("sessionId is required")
	}
	url := strings.TrimSpace(req.URL)
	if url == "" {
		return fmt.Errorf("url is required")
	}

	if err := s.RealtimeClose(sessionID); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	headers := headersFrom(req.Headers)

	switch strings.ToLower(strings.TrimSpace(req.Kind)) {
	case "ws":
		client := websocket.NewClient(url, websocket.WithHeaders(headers))
		sess := &realtimeSession{kind: "ws", cancel: cancel, wsClient: client}
		realtimeMu.Lock()
		realtimeSessions[sessionID] = sess
		realtimeMu.Unlock()

		if err := client.Dial(ctx); err != nil {
			s.cleanupSession(sessionID, sess)
			cancel()
			return fmt.Errorf("dial failed: %w", err)
		}
		sess.emit(&RealtimeFrame{SessionID: sessionID, Type: "status", Data: string(websocket.StatusConnected)})
		go s.wsReadLoop(ctx, sessionID, sess, client)
		return nil

	case "sse":
		client := sse.NewClient(url, sse.WithHeaders(headers))
		sess := &realtimeSession{kind: "sse", cancel: cancel, sseClient: client}
		realtimeMu.Lock()
		realtimeSessions[sessionID] = sess
		realtimeMu.Unlock()

		if err := client.Connect(ctx); err != nil {
			s.cleanupSession(sessionID, sess)
			cancel()
			return fmt.Errorf("connect failed: %w", err)
		}
		sess.emit(&RealtimeFrame{SessionID: sessionID, Type: "status", Data: "connected"})
		go s.sseReadLoop(ctx, sessionID, sess, client)
		return nil

	default:
		cancel()
		return fmt.Errorf("unknown kind %q: pick ws or sse", req.Kind)
	}
}

func (s *AppService) cleanupSession(sessionID string, sess *realtimeSession) {
	realtimeMu.Lock()
	if realtimeSessions[sessionID] == sess {
		delete(realtimeSessions, sessionID)
	}
	realtimeMu.Unlock()
}

// wsReadLoop forwards incoming frames until the context is cancelled or the
// receive fails; either way the session is torn down and a closed event sent.
func (s *AppService) wsReadLoop(ctx context.Context, sessionID string, sess *realtimeSession, client *websocket.Client) {
	defer func() {
		s.cleanupSession(sessionID, sess)
		sess.emit(&RealtimeFrame{SessionID: sessionID, Type: "closed"})
	}()
	for {
		msg, err := client.Receive(ctx)
		if err != nil {
			if ctx.Err() == nil {
				sess.emit(&RealtimeFrame{SessionID: sessionID, Type: "error", Data: err.Error()})
			}
			return
		}
		frame := &RealtimeFrame{
			SessionID: sessionID,
			Type:      "message",
			Direction: "in",
		}
		if msg.Type == websocket.MessageBinary {
			frame.Encoding = "base64"
			frame.Data = base64.StdEncoding.EncodeToString(msg.Data)
		} else {
			frame.Data = string(msg.Data)
		}
		sess.emit(frame)
	}
}

// sseReadLoop mirrors wsReadLoop for Server-Sent Events.
func (s *AppService) sseReadLoop(ctx context.Context, sessionID string, sess *realtimeSession, client *sse.Client) {
	defer func() {
		s.cleanupSession(sessionID, sess)
		sess.emit(&RealtimeFrame{SessionID: sessionID, Type: "closed"})
	}()
	for {
		ev, err := client.Next(ctx)
		if err != nil {
			if ctx.Err() == nil {
				sess.emit(&RealtimeFrame{SessionID: sessionID, Type: "error", Data: err.Error()})
			}
			return
		}
		sess.emit(&RealtimeFrame{
			SessionID: sessionID,
			Type:      "message",
			Direction: "in",
			Name:      ev.Name,
			ID:        ev.ID,
			Data:      ev.Data,
			RetryMs:   ev.Retry.Milliseconds(),
		})
	}
}
