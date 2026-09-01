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

package request

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

func mustVars(t *testing.T, set func(*variables.Set)) *variables.Set {
	t.Helper()
	v := variables.NewSet()
	if set != nil {
		set(v)
	}
	return v
}

func TestExecuteReturnsStatusAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodPost,
		URL:    srv.URL,
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if resp.StatusText != "Created" {
		t.Fatalf("expected 'Created', got %q", resp.StatusText)
	}
	if !resp.OK() {
		t.Fatal("expected OK() to be true for 2xx")
	}
	if string(resp.Body) != `{"id":1}` {
		t.Fatalf("unexpected body %q", resp.Body)
	}
	if resp.Size != 8 {
		t.Fatalf("expected size 8, got %d", resp.Size)
	}
	if resp.Duration <= 0 {
		t.Fatal("expected positive duration")
	}
}

func TestExecuteInterpolatesVariables(t *testing.T) {
	var gotURL string
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		gotHeader = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	vars := mustVars(t, func(v *variables.Set) {
		v.Set(variables.ScopeEnvironment, "base", srv.URL)
		v.Set(variables.ScopeEnvironment, "apiKey", "secret-key")
	})
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    "{{base}}/users?active=true",
		Headers: []Header{
			{Key: "X-API-Key", Value: "{{apiKey}}"},
		},
	}, vars)
	if err != nil {
		t.Fatal(err)
	}

	if gotURL != "/users?active=true" {
		t.Fatalf("unexpected URL %q", gotURL)
	}
	if gotHeader != "secret-key" {
		t.Fatalf("expected header 'secret-key', got %q", gotHeader)
	}
}

func TestExecuteQueryParameters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.RawQuery))
	}))
	defer srv.Close()

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL + "?page=1",
		Query: []Parameter{
			{Key: "page", Value: "2"},
			{Key: "limit", Value: "10"},
		},
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	query := string(resp.Body)
	if !strings.Contains(query, "limit=10") {
		t.Fatalf("expected limit=10 in query %q", query)
	}
	if strings.Contains(query, "page=1") {
		t.Fatalf("expected page override, got %q", query)
	}
}

func TestExecuteSendsBody(t *testing.T) {
	var gotBody string
	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		gotBody = string(buf)
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodPost,
		URL:    srv.URL,
		Body:   `{"name":"{{name}}"}`,
	}, mustVars(t, func(v *variables.Set) {
		v.Set(variables.ScopeRequest, "name", "Reqly")
	}))
	if err != nil {
		t.Fatal(err)
	}

	if gotBody != `{"name":"Reqly"}` {
		t.Fatalf("unexpected body %q", gotBody)
	}
	if gotContentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", gotContentType)
	}
}

func TestExecuteBearerAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Auth: Auth{
			Type: "bearer",
			Config: map[string]string{
				"token": "{{token}}",
			},
		},
	}, mustVars(t, func(v *variables.Set) {
		v.Set(variables.ScopeRequest, "token", "abc123")
	}))
	if err != nil {
		t.Fatal(err)
	}

	if gotAuth != "Bearer abc123" {
		t.Fatalf("expected 'Bearer abc123', got %q", gotAuth)
	}
}

func TestExecuteOAuth2ClientCredentials(t *testing.T) {
	var gotAuth string
	var tokenCalls int
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok-oauth","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    apiSrv.URL,
		Auth: Auth{
			Type: "oauth2",
			Config: map[string]string{
				"grant_type":    "client_credentials",
				"token_url":     tokenSrv.URL,
				"client_id":     "client-123",
				"client_secret": "s3cr3t",
			},
		},
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if gotAuth != "Bearer tok-oauth" {
		t.Fatalf("expected 'Bearer tok-oauth', got %q", gotAuth)
	}
	if tokenCalls != 1 {
		t.Fatalf("token endpoint called %d times, want 1", tokenCalls)
	}
}

