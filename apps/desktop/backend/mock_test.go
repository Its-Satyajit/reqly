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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestMockStartStopRouteRoundTrip(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)

	status, err := svc.MockStart(MockStartRequest{
		Port: 0, // ephemeral; MockStart maps <=0 to 4010 — use free port below
		Routes: []MockRoute{
			{Method: "GET", Path: "/ping", Status: 200, Body: "pong", Enabled: true},
		},
	})
	if err != nil {
		t.Fatalf("MockStart: %v", err)
	}
	defer func() { _ = svc.MockStop() }()

	resp, err := http.Get(status.URL + "/ping")
	if err != nil {
		t.Fatalf("GET /ping: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || string(body) != "pong" {
		t.Fatalf("status=%d body=%q", resp.StatusCode, body)
	}

	snap := svc.MockStatusSnapshot()
	if !snap.Running || snap.URL != status.URL {
		t.Errorf("snapshot = %+v", snap)
	}

	if err := svc.MockStop(); err != nil {
		t.Fatalf("MockStop: %v", err)
	}
	snap = svc.MockStatusSnapshot()
	if snap.Running {
		t.Error("still running after stop")
	}
}

func TestMockRestartReplacesServer(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)
	first, err := svc.MockStart(MockStartRequest{Port: 49160, Routes: []MockRoute{
		{Path: "/a", Status: 200, Body: "1", Enabled: true},
	}})
	if err != nil {
		t.Fatalf("first start: %v", err)
	}
	second, err := svc.MockStart(MockStartRequest{Port: 49161, Routes: []MockRoute{
		{Path: "/b", Status: 200, Body: "2", Enabled: true},
	}})
	if err != nil {
		t.Fatalf("second start: %v", err)
	}
	defer func() { _ = svc.MockStop() }()
	if first.URL == second.URL {
		t.Error("restart reused the same address")
	}
	if _, err := http.Get(first.URL + "/a"); err == nil {
		t.Error("old server still serving after restart")
	}
}

func TestMockSpecLoadError(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	if _, err := svc.MockStart(MockStartRequest{Port: 49162, SpecPath: filepath.Join(wsDir, "nope.yaml")}); err == nil {
		t.Fatal("start with missing spec succeeded, want error")
	}
	_ = os.WriteFile(filepath.Join(wsDir, "x"), []byte(""), 0o600)
}

func TestMockSpecServesExamples(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	spec := `openapi: 3.0.3
info: {title: T, version: "1.0.0"}
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: array
                items: {type: string}
`
	if err := os.WriteFile(filepath.Join(wsDir, "pets.yaml"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := svc.MockStart(MockStartRequest{Port: 49163, SpecPath: "pets.yaml"})
	if err != nil {
		t.Fatalf("MockStart spec: %v", err)
	}
	defer func() { _ = svc.MockStop() }()
	resp, err := http.Get(status.URL + "/pets")
	if err != nil {
		t.Fatalf("GET /pets: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
