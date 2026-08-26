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
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// frameRecorder accumulates every realtime frame so assertions can scan the
// full history instead of racing a consume-once channel.
type frameRecorder struct {
	mu         sync.Mutex
	frames     []*RealtimeFrame
	grpcEvents []*GrpcEvent
}

func (r *frameRecorder) all() []*RealtimeFrame {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*RealtimeFrame(nil), r.frames...)
}

func (r *frameRecorder) grpcAll() []*GrpcEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*GrpcEvent(nil), r.grpcEvents...)
}

func captureRealtimeFrames(t *testing.T) *frameRecorder {
	t.Helper()
	rec := &frameRecorder{}
	orig := getEmitRunEvent()
	setEmitRunEvent(func(name string, data any) {
		if f, ok := data.(*RealtimeFrame); ok {
			rec.mu.Lock()
			rec.frames = append(rec.frames, f)
			rec.mu.Unlock()
		}
	})
	t.Cleanup(func() { setEmitRunEvent(orig) })
	return rec
}

func waitFrame(t *testing.T, rec *frameRecorder, match func(*RealtimeFrame) bool) *RealtimeFrame {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		for _, f := range rec.all() {
			if match(f) {
				return f
			}
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for frame; got %+v", rec.all())
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func TestRealtimeOpenRequiresFields(t *testing.T) {
	svc := &AppService{}
	if err := svc.RealtimeOpen(RealtimeOpenRequest{Kind: "ws"}); err == nil {
		t.Fatal("open without url succeeded, want error")
	}
	if err := svc.RealtimeOpen(RealtimeOpenRequest{URL: "ws://x"}); err == nil {
		t.Fatal("open without sessionId succeeded, want error")
	}
	if err := svc.RealtimeOpen(RealtimeOpenRequest{SessionID: "s", Kind: "grpc", URL: "x"}); err == nil {
		t.Fatal("open with unknown kind succeeded, want error")
	}
}

func TestRealtimeWSClose(t *testing.T) {
	svc := &AppService{}
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "")
		<-done
	}))
	defer srv.Close()
	if err := svc.RealtimeOpen(RealtimeOpenRequest{
		SessionID: "t1",
		Kind:      "ws",
		URL:       "ws://" + strings.TrimPrefix(srv.URL, "http://"),
	}); err != nil {
		t.Fatalf("RealtimeOpen ws: %v", err)
	}
	realtimeMu.Lock()
	_, ok := realtimeSessions["t1"]
	realtimeMu.Unlock()
	if !ok {
		t.Fatal("session not registered after dial")
	}
}

func TestRealtimeCloseUnknownIsNoop(t *testing.T) {
	svc := &AppService{}
	if err := svc.RealtimeClose("missing"); err != nil {
		t.Fatalf("close of unknown session errored: %v", err)
	}
}

func TestRealtimeWSSendEchoLoopback(t *testing.T) {
	svc := &AppService{}
	frames := captureRealtimeFrames(t)
	upgraded := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		upgraded <- c
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer c.Close(websocket.StatusNormalClosure, "")
		for {
			typ, data, err := c.Read(ctx)
			if err != nil {
				return
			}
			if err := c.Write(ctx, typ, data); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	err := svc.RealtimeOpen(RealtimeOpenRequest{SessionID: "echo", Kind: "ws", URL: wsURL})
	if err != nil {
		t.Fatalf("RealtimeOpen: %v", err)
	}
	defer func() { _ = svc.RealtimeClose("echo") }()

	waitFrame(t, frames, func(f *RealtimeFrame) bool {
		return f.Type == "status" && f.Data == "connected"
	})

	if err := svc.RealtimeSend("echo", `{"ping":1}`); err != nil {
		t.Fatalf("RealtimeSend: %v", err)
	}
	echoed := waitFrame(t, frames, func(f *RealtimeFrame) bool {
		return f.Type == "message" && f.Direction == "in"
	})
	if echoed.Data != `{"ping":1}` {
		t.Errorf("echoed payload = %q", echoed.Data)
	}
	out := waitFrame(t, frames, func(f *RealtimeFrame) bool {
		return f.Direction == "out"
	})
	if out.Data != `{"ping":1}` {
		t.Errorf("outgoing record = %q", out.Data)
	}
}

func TestRealtimeSSEStreamEvents(t *testing.T) {
	svc := &AppService{}
	frames := captureRealtimeFrames(t)
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: tick\nid: 1\ndata: hello\n\n")
		flusher.Flush()
		<-done
	}))
	defer srv.Close()
	defer close(done)

	err := svc.RealtimeOpen(RealtimeOpenRequest{SessionID: "sse1", Kind: "sse", URL: srv.URL + "/events"})
	if err != nil {
		t.Fatalf("RealtimeOpen sse: %v", err)
	}
	defer func() { _ = svc.RealtimeClose("sse1") }()

	ev := waitFrame(t, frames, func(f *RealtimeFrame) bool {
		return f.Type == "message" && f.Name == "tick"
	})
	if ev.ID != "1" || ev.Data != "hello" || ev.Direction != "in" {
		t.Errorf("event frame = %+v", ev)
	}
}

