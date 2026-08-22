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

package auth_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// driveBrowser simulates a user's browser: it follows the authorization URL
// (the fake provider redirects to the loopback callback) and returns once the
// callback page has been served.
func driveBrowser(authorizationURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(authorizationURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// authCodeProvider returns a fake provider: an authorization endpoint that
// redirects to the flow's callback with a canned code and the echoed state,
// plus a token endpoint that records the exchange and responds with payload.
func authCodeProvider(t *testing.T, payload string, onToken func(r *http.Request)) (*httptest.Server, *httptest.Server) {
	t.Helper()
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		callback := r.URL.Query().Get("redirect_uri")
		cb, err := url.Parse(callback)
		if err != nil {
			http.Error(w, "bad redirect_uri", http.StatusBadRequest)
			return
		}
		q := cb.Query()
		q.Set("code", "auth-code-1")
		q.Set("state", state)
		cb.RawQuery = q.Encode()
		http.Redirect(w, r, cb.String(), http.StatusFound)
	}))
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if onToken != nil {
			onToken(r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(payload))
	}))
	t.Cleanup(authSrv.Close)
	t.Cleanup(tokenSrv.Close)
	return authSrv, tokenSrv
}

func authCodeConfig(authSrv, tokenSrv *httptest.Server, extra map[string]string) map[string]string {
	cfg := map[string]string{
		"authorization_url": authSrv.URL,
		"token_url":         tokenSrv.URL,
		"client_id":         "client-123",
		"client_secret":     "s3cr3t",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return cfg
}

func TestPKCEChallengeMatchesRFC7636Vector(t *testing.T) {
	// RFC 7636 Appendix B test vector.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got := auth.PKCEChallenge(verifier); got != want {
		t.Fatalf("PKCEChallenge = %q, want %q", got, want)
	}
}

func TestPKCEVerifierWithinRange(t *testing.T) {
	v, err := auth.PKCEVerifier()
	if err != nil {
		t.Fatalf("PKCEVerifier: %v", err)
	}
	if len(v) < 43 || len(v) > 128 {
		t.Fatalf("verifier length %d outside RFC 7636 43-128 range", len(v))
	}
	if got := auth.PKCEChallenge(v); got == "" {
		t.Fatal("challenge is empty")
	}
}

