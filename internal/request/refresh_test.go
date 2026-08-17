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

package request

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// oauth2CachedClient returns a client with a fresh store-backed token cache.
func oauth2CachedClient(t *testing.T) (*Client, *secrets.FileStore, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	return NewClient(WithTokenCache(store, dir)), store, dir
}

func oauth2Req(tokenURL, apiURL string) *Request {
	return &Request{
		Method: MethodGet,
		URL:    apiURL,
		Auth: Auth{
			Type: "oauth2",
			Config: map[string]string{
				"grant_type":    "client_credentials",
				"token_url":     tokenURL,
				"client_id":     "client-123",
				"client_secret": "s3cr3t",
			},
		},
	}
}

func TestExecuteOAuth2ExpiredTokenRefetched(t *testing.T) {
	client, _, _ := oauth2CachedClient(t)

	var tokenCalls, apiCalls int
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		// First token already-expired when cached with a 1s lifetime; the
		// second is long-lived. A fresh acquire must happen before the send.
		w.Header().Set("Content-Type", "application/json")
		if tokenCalls == 1 {
			w.Write([]byte(`{"access_token":"tok-expired","token_type":"Bearer","expires_in":1}`))
			return
		}
		w.Write([]byte(`{"access_token":"tok-fresh","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	var gotAuth []string
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	req := oauth2Req(tokenSrv.URL, apiSrv.URL)
	for i := 0; i < 2; i++ {
		if _, err := client.Execute(context.Background(), req, variables.NewSet()); err != nil {
			t.Fatalf("Execute %d: %v", i, err)
		}
		// Let the 1s token actually expire between requests.
		time.Sleep(1100 * time.Millisecond)
	}

	if tokenCalls != 2 {
		t.Fatalf("token endpoint called %d times, want 2 (expiry-driven refresh)", tokenCalls)
	}
	if apiCalls != 2 {
		t.Fatalf("api called %d times, want 2", apiCalls)
	}
	if gotAuth[0] != "Bearer tok-expired" || gotAuth[1] != "Bearer tok-fresh" {
		t.Fatalf("authorizations = %v", gotAuth)
	}
}

func TestExecuteOAuth2Reactive401RefreshesAndRetries(t *testing.T) {
	client, _, _ := oauth2CachedClient(t)

	var tokenCalls int
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		w.Header().Set("Content-Type", "application/json")
		if tokenCalls == 1 {
			w.Write([]byte(`{"access_token":"tok-stale","token_type":"Bearer","expires_in":3600}`))
			return
		}
		w.Write([]byte(`{"access_token":"tok-new","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	var apiCalls int
	var gotAuth []string
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		// First attempt with the stale token is rejected; the retried
		// request with the refreshed token succeeds.
		if r.Header.Get("Authorization") == "Bearer tok-stale" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	req := oauth2Req(tokenSrv.URL, apiSrv.URL)
	resp, err := client.Execute(context.Background(), req, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 after forced refresh+retry", resp.StatusCode)
	}
	if apiCalls != 2 {
		t.Fatalf("api called %d times, want 2 (original + retry)", apiCalls)
	}
	if tokenCalls != 2 {
		t.Fatalf("token endpoint called %d times, want 2 (initial + forced refresh)", tokenCalls)
	}
	if len(gotAuth) != 2 || gotAuth[0] != "Bearer tok-stale" || gotAuth[1] != "Bearer tok-new" {
		t.Fatalf("authorizations = %v", gotAuth)
	}
}

func TestExecuteOAuth2NoRetryLoop(t *testing.T) {
	client, _, _ := oauth2CachedClient(t)

	var tokenCalls, apiCalls int
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok-always-401","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer apiSrv.Close()

	req := oauth2Req(tokenSrv.URL, apiSrv.URL)
	resp, err := client.Execute(context.Background(), req, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (retry result returned as-is)", resp.StatusCode)
	}
	if apiCalls != 2 {
		t.Fatalf("api called %d times, want 2 (exactly one retry)", apiCalls)
	}
	if tokenCalls != 2 {
		t.Fatalf("token endpoint called %d times, want 2 (no extra acquires)", tokenCalls)
	}
}

func TestExecuteOAuth2ConcurrentNoDoubleAcquire(t *testing.T) {
	client, _, _ := oauth2CachedClient(t)

	var tokenCalls atomic.Int64
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls.Add(1)
		time.Sleep(20 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	req := oauth2Req(tokenSrv.URL, apiSrv.URL)

	const n = 8
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := client.Execute(context.Background(), req, variables.NewSet()); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent Execute: %v", err)
	}

	if calls := tokenCalls.Load(); calls != 1 {
		t.Fatalf("token endpoint called %d times for %d concurrent requests, want 1", calls, n)
	}
}

// TestExecuteOAuth2Reactive401UsesRefreshTokenGrant verifies the reactive
// path renews a stale auth-code token via the refresh-token grant — never
// re-running the browser flow — and retries exactly once.
func TestExecuteOAuth2Reactive401UsesRefreshTokenGrant(t *testing.T) {
	// Token endpoint: serves the refresh-token grant with a fresh token and
	// records every grant_type it sees.
	var grants []string
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		grants = append(grants, form.Get("grant_type"))
		if form.Get("grant_type") != "refresh_token" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok-refreshed","token_type":"Bearer","expires_in":3600,"refresh_token":"rt-2"}`))
	}))
	defer tokenSrv.Close()

	// API server: rejects the stale token once, accepts the refreshed one.
	var apiCalls int
	var gotAuth []string
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		if r.Header.Get("Authorization") == "Bearer tok-stale" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	cfg := map[string]string{
		"grant_type":    "authorization_code",
		"token_url":     tokenSrv.URL,
		"client_id":     "client-123",
		"client_secret": "s3cr3t",
	}
	// Seed a fresh-but-stale auth-code token: the proactive path keeps it
	// (fresh expiry), the API 401 forces the refresh-token grant.
	blob, err := json.Marshal(map[string]any{
		"access_token":  "tok-stale",
		"token_type":    "Bearer",
		"expiry":        time.Now().Add(1 * time.Hour),
		"refresh_token": "rt-1",
		"grant_type":    "authorization_code",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set(auth.TokenCacheKey(dir, cfg), string(blob)); err != nil {
		t.Fatal(err)
	}

	client := NewClient(WithTokenCache(store, dir))
	req := &Request{
		Method: MethodGet,
		URL:    apiSrv.URL,
		Auth: Auth{
			Type:   "oauth2",
			Config: cfg,
		},
	}
	resp, err := client.Execute(context.Background(), req, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 after refresh-token renewal", resp.StatusCode)
	}
	if apiCalls != 2 {
		t.Fatalf("api called %d times, want 2 (original + retry)", apiCalls)
	}
	if len(gotAuth) != 2 || gotAuth[0] != "Bearer tok-stale" || gotAuth[1] != "Bearer tok-refreshed" {
		t.Fatalf("authorizations = %v", gotAuth)
	}
	for _, g := range grants {
		if g != "refresh_token" {
			t.Fatalf("token endpoint saw grant_type %q, want only refresh_token (no browser/exchange)", g)
		}
	}
	if len(grants) != 1 {
		t.Fatalf("token endpoint called %d times, want 1 (exactly one refresh)", len(grants))
	}

	// The store must hold the refreshed token with the rotated refresh token.
	raw, err := store.Get(auth.TokenCacheKey(dir, cfg))
	if err != nil {
		t.Fatal(err)
	}
	cached, err := auth.ParseCachedToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	if cached.AccessToken != "tok-refreshed" || cached.RefreshToken != "rt-2" {
		t.Fatalf("stored = %q/%q, want tok-refreshed/rt-2", cached.AccessToken, cached.RefreshToken)
	}
}