func TestExecuteOAuth2TokenCachedAcrossRequests(t *testing.T) {
	var tokenCalls int
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok-cached","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	client := NewClient(WithTokenCache(store, dir))

	req := &Request{
		Method: MethodGet,
		URL:    apiSrv.URL,
		Auth: Auth{
			Type: "oauth2",
			Config: map[string]string{
				"grant_type":    "client_credentials",
				"token_url":     tokenSrv.URL,
				"client_id":     "client-123",
				"client_secret": "s3cr3t",
			},
		},
	}

	if _, err := client.Execute(context.Background(), req, variables.NewSet()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Execute(context.Background(), req, variables.NewSet()); err != nil {
		t.Fatal(err)
	}
	if tokenCalls != 1 {
		t.Fatalf("token endpoint called %d times across two requests, want 1 (cached)", tokenCalls)
	}
}

func TestExecuteOAuth2CacheKeyScopedToConfig(t *testing.T) {
	var tokenCalls int
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	client := NewClient(WithTokenCache(store, dir))

	base := func(clientID string) *Request {
		return &Request{
			Method: MethodGet,
			URL:    apiSrv.URL,
			Auth: Auth{
				Type: "oauth2",
				Config: map[string]string{
					"grant_type":    "client_credentials",
					"token_url":     tokenSrv.URL,
					"client_id":     clientID,
					"client_secret": "s3cr3t",
				},
			},
		}
	}

	if _, err := client.Execute(context.Background(), base("client-a"), variables.NewSet()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Execute(context.Background(), base("client-b"), variables.NewSet()); err != nil {
		t.Fatal(err)
	}
	if tokenCalls != 2 {
		t.Fatalf("token endpoint called %d times for distinct client_ids, want 2", tokenCalls)
	}
}

func TestExecuteBasicAuth(t *testing.T) {
	var gotUser, gotPass string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Auth: Auth{
			Type: "basic",
			Config: map[string]string{
				"username": "user",
				"password": "pass",
			},
		},
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if gotUser != "user" || gotPass != "pass" {
		t.Fatalf("expected user/pass, got %q/%q", gotUser, gotPass)
	}
}

func TestExecuteDigestAuth(t *testing.T) {
	const (
		realm    = "testrealm@host.com"
		nonce    = "dcd98b7102dd2f0e8b11d0f600bfb0c093"
		username = "Mufasa"
		password = "Circle Of Life"
	)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.Header().Set("WWW-Authenticate",
				`Digest realm="testrealm@host.com", qop="auth", nonce="`+nonce+`"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Server-side verification: recompute the expected response from the
		// client-supplied nonce, cnonce, nc, and uri.
		params := parseTestDigestHeader(t, auth)
		HA1 := md5HexForTest(username + ":" + realm + ":" + password)
		HA2 := md5HexForTest(r.Method + ":" + r.URL.RequestURI())
		want := md5HexForTest(HA1 + ":" + params["nonce"] + ":" + params["nc"] +
			":" + params["cnonce"] + ":auth:" + HA2)
		if params["response"] != want {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Auth: Auth{
			Type: "digest",
			Config: map[string]string{
				"username": username,
				"password": password,
			},
		},
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 after digest retry, got %d (calls=%d)", resp.StatusCode, calls)
	}
	if calls != 2 {
		t.Fatalf("expected exactly 2 calls (challenge + retry), got %d", calls)
	}
}

func TestExecuteDigestRetryIsBounded(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("WWW-Authenticate",
			`Digest realm="r", qop="auth", nonce="neverchanges"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Auth: Auth{
			Type: "digest",
			Config: map[string]string{
				"username": "u",
				"password": "p",
			},
		},
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if calls != 2 {
		t.Fatalf("expected exactly 2 requests (initial + one retry), got %d", calls)
	}
}

func TestExecuteDigestNonDigest401NotRetried(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		// A Basic challenge is not a Digest challenge: the client must not
		// attempt a digest retry nor turn the 401 into an error.
		w.Header().Set("WWW-Authenticate", `Basic realm="example"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Auth: Auth{
			Type: "digest",
			Config: map[string]string{
				"username": "u",
				"password": "p",
			},
		},
	}, variables.NewSet())
	if err != nil {
		t.Fatalf("non-Digest 401 must not error, got %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if calls != 1 {
		t.Fatalf("expected only 1 request for non-Digest challenge, got %d", calls)
	}
}

func TestExecuteDigestAuthMissingPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Auth: Auth{
			Type:   "digest",
			Config: map[string]string{"username": "u"},
		},
	}, variables.NewSet())
	if err == nil {
		t.Fatal("expected error for missing digest password")
	}
}

func parseTestDigestHeader(t *testing.T, hdr string) map[string]string {
	t.Helper()
	params := make(map[string]string)
	rest := strings.TrimPrefix(hdr, "Digest ")
	for _, part := range strings.Split(rest, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		params[kv[0]] = strings.Trim(kv[1], `"`)
	}
	return params
}

func md5HexForTest(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestExecuteAPIKeyInHeader(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Auth: Auth{
			Type: "apikey",
			Config: map[string]string{
				"key":   "X-API-Key",
				"value": "k-123",
				"in":    "header",
			},
		},
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if gotKey != "k-123" {
		t.Fatalf("expected 'k-123', got %q", gotKey)
	}
}

func TestExecuteRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	start := time.Now()
	_, err := client.Execute(context.Background(), &Request{
		Method:  MethodGet,
		URL:     srv.URL,
		Timeout: 50,
	}, variables.NewSet())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if time.Since(start) > 150*time.Millisecond {
		t.Fatalf("timeout took too long: %s", time.Since(start))
	}
}

func TestExecuteUndefinedVariable(t *testing.T) {
	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    "https://example.com/{{missing}}",
	}, variables.NewSet())
	if err == nil {
		t.Fatal("expected undefined-variable error")
	}
	if !strings.Contains(err.Error(), "undefined variable") {
		t.Fatalf("unexpected error %q", err)
	}
}

func TestExecuteInvalidURL(t *testing.T) {
	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    "://bad url",
	}, variables.NewSet())
	if err == nil {
		t.Fatal("expected URL parse error")
	}
}

