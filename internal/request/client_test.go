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
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/response"
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
