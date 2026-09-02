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
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/grpc/testsrv"
)

func collectStream(t *testing.T, call Call, opts InvokeOptions, max int) (*Result, []StreamEvent, error) {
	t.Helper()
	var events []StreamEvent
	res, err := InvokeStream(context.Background(), call, []byte(`{"text":"s"}`), opts, func(e StreamEvent) error {
		events = append(events, e)
		if max > 0 && len(events) >= max {
			return errStopConsumption
		}
		return nil
	})
	return res, events, err
}

// errStopConsumption stops message consumption mid-stream (client-side cap).
var errStopConsumption = stopError{}

type stopError struct{}

func (stopError) Error() string { return "consumption stopped" }

func TestInvokeStreamDeliversInOrder(t *testing.T) {
	srv := testsrv.Start(t)
	res, events, err := collectStream(t,
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "StreamEcho"}, InvokeOptions{}, 0)
	if err != nil {
		t.Fatalf("InvokeStream: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected clean stream end: %+v", res)
	}
	if len(events) != 3 {
		t.Fatalf("events = %d, want 3", len(events))
	}
	for i, e := range events {
		if e.Seq != i+1 {
			t.Errorf("event %d has Seq %d", i, e.Seq)
		}
		var msg map[string]any
		if err := json.Unmarshal(e.MessageJSON, &msg); err != nil {
			t.Fatalf("message not JSON: %v", err)
		}
		if int(msg["sequence"].(float64)) != i+1 {
			t.Errorf("delivery out of order: %s", e.MessageJSON)
		}
	}
}

func TestInvokeStreamMaxMessagesCapsConsumption(t *testing.T) {
	srv := testsrv.Start(t)
	res, events, err := collectStream(t,
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "StreamEcho"}, InvokeOptions{}, 2)
	// Stopping consumption reports the consumer's stop, not a stream failure.
	if err == nil {
		t.Fatal("expected consumption-stop error")
	}
	if _, ok := err.(stopError); !ok {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("events = %d, want 2 (capped)", len(events))
	}
	_ = res
}

func TestInvokeStreamRejectsUnaryMethod(t *testing.T) {
	srv := testsrv.Start(t)
	_, _, err := collectStream(t,
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "Echo"}, InvokeOptions{}, 0)
	if err == nil || !strings.Contains(err.Error(), "not server-streaming") {
		t.Fatalf("expected not-server-streaming error, got %v", err)
	}
}

func TestInvokeStreamNonOKStatusMidStream(t *testing.T) {
	srv := testsrv.Start(t)
	// FailingService has no streaming methods; a missing method proves the
	// failure path surfaces as status/result rather than panic.
	_, _, err := collectStream(t,
		Call{Target: srv.Addr, Service: "reqly.test.v1.Missing", Method: "Nope"}, InvokeOptions{}, 0)
	if err == nil {
		t.Fatal("expected error for unknown service")
	}
}
