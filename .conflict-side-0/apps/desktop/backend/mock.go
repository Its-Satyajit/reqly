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
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/Its-Satyajit/reqly/internal/mocking"
	"github.com/Its-Satyajit/reqly/internal/openapi"
)

// MockRoute is one manually defined mock response edited in the panel.
type MockRoute struct {
	Method  string            `json:"method,omitempty"`
	Path    string            `json:"path"`
	Status  int               `json:"status"`
	Body    string            `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Enabled bool              `json:"enabled"`
}

// MockStartRequest starts (or restarts) the workspace mock server.
// SpecPath is an optional workspace-relative or absolute OpenAPI file;
// routes are manual overrides served first. DelayMs adds latency to every
// response; FailEvery simulates a 500 on every Nth request (0 disables).
type MockStartRequest struct {
	SpecPath  string      `json:"specPath,omitempty"`
	Port      int         `json:"port"`
	DelayMs   int         `json:"delayMs,omitempty"`
	FailEvery int         `json:"failEvery,omitempty"`
	Routes    []MockRoute `json:"routes,omitempty"`
}

// MockStatus reports whether the mock server is running, and where.
type MockStatus struct {
	Running bool   `json:"running"`
	URL     string `json:"url,omitempty"`
	Port    int    `json:"port,omitempty"`
	Error   string `json:"error,omitempty"`
}

var (
	mockMu     sync.Mutex
	mockCancel context.CancelFunc
	mockSrv    *http.Server
	mockAddr   string
)

// MockStart spins up the shared mock server. A running server is stopped
// first — the panel owns exactly one instance.
func (s *AppService) MockStart(req MockStartRequest) (*MockStatus, error) {
	if err := s.MockStop(); err != nil {
		return nil, err
	}
	// Port <= 0 lets the kernel pick a free ephemeral port; the panel always
	// sends an explicit value, tests rely on the ephemeral path.
	port := req.Port
	if port < 0 || port > 65535 {
		port = 4010
	}

	opts := []mocking.Option{
		mocking.WithRoutes(routesFromMock(req.Routes)),
	}
	if req.DelayMs > 0 {
		opts = append(opts, mocking.WithDelay(time.Duration(req.DelayMs)*time.Millisecond))
	}
	if req.FailEvery > 1 {
		opts = append(opts, mocking.WithFailureRate(req.FailEvery))
	}

	var handler http.Handler
	if spec := req.SpecPath; spec != "" {
		abs := spec
		if !filepath.IsAbs(abs) && s.root != "" {
			abs = filepath.Join(s.root, filepath.FromSlash(spec))
		}
		doc, err := openapi.LoadFile(abs)
		if err != nil {
			return nil, fmt.Errorf("load spec: %w", err)
		}
		handler, err = mocking.NewServer(doc, opts...)
		if err != nil {
			return nil, fmt.Errorf("build mock: %w", err)
		}
	} else {
		h, err := mocking.NewServer(nil, opts...)
		if err != nil {
			return nil, fmt.Errorf("build mock: %w", err)
		}
		handler = h
	}

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("mock: cannot listen on %s: %w", addr, err)
	}

	_, cancel := context.WithCancel(context.Background())
	server := &http.Server{Handler: handler}
	go func() {
		_ = server.Serve(listener)
	}()

	mockMu.Lock()
	mockCancel = cancel
	mockSrv = server
	mockAddr = listener.Addr().String()
	mockMu.Unlock()

	return &MockStatus{Running: true, URL: "http://" + mockAddr, Port: listener.Addr().(*net.TCPAddr).Port}, nil
}

// MockStop gracefully shuts down the running mock server; stopping a stopped
// server is a no-op.
func (s *AppService) MockStop() error {
	mockMu.Lock()
	cancel, server := mockCancel, mockSrv
	mockCancel, mockSrv, mockAddr = nil, nil, ""
	mockMu.Unlock()
	if cancel == nil {
		return nil
	}
	shutdownCtx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()
	_ = server.Shutdown(shutdownCtx)
	cancel()
	return nil
}

// MockStatusSnapshot reports the current server state.
func (s *AppService) MockStatusSnapshot() *MockStatus {
	mockMu.Lock()
	defer mockMu.Unlock()
	if mockCancel == nil {
		return &MockStatus{Running: false}
	}
	return &MockStatus{Running: true, URL: "http://" + mockAddr, Port: portOf(mockAddr)}
}

func portOf(addr string) int {
	if _, p, err := net.SplitHostPort(addr); err == nil {
		var port int
		_, _ = fmt.Sscanf(p, "%d", &port)
		return port
	}
	return 0
}

func routesFromMock(list []MockRoute) []mocking.Route {
	out := make([]mocking.Route, 0, len(list))
	for _, r := range list {
		out = append(out, mocking.Route{
			Method:  r.Method,
			Path:    r.Path,
			Status:  r.Status,
			Body:    r.Body,
			Headers: r.Headers,
			Enabled: r.Enabled,
		})
	}
	return out
}
