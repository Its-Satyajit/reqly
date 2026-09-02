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

	"github.com/Its-Satyajit/reqly/internal/socketio"
)

func socketIOEventName(sessionID string) string {
	return "reqly.socketio." + sessionID
}

// SocketIOConnectRequest mirrors CLI `socketio connect <url> [--namespace]`.
type SocketIOConnectRequest struct {
	SessionID string `json:"sessionId"`
	URL       string `json:"url"`
	Namespace string `json:"namespace,omitempty"`
}

// SocketIOEmitRequest mirrors CLI `socketio emit <url> --event --data [--namespace]`.
type SocketIOEmitRequest struct {
	SessionID string `json:"sessionId"`
	URL       string `json:"url"`
	Event     string `json:"event"`
	Data      any    `json:"data,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// SocketIOFrame streams over `reqly.socketio.<sessionID>`.
type SocketIOFrame struct {
	SessionID string `json:"sessionId"`
	Type      string `json:"type"` // "message" | "status" | "error" | "closed"
	Namespace string `json:"namespace,omitempty"`
	Event     string `json:"event,omitempty"`
	Data      any    `json:"data,omitempty"`
	Raw       string `json:"raw,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type socketIOSession struct {
	cancel context.CancelFunc
	mu     sync.Mutex
	closed bool
}

var (
	socketIOMu       sync.Mutex
	socketIOSessions = make(map[string]*socketIOSession)
)

var emitSocketIOFrame = func(frame *SocketIOFrame) {
	emitEvent(socketIOEventName(frame.SessionID), frame)
}

func (s *socketIOSession) emit(f *SocketIOFrame) {
	f.Timestamp = time.Now().UnixMilli()
	emitSocketIOFrame(f)
}

// SocketIOConnect opens a Socket.IO session streaming events to `reqly.socketio.<sessionID>`.
func (s *AppService) SocketIOConnect(req SocketIOConnectRequest) error {
	sessionID := strings.TrimSpace(req.SessionID)
	rawURL := strings.TrimSpace(req.URL)
	if sessionID == "" {
		return fmt.Errorf("sessionId is required")
	}
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}
	if err := s.SocketIOClose(sessionID); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	sess := &socketIOSession{cancel: cancel}
	socketIOMu.Lock()
	socketIOSessions[sessionID] = sess
	socketIOMu.Unlock()

	opts := socketio.Options{Namespace: req.Namespace}
	sess.emit(&SocketIOFrame{SessionID: sessionID, Type: "status", Data: "connecting", Namespace: opts.Namespace})
	go func() {
		defer func() {
			socketIOMu.Lock()
			if socketIOSessions[sessionID] == sess {
				delete(socketIOSessions, sessionID)
			}
			socketIOMu.Unlock()
			sess.mu.Lock()
			alreadyClosed := sess.closed
			sess.closed = true
			sess.mu.Unlock()
			if !alreadyClosed {
				sess.emit(&SocketIOFrame{SessionID: sessionID, Type: "closed"})
			}
		}()
		handler := func(ev socketio.Event) error {
			sess.emit(&SocketIOFrame{
				SessionID: sessionID,
				Type:      "message",
				Namespace: ev.Namespace,
				Event:     ev.Name,
				Data:      ev.Data,
			})
			return nil
		}
		if err := socketio.Connect(ctx, rawURL, handler, opts); err != nil {
			if ctx.Err() == nil {
				sess.emit(&SocketIOFrame{SessionID: sessionID, Type: "error", Raw: err.Error()})
			} else {
				sess.emit(&SocketIOFrame{SessionID: sessionID, Type: "status", Data: "cancelled"})
			}
			return
		}
		sess.emit(&SocketIOFrame{SessionID: sessionID, Type: "status", Data: "connected"})
		<-ctx.Done()
	}()
	return nil
}

// SocketIOEmit sends one event to a Socket.IO endpoint (one-shot, no session required but sessionId is echoed).
func (s *AppService) SocketIOEmit(req SocketIOEmitRequest) error {
	rawURL := strings.TrimSpace(req.URL)
	event := strings.TrimSpace(req.Event)
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}
	if event == "" {
		return fmt.Errorf("event is required")
	}
	opts := socketio.Options{Namespace: req.Namespace}
	if err := socketio.Emit(context.Background(), rawURL, event, req.Data, opts); err != nil {
		return err
	}
	if req.SessionID != "" {
		emitEvent(socketIOEventName(req.SessionID), &SocketIOFrame{
			SessionID: req.SessionID,
			Type:      "message",
			Namespace: opts.Namespace,
			Event:     event,
			Data:      req.Data,
			Timestamp: time.Now().UnixMilli(),
		})
	}
	return nil
}

// SocketIOClose tears down one session by sessionID.
func (s *AppService) SocketIOClose(sessionID string) error {
	socketIOMu.Lock()
	sess, ok := socketIOSessions[sessionID]
	if ok {
		delete(socketIOSessions, sessionID)
	}
	socketIOMu.Unlock()
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
