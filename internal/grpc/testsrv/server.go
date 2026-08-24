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

// Package testsrv provides the shared in-process gRPC fixture used by M43
// tests across packages: a plaintext server with reflection enabled, one
// unary echo method, and one server-streaming echo method.
package testsrv

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server is the running fixture; Addr is the plaintext host:port clients dial.
type Server struct {
	Addr string
	srv  *grpc.Server
}

// Greeter is implemented by tests to observe incoming calls.
type Greeter interface {
	Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error)
}

// Start runs the fixture on 127.0.0.1:0 with reflection enabled and registers
// cleanup on t.
func Start(t testing.TB) *Server {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testsrv listen: %v", err)
	}
	srv := grpc.NewServer()
	RegisterEchoServiceServer(srv, &fixture{})
	RegisterFailingServiceServer(srv, &fixture{})
	reflection.Register(srv)
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(srv.Stop)
	return &Server{Addr: ln.Addr().String(), srv: srv}
}

// fixture echoes requests back; streaming responses carry their sequence
// number so ordering assertions are possible.
type fixture struct {
	UnimplementedEchoServiceServer
	UnimplementedFailingServiceServer
}

func (f *fixture) Echo(_ context.Context, req *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{Text: req.GetText(), Labels: req.GetLabels()}, nil
}

func (f *fixture) Boom(_ context.Context, _ *EchoRequest) (*EchoResponse, error) {
	return nil, status.Errorf(codes.NotFound, "thing missing")
}

func (f *fixture) StreamEcho(req *EchoRequest, stream EchoService_StreamEchoServer) error {
	for i := range 3 {
		if err := stream.Send(&EchoResponse{Text: req.GetText(), Sequence: int32(i + 1)}); err != nil {
			return err
		}
	}
	return nil
}

// StartPlain runs the fixture's echo services WITHOUT reflection — the
// negative-path fixture for reflection-disabled servers.
func StartPlain(t testing.TB) *Server {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testsrv listen: %v", err)
	}
	srv := grpc.NewServer()
	RegisterEchoServiceServer(srv, &fixture{})
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(srv.Stop)
	return &Server{Addr: ln.Addr().String(), srv: srv}
}