func TestExecuteNoURL(t *testing.T) {
	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
	}, variables.NewSet())
	if err == nil {
		t.Fatal("expected missing-URL error")
	}
}

func TestExecuteHeadersPresentOnWire(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Headers: []Header{
			{Key: "X-Custom", Value: "hello"},
		},
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if gotHeader != "hello" {
		t.Fatalf("expected 'hello', got %q", gotHeader)
	}
}

func TestExecuteServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 500 {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	if resp.OK() {
		t.Fatal("expected OK() to be false for 5xx")
	}
	if string(resp.Body) != "boom\n" {
		t.Fatalf("unexpected body %q", resp.Body)
	}
}

func TestExecuteContextCancellation(t *testing.T) {
	started := make(chan struct{})
	observed := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(observed)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())

	client := NewClient()
	done := make(chan error, 1)
	go func() {
		_, err := client.Execute(ctx, &Request{
			Method: MethodGet,
			URL:    srv.URL,
		}, variables.NewSet())
		done <- err
	}()

	<-started
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected context cancellation error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Execute did not return after cancellation")
	}

	select {
	case <-observed:
	case <-time.After(2 * time.Second):
		t.Fatal("expected server to observe cancellation")
	}
}

func TestResponseOKBoundaries(t *testing.T) {
	for _, tc := range []struct {
		code int
		ok   bool
	}{
		{199, false},
		{200, true},
		{299, true},
		{300, false},
	} {
		if got := (&response.Response{StatusCode: tc.code}).OK(); got != tc.ok {
			t.Fatalf("expected OK()=%v for %d", tc.ok, tc.code)
		}
	}
}

func TestExecuteConcurrent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	const n = 20
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			_, err := client.Execute(context.Background(), &Request{
				Method: MethodGet,
				URL:    srv.URL,
			}, variables.NewSet())
			errs <- err
		}()
	}
	for i := 0; i < n; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
}