func TestRealtimeWSSendBinaryEchoLoopback(t *testing.T) {
	svc := &AppService{}
	frames := captureRealtimeFrames(t)
	upgraded := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		upgraded <- c
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer c.Close(websocket.StatusNormalClosure, "")
		for {
			typ, data, err := c.Read(ctx)
			if err != nil {
				return
			}
			if typ != websocket.MessageBinary {
				t.Errorf("server got message type %v, want binary", typ)
				return
			}
			if err := c.Write(ctx, typ, data); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	err := svc.RealtimeOpen(RealtimeOpenRequest{SessionID: "bin", Kind: "ws", URL: wsURL})
	if err != nil {
		t.Fatalf("RealtimeOpen: %v", err)
	}
	defer func() { _ = svc.RealtimeClose("bin") }()

	waitFrame(t, frames, func(f *RealtimeFrame) bool {
		return f.Type == "status" && f.Data == "connected"
	})

	payload := base64.StdEncoding.EncodeToString([]byte{0x00, 0xff, 0x10, 0xfe})
	if err := svc.RealtimeSendBinary("bin", payload); err != nil {
		t.Fatalf("RealtimeSendBinary: %v", err)
	}
	echoed := waitFrame(t, frames, func(f *RealtimeFrame) bool {
		return f.Type == "message" && f.Direction == "in"
	})
	if echoed.Encoding != "base64" || echoed.Data != payload {
		t.Errorf("echoed binary frame = %+v", echoed)
	}
	out := waitFrame(t, frames, func(f *RealtimeFrame) bool {
		return f.Direction == "out"
	})
	if out.Encoding != "base64" || out.Data != payload {
		t.Errorf("outgoing binary record = %+v", out)
	}
}

func TestRealtimeSendBinaryRejectsNonBase64(t *testing.T) {
	svc := &AppService{}
	if err := svc.RealtimeSendBinary("missing-session", "not base64!!!"); err == nil {
		t.Fatal("expected error for non-base64 payload, got nil")
	}
}

func TestRealtimeSSERetryHintForwarded(t *testing.T) {
	svc := &AppService{}
	frames := captureRealtimeFrames(t)
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "retry: 3000\n\n")
		fmt.Fprint(w, "data: after-retry\n\n")
		flusher.Flush()
		<-done
	}))
	defer srv.Close()
	defer close(done)

	err := svc.RealtimeOpen(RealtimeOpenRequest{SessionID: "sse2", Kind: "sse", URL: srv.URL + "/events"})
	if err != nil {
		t.Fatalf("RealtimeOpen sse: %v", err)
	}
	defer func() { _ = svc.RealtimeClose("sse2") }()

	ev := waitFrame(t, frames, func(f *RealtimeFrame) bool {
		return f.Type == "message" && f.Data == "after-retry"
	})
	if ev.RetryMs != 3000 {
		t.Errorf("RetryMs = %d, want 3000 (frame %+v)", ev.RetryMs, ev)
	}
}
