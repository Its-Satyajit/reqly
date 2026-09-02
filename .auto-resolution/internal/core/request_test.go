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

package core

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestSendReturnsBridgeFriendlyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Mock", "yes")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2}`))
	}))
	defer srv.Close()

	svc := NewRequestService()
	resp, err := svc.Send(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !resp.OK {
		t.Fatal("expected OK to be true for 2xx")
	}
	if resp.StatusText != "OK" {
		t.Fatalf("expected status text OK, got %q", resp.StatusText)
	}
	if resp.Body != `{"count":2}` {
		t.Fatalf("unexpected body: %q", resp.Body)
	}
	if resp.Headers["X-Mock"][0] != "yes" {
		t.Fatalf("unexpected headers: %v", resp.Headers)
	}
	if resp.Proto == "" {
		t.Fatal("expected proto to be set")
	}
	if resp.Size == 0 {
		t.Fatal("expected size to be set")
	}
	if resp.DurationMS < 0 {
		t.Fatalf("expected non-negative duration, got %d", resp.DurationMS)
	}
}

func TestSendReportsNonSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	}))
	defer srv.Close()

	svc := NewRequestService()
	resp, err := svc.Send(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", resp.StatusCode)
	}
	if resp.OK {
		t.Fatal("expected OK to be false for 4xx")
	}
}

func TestSendReturnsErrorForInvalidURL(t *testing.T) {
	svc := NewRequestService()
	_, err := svc.Send(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    "://not-a-url",
	})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestSendHonoursContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := NewRequestService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	_, err := svc.Send(ctx, request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("cancelled send took %v, expected prompt return", elapsed)
	}
}

func TestSendUsesProvidedVariableSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer env-token" {
			t.Errorf("expected interpolated env token, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	vars := variables.NewSet()
	vars.Set(variables.ScopeEnvironment, "TOKEN", "env-token")

	svc := NewRequestService()
	_, err := svc.Send(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
		Headers: []request.Header{
			{Key: "Authorization", Value: "Bearer {{TOKEN}}"},
		},
	}, vars)
	if err != nil {
		t.Fatal(err)
	}
}