// TestExecuteOAuth2AuthorizationCodeFlow drives a full Authorization Code +
// PKCE exchange through the request engine: a registered TokenSource scheme
// (test-only name, browser driver injected via the Open hook) acquires a
// token from a loopback callback and Bearer-attaches it to the API request.
func TestExecuteOAuth2AuthorizationCodeFlow(t *testing.T) {
	// Fake provider: authorization endpoint redirects to the flow's callback
	// with a code and the echoed state.
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("response_type"); got != "code" {
			t.Errorf("response_type = %q, want code", got)
		}
		if got := q.Get("code_challenge_method"); got != "S256" {
			t.Errorf("code_challenge_method = %q, want S256", got)
		}
		if got := q.Get("code_challenge"); got == "" {
			t.Error("code_challenge is empty")
		}
		if got := q.Get("state"); got == "" {
			t.Error("state is empty")
		}
		cb, err := url.Parse(q.Get("redirect_uri"))
		if err != nil {
			http.Error(w, "bad redirect_uri", http.StatusBadRequest)
			return
		}
		cbq := cb.Query()
		cbq.Set("code", "engine-code")
		cbq.Set("state", q.Get("state"))
		cb.RawQuery = cbq.Encode()
		http.Redirect(w, r, cb.String(), http.StatusFound)
	}))
	defer authSrv.Close()

	// Fake token endpoint: records the exchange and issues a token.
	var gotUser, gotPass string
	var gotForm url.Values
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		body, _ := io.ReadAll(r.Body)
		gotForm, _ = url.ParseQuery(string(body))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"engine-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"engine-rt"}`))
	}))
	defer tokenSrv.Close()

	// The API server records the Authorization header it receives.
	var gotAuth string
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	// A test-only scheme name lets the engine exercise the auth-code flow
	// with the browser driver injected (no global state).
	if _, ok := auth.Lookup("oauth2-authcode-test"); !ok {
		auth.Register("oauth2-authcode-test", authCodeTestScheme{
			src: &auth.AuthorizationCodeSource{
				Open: func(_ context.Context, authorizationURL string) error {
					// Simulate the browser: follow the provider redirect back
					// to the loopback callback.
					client := &http.Client{Timeout: 10 * time.Second}
					resp, err := client.Get(authorizationURL)
					if err != nil {
						return err
					}
					defer resp.Body.Close()
					_, _ = io.Copy(io.Discard, resp.Body)
					return nil
				},
			},
		})
	}

	client := NewClient()
	req := &Request{
		Method: MethodGet,
		URL:    apiSrv.URL,
		Auth: Auth{
			Type: "oauth2-authcode-test",
			Config: map[string]string{
				"authorization_url": authSrv.URL,
				"token_url":         tokenSrv.URL,
				"client_id":         "engine-client",
				"client_secret":     "engine-secret",
			},
		},
	}
	resp, err := client.Execute(context.Background(), req, variables.NewSet())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if gotAuth != "Bearer engine-tok" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer engine-tok")
	}
	if gotUser != "engine-client" || gotPass != "engine-secret" {
		t.Fatalf("token Basic auth = %q/%q, want engine-client/engine-secret", gotUser, gotPass)
	}
	if gotForm.Get("grant_type") != "authorization_code" {
		t.Errorf("grant_type = %q, want authorization_code", gotForm.Get("grant_type"))
	}
	if gotForm.Get("code") != "engine-code" {
		t.Errorf("code = %q, want engine-code", gotForm.Get("code"))
	}
	if v := gotForm.Get("code_verifier"); len(v) < 43 || len(v) > 128 {
		t.Errorf("code_verifier length %d outside 43-128", len(v))
	}
	if gotForm.Get("client_id") != "engine-client" {
		t.Errorf("client_id = %q, want engine-client", gotForm.Get("client_id"))
	}
	if gotForm.Get("redirect_uri") == "" {
		t.Error("redirect_uri missing from exchange body")
	}
}

// authCodeTestScheme wraps an AuthorizationCodeSource as a full Scheme so
// the engine integration test can register it (the product AuthorizationCode
// Source stays a pure TokenSource; oauth2Scheme provides Apply in production).
type authCodeTestScheme struct {
	src *auth.AuthorizationCodeSource
}

func (s authCodeTestScheme) Token(ctx context.Context, cfg map[string]string, vars auth.Interpolator) (auth.Token, error) {
	return s.src.Token(ctx, cfg, vars)
}

func (authCodeTestScheme) Apply(req *http.Request, cfg map[string]string, _ auth.Interpolator) error {
	token := cfg["token"]
	if token == "" {
		return fmt.Errorf("oauth2 auth requires a token; was a token acquired before apply?")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (authCodeTestScheme) SecretKeys() []string { return []string{"client_secret"} }

// TestExecuteOAuth2AuthorizationCodeAutoLogin verifies first-request
// auto-login: with the CLI-style browser opener installed, a request with
// grant_type authorization_code and no cached token completes the browser
// flow end to end, and a second request reuses the cached token without
// re-running the flow.
func TestExecuteOAuth2AuthorizationCodeAutoLogin(t *testing.T) {
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		cb, err := url.Parse(q.Get("redirect_uri"))
		if err != nil {
			http.Error(w, "bad redirect_uri", http.StatusBadRequest)
			return
		}
		cbq := cb.Query()
		cbq.Set("code", "auto-code")
		cbq.Set("state", q.Get("state"))
		cb.RawQuery = cbq.Encode()
		http.Redirect(w, r, cb.String(), http.StatusFound)
	}))
	defer authSrv.Close()

	var tokenCalls int
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"auto-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"auto-rt"}`))
	}))
	defer tokenSrv.Close()

	var gotAuth []string
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	// Install the browser opener the CLI would; restore after.
	prev := auth.SetOAuth2BrowserOpener(func(_ context.Context, authorizationURL string) error {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(authorizationURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	})
	defer func() { auth.SetOAuth2BrowserOpener(prev) }()

	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(WithTokenCache(store, dir))
	req := &Request{
		Method: MethodGet,
		URL:    apiSrv.URL,
		Auth: Auth{
			Type: "oauth2",
			Config: map[string]string{
				"grant_type":        "authorization_code",
				"authorization_url": authSrv.URL,
				"token_url":         tokenSrv.URL,
				"client_id":         "auto-client",
				"client_secret":     "auto-secret",
			},
		},
	}

	// First request: cold cache triggers the browser flow automatically.
	resp, err := client.Execute(context.Background(), req, variables.NewSet())
	if err != nil {
		t.Fatalf("first Execute: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first status = %d, want 200", resp.StatusCode)
	}
	// Second request: reuses the cached token — no browser flow re-run.
	if _, err := client.Execute(context.Background(), req, variables.NewSet()); err != nil {
		t.Fatalf("second Execute: %v", err)
	}

	if tokenCalls != 1 {
		t.Fatalf("token endpoint called %d times across two requests, want 1 (auto-login then reuse)", tokenCalls)
	}
	if len(gotAuth) != 2 || gotAuth[0] != "Bearer auto-tok" || gotAuth[1] != "Bearer auto-tok" {
		t.Fatalf("authorizations = %v, want Bearer auto-tok on both requests", gotAuth)
	}
}

