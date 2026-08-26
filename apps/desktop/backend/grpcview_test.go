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
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/grpc/testsrv"
	"github.com/Its-Satyajit/reqly/internal/request"
)

func TestGrpcServicesViaBridge(t *testing.T) {
	svc := &AppService{}
	srv := testsrv.Start(t)
	services, err := svc.GrpcServices(GrpcServicesRequest{Target: srv.Addr})
	if err != nil {
		t.Fatalf("GrpcServices: %v", err)
	}
	var found bool
	for _, s := range services {
		if s.Name == "reqly.test.v1.EchoService" {
			found = true
		}
	}
	if !found {
		t.Errorf("EchoService missing: %+v", services)
	}
}

func TestGrpcServicesRequiresTarget(t *testing.T) {
	svc := &AppService{}
	if _, err := svc.GrpcServices(GrpcServicesRequest{}); err == nil {
		t.Fatal("expected error for empty target")
	}
}

func grpcInvokeFixture(t *testing.T) (*AppService, string) {
	t.Helper()
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)
	return svc, wsDir
}

func TestGrpcInvokeUnaryEcho(t *testing.T) {
	svc, _ := grpcInvokeFixture(t)
	srv := testsrv.Start(t)

	res, err := svc.GrpcInvoke(GrpcInvokeRequest{
		SessionID: "s1",
		Request: request.Request{
			URL:     srv.Addr,
			Headers: []request.Header{{Key: "x-reqly", Value: "1"}},
			GRPC: &request.GRPC{
				Service: "reqly.test.v1.EchoService",
				Method:  "Echo",
				Message: map[string]any{"text": "bridge echo"},
			},
		},
	})
	if err != nil {
		t.Fatalf("GrpcInvoke: %v", err)
	}
	if !strings.Contains(string(res.MessageJSON), "bridge echo") {
		t.Errorf("response = %s", res.MessageJSON)
	}
}

func TestGrpcInvokeNonOKStatusIsError(t *testing.T) {
	svc, _ := grpcInvokeFixture(t)
	srv := testsrv.Start(t)

	_, err := svc.GrpcInvoke(GrpcInvokeRequest{
		SessionID: "s2",
		Request: request.Request{
			URL:  srv.Addr,
			GRPC: &request.GRPC{Service: "reqly.test.v1.FailingService", Method: "Boom"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "NotFound") {
		t.Fatalf("expected NotFound status error, got %v", err)
	}
}

func captureGRPCEvents(t *testing.T) *frameRecorder {
	t.Helper()
	rec := &frameRecorder{}
	orig := getEmitRunEvent()
	setEmitRunEvent(func(name string, data any) {
		if e, ok := data.(*GrpcEvent); ok {
			rec.mu.Lock()
			rec.frames = append(rec.frames, &RealtimeFrame{Data: e.Type + "|" + e.Data})
			rec.grpcEvents = append(rec.grpcEvents, e)
			rec.mu.Unlock()
		}
	})
	t.Cleanup(func() { setEmitRunEvent(orig) })
	return rec
}

func TestGrpcStreamDeliversMessagesAndDone(t *testing.T) {
	svc, _ := grpcInvokeFixture(t)
	srv := testsrv.Start(t)
	frames := captureGRPCEvents(t)

	err := svc.GrpcStream(GrpcInvokeRequest{
		SessionID: "stream1",
		Request: request.Request{
			URL:  srv.Addr,
			GRPC: &request.GRPC{Service: "reqly.test.v1.EchoService", Method: "StreamEcho"},
		},
	})
	if err != nil {
		t.Fatalf("GrpcStream: %v", err)
	}

	deadline := time.After(5 * time.Second)
	var done *GrpcEvent
	for done == nil {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for done event")
		default:
		}
		for _, e := range frames.grpcAll() {
			if e.Type == "done" {
				done = e
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	msgs := 0
	for _, e := range frames.grpcAll() {
		if e.Type == "message" && strings.Contains(e.Data, `"sequence"`) {
			msgs++
		}
	}
	if msgs != 3 {
		t.Errorf("streamed %d messages, want 3", msgs)
	}
}

func TestGrpcStreamRejectsDuplicateSession(t *testing.T) {
	svc, _ := grpcInvokeFixture(t)
	srv := testsrv.Start(t)
	req := GrpcInvokeRequest{
		SessionID: "dup",
		Request: request.Request{
			URL:  srv.Addr,
			GRPC: &request.GRPC{Service: "reqly.test.v1.SlowService", Method: "Slow"},
		},
	}
	if err := svc.GrpcStream(req); err != nil {
		t.Fatalf("first stream: %v", err)
	}
	defer func() { _ = svc.GrpcCancel("dup") }()
	time.Sleep(50 * time.Millisecond) // let the goroutine register
	if err := svc.GrpcStream(req); err == nil {
		t.Fatal("expected duplicate-session error")
	}
}

func TestGrpcCancelUnknownSessionIsNoop(t *testing.T) {
	svc := &AppService{}
	if err := svc.GrpcCancel("nope"); err != nil {
		t.Fatalf("cancel unknown session: %v", err)
	}
}
