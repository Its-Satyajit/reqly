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
	"net"
	"slices"
	"strings"
	"testing"

	"google.golang.org/grpc"

	"github.com/Its-Satyajit/reqly/internal/grpc/testsrv"
)

func TestDiscoverListsFixtureServices(t *testing.T) {
	srv := testsrv.Start(t)
	services, err := Discover(context.Background(), srv.Addr)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	idx := slices.IndexFunc(services, func(s Service) bool { return s.Name == "reqly.test.v1.EchoService" })
	if idx < 0 {
		t.Fatalf("EchoService not discovered: %+v", services)
	}
	echo := services[idx]
	if len(echo.Methods) != 2 {
		t.Fatalf("methods = %d, want 2 (%+v)", len(echo.Methods), echo.Methods)
	}
	var unary, stream *Method
	for i := range echo.Methods {
		switch echo.Methods[i].Name {
		case "Echo":
			unary = &echo.Methods[i]
		case "StreamEcho":
			stream = &echo.Methods[i]
		}
	}
	if unary == nil || stream == nil {
		t.Fatalf("missing methods: %+v", echo.Methods)
	}
	if unary.ServerStreaming {
		t.Errorf("Echo must not be server-streaming")
	}
	if !stream.ServerStreaming {
		t.Errorf("StreamEcho must be server-streaming")
	}
	if unary.FullName != "/reqly.test.v1.EchoService/Echo" {
		t.Errorf("FullName = %q", unary.FullName)
	}
	if unary.InputType != "reqly.test.v1.EchoRequest" {
		t.Errorf("InputType = %q", unary.InputType)
	}
}

func TestDiscoverServicesSorted(t *testing.T) {
	srv := testsrv.Start(t)
	services, err := Discover(context.Background(), srv.Addr)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	names := make([]string, 0, len(services))
	for _, s := range services {
		names = append(names, s.Name)
	}
	if !slices.IsSorted(names) {
		t.Errorf("services not sorted: %v", names)
	}
}

func TestDiscoverRejectsReflectionDisabledServer(t *testing.T) {
	addr := startPlainServer(t)
	_, err := Discover(context.Background(), addr)
	if err == nil {
		t.Fatal("expected error against reflection-disabled server, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "reflection") {
		t.Errorf("error should name reflection: %v", err)
	}
}

func TestDiscoverUnreachableTarget(t *testing.T) {
	_, err := Discover(context.Background(), "127.0.0.1:1")
	if err == nil {
		t.Fatal("expected error for unreachable target, got nil")
	}
}

// startPlainServer serves gRPC without reflection on 127.0.0.1:0.
func startPlainServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(srv.Stop)
	return ln.Addr().String()
}