// TestExecuteOAuth2DeviceCode drives the RFC 8628 device flow through the
// engine: cold cache runs the device authorization + poll loop, the token is
// attached as Bearer, and a second request reuses the cached token without
// re-running the flow.
func TestExecuteOAuth2DeviceCode(t *testing.T) {
	var gotAuth []string
	var deviceCalls int
	var tokenCalls int

	deviceSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceCalls++
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if got := r.PostForm.Get("client_id"); got != "dev-client" {
			http.Error(w, "bad client_id", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"device_code":"dev-code","user_code":"AB-1234","verification_uri":"https://idp.example.com/device","interval":1}`))
	}))
	defer deviceSrv.Close()

	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		if tokenCalls == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"authorization_pending"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok-device","token_type":"Bearer","expires_in":3600,"refresh_token":"dev-rt"}`))
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	client := NewClient(WithTokenCache(store, dir))

	req := &Request{
		Method: MethodGet,
		URL:    apiSrv.URL,
		Auth: Auth{
			Type: "oauth2",
			Config: map[string]string{
				"grant_type":               "device_code",
				"device_authorization_url": deviceSrv.URL,
				"token_url":                tokenSrv.URL,
				"client_id":                "dev-client",
				"client_secret":            "dev-secret",
			},
		},
	}

	// First request: cold cache runs the device flow (pending, then grant).
	if _, err := client.Execute(context.Background(), req, variables.NewSet()); err != nil {
		t.Fatalf("first Execute: %v", err)
	}
	// Second request: reuses the cached token — no device flow re-run.
	if _, err := client.Execute(context.Background(), req, variables.NewSet()); err != nil {
		t.Fatalf("second Execute: %v", err)
	}

	if deviceCalls != 1 {
		t.Fatalf("device endpoint called %d times across two requests, want 1", deviceCalls)
	}
	if tokenCalls != 2 {
		t.Fatalf("token endpoint polled %d times, want 2 (pending + grant)", tokenCalls)
	}
	if len(gotAuth) != 2 || gotAuth[0] != "Bearer tok-device" || gotAuth[1] != "Bearer tok-device" {
		t.Fatalf("authorizations = %v, want Bearer tok-device on both requests", gotAuth)
	}
}

func TestExecuteTLSInsecureSkipVerify(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("tls-ok"))
	}))
	defer srv.Close()

	client := NewClient()
	req := &Request{
		Method: MethodGet,
		URL:    srv.URL,
		TLS: &TLSConfig{
			InsecureSkipVerify: true,
		},
	}

	resp, err := client.Execute(context.Background(), req, variables.NewSet())
	if err != nil {
		t.Fatalf("Execute with InsecureSkipVerify failed: %v", err)
	}
	if string(resp.Body) != "tls-ok" {
		t.Fatalf("expected tls-ok body, got %s", string(resp.Body))
	}
}

func TestExecuteWithCustomProxy(t *testing.T) {
	var proxied bool
	proxySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxied = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxied-ok"))
	}))
	defer proxySrv.Close()

	client := NewClient()
	req := &Request{
		Method: MethodGet,
		URL:    "http://target.example.com/api",
		Proxy:  proxySrv.URL,
	}

	resp, err := client.Execute(context.Background(), req, variables.NewSet())
	if err != nil {
		t.Fatalf("Execute with proxy failed: %v", err)
	}
	if !proxied {
		t.Fatal("expected request to hit proxy server")
	}
	if string(resp.Body) != "proxied-ok" {
		t.Fatalf("expected proxied-ok body, got %s", string(resp.Body))
	}
}
