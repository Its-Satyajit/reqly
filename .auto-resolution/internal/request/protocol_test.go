// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_HTTPVersionAndKeepAlive(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer ts.Close()

	client := NewClient()
	req := &Request{
		Method:            MethodGet,
		URL:               ts.URL,
		HTTPVersion:       "http1.1",
		DisableKeepAlives: true,
	}

	resp, err := client.Execute(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error executing request with HTTP controls: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
}
