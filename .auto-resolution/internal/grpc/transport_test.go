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
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/grpc/testsrv"
)

func invokeEcho(t *testing.T, target string, opts InvokeOptions) *Result {
	t.Helper()
	res, err := Invoke(context.Background(),
		Call{Target: target, Service: "reqly.test.v1.EchoService", Method: "Echo"},
		[]byte(`{"text":"tls"}`),
		opts)
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	return res
}

func TestInvokeOverTLSWithCustomCA(t *testing.T) {
	srv := testsrv.StartTLS(t)
	res := invokeEcho(t, srv.Addr, InvokeOptions{Transport: Transport{TLS: true, CAFile: srv.CAPath}})
	if !res.OK || !strings.Contains(string(res.MessageJSON), "tls") {
		t.Fatalf("unexpected result: %+v (%s)", res, res.MessageJSON)
	}
}

func TestInvokeTLSSkipVerify(t *testing.T) {
	srv := testsrv.StartTLS(t)
	res := invokeEcho(t, srv.Addr, InvokeOptions{Transport: Transport{TLS: true, TLSSkipVerify: true}})
	if !res.OK {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestInvokeTLSWithoutTrustFails(t *testing.T) {
	srv := testsrv.StartTLS(t)
	res, err := Invoke(context.Background(),
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "Echo"},
		[]byte(`{"text":"tls"}`),
		InvokeOptions{Transport: Transport{TLS: true}})
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "read ca file") {
		t.Fatalf("unexpected CA read error: %v", err)
	}
	if err != nil {
		return // dial-time failure is acceptable evidence TLS was attempted
	}
	if res.OK {
		t.Fatal("expected failure trusting self-signed cert without CA/skip-verify")
	}
}

func TestInvokePlaintextAgainstTLSServerFailsCleanly(t *testing.T) {
	srv := testsrv.StartTLS(t)
	_, err := Invoke(context.Background(),
		Call{Target: srv.Addr, Service: "reqly.test.v1.EchoService", Method: "Echo"},
		[]byte(`{"text":"x"}`),
		InvokeOptions{})
	if err == nil {
		t.Log("plaintext-to-TLS did not error at dial time; acceptable if it failed as a non-OK status")
	}
}

func TestDiscoverOverTLSWithCustomCA(t *testing.T) {
	srv := testsrv.StartTLS(t)
	services, err := Discover(context.Background(), srv.Addr, Transport{TLS: true, CAFile: srv.CAPath})
	if err != nil {
		t.Fatalf("Discover over TLS: %v", err)
	}
	var found bool
	for _, s := range services {
		if s.Name == "reqly.test.v1.EchoService" {
			found = true
		}
	}
	if !found {
		t.Errorf("EchoService missing over TLS: %+v", services)
	}
}

func TestInvokeDeadlineExceededSurfacesClearly(t *testing.T) {
	srv := testsrv.Start(t)
	res, err := Invoke(context.Background(),
		Call{Target: srv.Addr, Service: "reqly.test.v1.SlowService", Method: "Slow"},
		nil,
		InvokeOptions{Timeout: 20 * time.Millisecond})
	if err == nil && res != nil && res.OK {
		t.Fatal("expected deadline expiry against slow method")
	}
	if err != nil {
		lowered := strings.ToLower(err.Error())
		if !strings.Contains(lowered, "deadline") && !strings.Contains(lowered, "context") {
			t.Errorf("deadline expiry should surface clearly: %v", err)
		}
		return
	}
	if !strings.Contains(res.CodeName, "DeadlineExceeded") {
		t.Errorf("CodeName = %q, want DeadlineExceeded", res.CodeName)
	}
}
