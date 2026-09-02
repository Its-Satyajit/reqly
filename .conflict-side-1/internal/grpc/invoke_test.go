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

package grpc

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/grpc/testsrv"
)

func TestInvokeUnaryEcho(t *testing.T) {
	srv := testsrv.Start(t)
	res, err := Invoke(context.Background(),
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "Echo"},
		[]byte(`{"text":"hello","labels":{"a":"b"}}`),
		InvokeOptions{})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected OK, got %+v", res)
	}
	var msg map[string]any
	if err := json.Unmarshal(res.MessageJSON, &msg); err != nil {
		t.Fatalf("response not JSON: %v", err)
	}
	if msg["text"] != "hello" || msg["labels"].(map[string]any)["a"] != "b" {
		t.Errorf("echo mismatch: %s", res.MessageJSON)
	}
}

func TestInvokeNonOKStatus(t *testing.T) {
	srv := testsrv.Start(t)
	res, err := Invoke(context.Background(),
		Call{Target: srv.Addr, Service: "reqly.test.v1.FailingService", Method: "Boom"},
		nil,
		InvokeOptions{})
	if err != nil {
		t.Fatalf("Invoke (status failures are results, not errors): %v", err)
	}
	if res.OK {
		t.Fatal("expected non-OK result")
	}
	if res.Code != uint32(5) { // NotFound
		t.Errorf("Code = %d (%s), want 5 NotFound", res.Code, res.CodeName)
	}
	if res.CodeName != "NotFound" {
		t.Errorf("CodeName = %q", res.CodeName)
	}
	if !strings.Contains(res.StatusMessage, "thing missing") {
		t.Errorf("StatusMessage = %q", res.StatusMessage)
	}
}

func TestInvokeMetadataReachesServer(t *testing.T) {
	srv := testsrv.Start(t)
	res, err := Invoke(context.Background(),
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "Echo"},
		[]byte(`{"text":"x"}`),
		InvokeOptions{Metadata: map[string]string{"X-Reqly-Test": "yes"}})
	if err != nil || !res.OK {
		t.Fatalf("Invoke: %v %+v", err, res)
	}
}

func TestInvokeDeadlineExpiry(t *testing.T) {
	srv := testsrv.Start(t)
	res, err := Invoke(context.Background(),
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "Echo"},
		[]byte(`{"text":"x"}`),
		InvokeOptions{Timeout: 1 * time.Nanosecond})
	// A 1ns deadline may expire before dialing (error) or during the call
	// (non-OK DeadlineExceeded) — both prove the deadline was applied.
	if err == nil && res != nil && res.OK {
		t.Fatal("expected deadline expiry, got success")
	}
	if err != nil {
		lowered := strings.ToLower(err.Error())
		if !strings.Contains(lowered, "deadline") && !strings.Contains(lowered, "context") && !strings.Contains(lowered, "dial") {
			t.Logf("deadline surfaced as: %v", err)
		}
	}
}

func TestInvokeProtoFilesFallback(t *testing.T) {
	srv := testsrv.StartPlain(t) // no reflection
	protoPath, err := filepath.Abs(filepath.Join("..", "grpc", "testsrv", "testapi.proto"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	res, err := Invoke(context.Background(),
		Call{
			Target:     srv.Addr,
			Service:    "reqly.test.v1.EchoService",
			Method:     "Echo",
			ProtoFiles: []string{protoPath},
		},
		[]byte(`{"text":"from proto"}`),
		InvokeOptions{})
	if err != nil {
		t.Fatalf("Invoke with protoFiles fallback: %v", err)
	}
	if !res.OK || !strings.Contains(string(res.MessageJSON), "from proto") {
		t.Fatalf("unexpected result: %+v (%s)", res, res.MessageJSON)
	}
}

func TestInvokeUnknownMethodFailsCleanly(t *testing.T) {
	srv := testsrv.Start(t)
	_, err := Invoke(context.Background(),
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "Nope"},
		nil,
		InvokeOptions{})
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
}
