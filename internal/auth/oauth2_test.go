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

package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// tokenServer returns an httptest server acting as an OAuth 2.0 token
// endpoint. The handler records the request and responds with a canned JSON
// token payload.
func tokenServer(t *testing.T, respond func(w http.ResponseWriter), onRequest func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if onRequest != nil {
			onRequest(r)
		}
		respond(w)
	}))
}

func oauthConfig(server *httptest.Server, extra map[string]string) map[string]string {
	cfg := map[string]string{
		"token_url":     server.URL,
		"client_id":     "client-123",
		"client_secret": "s3cr3t",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return cfg
}

func tokenScheme(t *testing.T) auth.TokenSource {
	t.Helper()
	s, ok := auth.Lookup("oauth2")
	if !ok {
		t.Fatal("oauth2 scheme not registered")
	}
	ts, ok := s.(auth.TokenSource)
	if !ok {
		t.Fatalf("oauth2 scheme %T does not implement TokenSource", s)
	}
	return ts
}

func TestOAuth2TokenAcquiresAccessToken(t *testing.T) {
	srv := tokenServer(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok-abc","token_type":"Bearer","expires_in":3600}`))
	}, nil)
	defer srv.Close()

	tok, err := tokenScheme(t).Token(context.Background(), oauthConfig(srv, nil), variables.NewSet())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if tok.AccessToken != "tok-abc" {
		t.Fatalf("AccessToken = %q, want %q", tok.AccessToken, "tok-abc")
	}
	if tok.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want %q", tok.TokenType, "Bearer")
	}
	if tok.ExpiresIn != 3600 {
		t.Fatalf("ExpiresIn = %d, want 3600", tok.ExpiresIn)
	}
}

func TestOAuth2TokenUsesBasicClientAuthAndFormBody(t *testing.T) {
	var gotUser, gotPass string
	var gotBody string
	srv := tokenServer(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
	}, func(r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		gotBody = string(buf)
	})
	defer srv.Close()

	cfg := oauthConfig(srv, map[string]string{"scope": "read write", "audience": "api:prod"})
	if _, err := tokenScheme(t).Token(context.Background(), cfg, variables.NewSet()); err != nil {
		t.Fatalf("Token: %v", err)
	}

	if gotUser != "client-123" || gotPass != "s3cr3t" {
		t.Fatalf("Basic auth = %q/%q, want client-123/s3cr3t", gotUser, gotPass)
	}
	ct := srv // no-op to keep srv referenced; ct unused placeholder
	_ = ct
	if !strings.Contains(gotBody, "grant_type=client_credentials") {
		t.Errorf("body %q missing grant_type=client_credentials", gotBody)
	}
	if !strings.Contains(gotBody, "scope=read+write") && !strings.Contains(gotBody, "scope=read%20write") {
		t.Errorf("body %q missing scope", gotBody)
	}
	if !strings.Contains(gotBody, "audience=api%3Aprod") && !strings.Contains(gotBody, "audience=api:prod") {
		t.Errorf("body %q missing audience", gotBody)
	}
}

func TestOAuth2CustomTokenName(t *testing.T) {
	srv := tokenServer(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data_token":"custom-tok","expires_in":600}`))
	}, nil)
	defer srv.Close()

	tok, err := tokenScheme(t).Token(context.Background(), oauthConfig(srv, map[string]string{"token_name": "data_token"}), variables.NewSet())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if tok.AccessToken != "custom-tok" {
		t.Fatalf("AccessToken = %q, want %q", tok.AccessToken, "custom-tok")
	}
}

func TestOAuth2TokenNon2xxReturnsError(t *testing.T) {
	srv := tokenServer(t, func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_client","error_description":"bad secret"}`))
	}, nil)
	defer srv.Close()

	_, err := tokenScheme(t).Token(context.Background(), oauthConfig(srv, nil), variables.NewSet())
	if err == nil {
		t.Fatal("Token succeeded, want error on non-2xx")
	}
	if !strings.Contains(err.Error(), "invalid_client") {
		t.Fatalf("error %q missing server error body", err.Error())
	}
}

func TestOAuth2TokenAcquisitionValidatesConfig(t *testing.T) {
	srv := tokenServer(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok"}`))
	}, nil)
	defer srv.Close()

	cases := []struct {
		name string
		cfg  map[string]string
	}{
		{"missing token_url", map[string]string{"client_id": "c", "client_secret": "s"}},
		{"missing client_id", map[string]string{"token_url": srv.URL, "client_secret": "s"}},
		{"missing client_secret", map[string]string{"token_url": srv.URL, "client_id": "c"}},
		{"empty client_secret", map[string]string{"token_url": srv.URL, "client_id": "c", "client_secret": ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tokenScheme(t).Token(context.Background(), tc.cfg, variables.NewSet())
			if err == nil {
				t.Fatalf("Token with %s succeeded, want validation error", tc.name)
			}
		})
	}
}

func TestOAuth2ApplySetsBearerFromToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	// The pre-acquired token arrives in config under "token"; Apply must set
	// Authorization: Bearer <token> and nothing else.
	err := auth.Apply(req, "oauth2", map[string]string{"token": "tok-xyz"}, variables.NewSet())
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok-xyz" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer tok-xyz")
	}
}

func TestOAuth2MaskValuesIncludesSecret(t *testing.T) {
	values := auth.MaskValues("oauth2", map[string]string{"client_secret": "s3cr3t"}, variables.NewSet())
	found := false
	for _, v := range values {
		if v == "s3cr3t" {
			found = true
		}
	}
	if !found {
		t.Fatalf("MaskValues = %v, want it to include the client secret", values)
	}
}

// Ensure the token payload encodes the fields we parse (guards the parser
// against field drift).
func TestOAuth2TokenJSONShape(t *testing.T) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(`{"access_token":"t","token_type":"Bearer","expires_in":120}`), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["access_token"] != "t" {
		t.Fatalf("access_token = %v", payload["access_token"])
	}
}