func TestBuildAuthorizationURLParams(t *testing.T) {
	cfg := map[string]string{
		"authorization_url": "https://idp.example.com/authorize",
		"client_id":         "client-123",
		"scope":             "read write",
	}
	verifier, err := auth.PKCEVerifier()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := auth.BuildAuthorizationURL(cfg, variables.NewSet(), "http://127.0.0.1:0/callback", verifier, "state-1")
	if err != nil {
		t.Fatalf("BuildAuthorizationURL: %v", err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	q := u.Query()
	want := map[string]string{
		"response_type":         "code",
		"client_id":             "client-123",
		"redirect_uri":          "http://127.0.0.1:0/callback",
		"code_challenge":        auth.PKCEChallenge(verifier),
		"code_challenge_method": "S256",
		"state":                 "state-1",
		"scope":                 "read write",
	}
	for k, v := range want {
		if got := q.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestStartAuthorizationFlowValidatesConfig(t *testing.T) {
	_, tokenSrv := authCodeProvider(t, `{"access_token":"t","expires_in":3600}`, nil)
	base := authCodeConfig(&httptest.Server{URL: "https://idp.example.com/authorize"}, tokenSrv, nil)

	cases := []struct {
		name string
		cfg  map[string]string
	}{
		{"missing authorization_url", drop(base, "authorization_url")},
		{"missing token_url", drop(base, "token_url")},
		{"missing client_id", drop(base, "client_id")},
		{"missing client_secret", drop(base, "client_secret")},
		{"non-loopback redirect_uri", withVal(base, "redirect_uri", "https://app.example.com/callback")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flow, err := auth.StartAuthorizationFlow(tc.cfg, variables.NewSet())
			if flow != nil {
				flow.Close()
			}
			if err == nil {
				t.Fatalf("StartAuthorizationFlow succeeded, want error")
			}
		})
	}
}

// drop returns a copy of cfg without key.
func drop(cfg map[string]string, key string) map[string]string {
	out := make(map[string]string, len(cfg))
	for k, v := range cfg {
		if k != key {
			out[k] = v
		}
	}
	return out
}

// withVal returns a copy of cfg with key set to val.
func withVal(cfg map[string]string, key, val string) map[string]string {
	out := make(map[string]string, len(cfg)+1)
	for k, v := range cfg {
		out[k] = v
	}
	out[key] = val
	return out
}

func TestAuthorizationCodeSourceEndToEnd(t *testing.T) {
	authSrv, tokenSrv := authCodeProvider(t, `{"access_token":"tok-auth","token_type":"Bearer","expires_in":3600,"refresh_token":"rt-1"}`, nil)
	src := &auth.AuthorizationCodeSource{
		Open: func(_ context.Context, u string) error { return driveBrowser(u) },
	}

	tok, err := src.Token(context.Background(), authCodeConfig(authSrv, tokenSrv, nil), variables.NewSet())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if tok.AccessToken != "tok-auth" {
		t.Fatalf("AccessToken = %q, want %q", tok.AccessToken, "tok-auth")
	}
	if tok.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want %q", tok.TokenType, "Bearer")
	}
	if tok.RefreshToken != "rt-1" {
		t.Fatalf("RefreshToken = %q, want %q", tok.RefreshToken, "rt-1")
	}
	if tok.Expiry.IsZero() {
		t.Fatal("Expiry is zero, want derived from expires_in")
	}
}

func TestAuthorizationCodeFlowExchangeBody(t *testing.T) {
	var gotUser, gotPass string
	var gotForm url.Values
	authSrv, tokenSrv := authCodeProvider(t, `{"access_token":"t","expires_in":3600}`, func(r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		body, _ := io.ReadAll(r.Body)
		gotForm, _ = url.ParseQuery(string(body))
	})
	src := &auth.AuthorizationCodeSource{
		Open: func(_ context.Context, u string) error { return driveBrowser(u) },
	}

	if _, err := src.Token(context.Background(), authCodeConfig(authSrv, tokenSrv, nil), variables.NewSet()); err != nil {
		t.Fatalf("Token: %v", err)
	}
	if gotUser != "client-123" || gotPass != "s3cr3t" {
		t.Fatalf("Basic auth = %q/%q, want client-123/s3cr3t", gotUser, gotPass)
	}
	for key, want := range map[string]string{
		"grant_type":    "authorization_code",
		"code":          "auth-code-1",
		"client_id":     "client-123",
		"code_verifier": "",
	} {
		got := gotForm.Get(key)
		if key == "code_verifier" {
			if len(got) < 43 || len(got) > 128 {
				t.Errorf("code_verifier length %d outside 43-128", len(got))
			}
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if gotForm.Get("redirect_uri") == "" {
		t.Error("redirect_uri missing from exchange body")
	}
}

func TestAuthorizationCodeFlowStateMismatch(t *testing.T) {
	authSrv, _ := authCodeProvider(t, `{"access_token":"t"}`, nil)
	cfg := authCodeConfig(authSrv, &httptest.Server{URL: "https://token.example.com"}, nil)

	flow, err := auth.StartAuthorizationFlow(cfg, variables.NewSet())
	if err != nil {
		t.Fatalf("StartAuthorizationFlow: %v", err)
	}
	defer flow.Close()

	authURL, err := url.Parse(flow.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	cb, err := url.Parse(authURL.Query().Get("redirect_uri"))
	if err != nil {
		t.Fatal(err)
	}
	cb.RawQuery = "code=x&state=wrong-state"

	resp, err := http.Get(cb.String())
	if err != nil {
		t.Fatalf("callback GET: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("callback status = %d, want 400", resp.StatusCode)
	}

	if _, err := flow.WaitCode(context.Background()); err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("WaitCode error = %v, want state mismatch", err)
	}
}

func TestAuthorizationCodeFlowErrorCallback(t *testing.T) {
	authSrv, _ := authCodeProvider(t, `{"access_token":"t"}`, nil)
	cfg := authCodeConfig(authSrv, &httptest.Server{URL: "https://token.example.com"}, nil)

	flow, err := auth.StartAuthorizationFlow(cfg, variables.NewSet())
	if err != nil {
		t.Fatalf("StartAuthorizationFlow: %v", err)
	}
	defer flow.Close()

	authURL, err := url.Parse(flow.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	cb, err := url.Parse(authURL.Query().Get("redirect_uri"))
	if err != nil {
		t.Fatal(err)
	}
	state := authURL.Query().Get("state")
	cb.RawQuery = "error=access_denied&state=" + state

	resp, err := http.Get(cb.String())
	if err != nil {
		t.Fatalf("callback GET: %v", err)
	}
	resp.Body.Close()

	if _, err := flow.WaitCode(context.Background()); err == nil || !strings.Contains(err.Error(), "access_denied") {
		t.Fatalf("WaitCode error = %v, want access_denied", err)
	}
}

func TestOAuth2SchemeRejectsUnknownGrantType(t *testing.T) {
	srv := tokenServer(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"t"}`))
	}, nil)
	defer srv.Close()

	cfg := oauthConfig(srv, map[string]string{"grant_type": "password"})
	_, err := tokenScheme(t).Token(context.Background(), cfg, variables.NewSet())
	if err == nil || !strings.Contains(err.Error(), `unsupported grant_type "password"`) {
		t.Fatalf("Token error = %v, want unsupported grant_type", err)
	}
}

// TestOAuth2SchemeAuthorizationCodeDispatchesToBrowserFlow proves
// grant_type authorization_code dispatches to the browser flow (rather than
// the client-credentials POST). Without a configured browser opener the flow
// must fail fast with a clear error instead of hanging or hitting the token
// endpoint.
func TestOAuth2SchemeAuthorizationCodeDispatchesToBrowserFlow(t *testing.T) {
	authSrv, tokenSrv := authCodeProvider(t, `{"access_token":"t"}`, nil)
	cfg := authCodeConfig(authSrv, tokenSrv, map[string]string{"grant_type": "authorization_code"})

	_, err := tokenScheme(t).Token(context.Background(), cfg, variables.NewSet())
	if err == nil || !strings.Contains(err.Error(), "browser opener") {
		t.Fatalf("Token error = %v, want browser opener error (dispatch to browser flow)", err)
	}
}

func TestOAuth2TokenParsesRefreshToken(t *testing.T) {
	srv := tokenServer(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"t","token_type":"Bearer","expires_in":120,"refresh_token":"rt-x"}`))
	}, nil)
	defer srv.Close()

	tok, err := tokenScheme(t).Token(context.Background(), oauthConfig(srv, nil), variables.NewSet())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if tok.RefreshToken != "rt-x" {
		t.Fatalf("RefreshToken = %q, want %q", tok.RefreshToken, "rt-x")
	}
}

func TestStartAuthorizationFlowCustomSchemeUnregistered(t *testing.T) {
	_, tokenSrv := authCodeProvider(t, `{"access_token":"t","expires_in":3600}`, nil)
	cfg := authCodeConfig(&httptest.Server{URL: "https://idp.example.com/authorize"}, tokenSrv, nil)
	cfg["redirect_uri"] = "reqly://callback"

	flow, err := auth.StartAuthorizationFlow(cfg, variables.NewSet())
	if flow != nil {
		flow.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "no registered receiver") {
		t.Fatalf("err = %v, want no-registered-receiver error", err)
	}
}

func TestAuthorizationCodeSourceCustomScheme(t *testing.T) {
	unregister := auth.RegisterCustomSchemeReceiver("reqly")
	t.Cleanup(unregister)

	var gotRedirect string
	_, tokenSrv := authCodeProvider(t, `{"access_token":"deep-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"deep-rt"}`, func(r *http.Request) {
		if err := r.ParseForm(); err == nil {
			gotRedirect = r.PostForm.Get("redirect_uri")
		}
	})
	cfg := authCodeConfig(&httptest.Server{URL: "https://idp.example.com/authorize"}, tokenSrv, nil)
	cfg["redirect_uri"] = "reqly://callback"

	src := &auth.AuthorizationCodeSource{
		Open: func(_ context.Context, authorizationURL string) error {
			u, err := url.Parse(authorizationURL)
			if err != nil {
				return err
			}
			cb := "reqly://callback?code=deep-code&state=" + url.QueryEscape(u.Query().Get("state"))
			return auth.DeliverCustomSchemeCallback(cb)
		},
	}
	tok, err := src.Token(context.Background(), cfg, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "deep-tok" || tok.RefreshToken != "deep-rt" {
		t.Fatalf("tok = %+v", tok)
	}
	if gotRedirect != "reqly://callback" {
		t.Fatalf("exchange redirect_uri = %q, want reqly://callback", gotRedirect)
	}

	// One-shot: the flow was removed from the registry at delivery.
	if err := auth.DeliverCustomSchemeCallback("reqly://callback?code=again&state=x"); err == nil || !strings.Contains(err.Error(), "no authorization flow waiting") {
		t.Fatalf("second delivery err = %v, want no-waiting-flow (one-shot)", err)
	}
}

func TestCustomSchemeCallbackStateMismatch(t *testing.T) {
	unregister := auth.RegisterCustomSchemeReceiver("reqly")
	t.Cleanup(unregister)

	_, tokenSrv := authCodeProvider(t, `{"access_token":"t","expires_in":3600}`, nil)
	cfg := authCodeConfig(&httptest.Server{URL: "https://idp.example.com/authorize"}, tokenSrv, nil)
	cfg["redirect_uri"] = "reqly://callback"

	flow, err := auth.StartAuthorizationFlow(cfg, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()

	err = auth.DeliverCustomSchemeCallback("reqly://callback?code=deep-code&state=wrong-state")
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("deliver err = %v, want state mismatch", err)
	}
	if _, err := flow.WaitCode(context.Background()); err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("WaitCode err = %v, want state mismatch", err)
	}
}

func TestDeliverCustomSchemeCallbackNoFlow(t *testing.T) {
	unregister := auth.RegisterCustomSchemeReceiver("reqly")
	t.Cleanup(unregister)

	err := auth.DeliverCustomSchemeCallback("reqly://callback?code=x&state=y")
	if err == nil || !strings.Contains(err.Error(), "no authorization flow waiting") {
		t.Fatalf("err = %v, want no-waiting-flow error", err)
	}
}
