// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
	resp, err := svc.Send(request.Request{
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
	resp, err := svc.Send(request.Request{
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
	_, err := svc.Send(request.Request{
		Method: request.MethodGet,
		URL:    "://not-a-url",
	})
	if err == nil {
		t.Fatal("expected error for invalid URL")
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
	_, err := svc.Send(request.Request{
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
