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
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// fakeDeviceAuthEndpoints returns device + token endpoints for tests. The
// token endpoint answers authorization_pending once, then grants.
func fakeDeviceAuthEndpoints(t *testing.T) (*httptest.Server, *httptest.Server) {
	t.Helper()
	deviceSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"device_code":"dev-code","user_code":"AB-1234","verification_uri":"https://idp.example.com/device","interval":1}`))
	}))
	var polls int
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		polls++
		if polls == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"authorization_pending"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"core-device-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"core-device-rt"}`))
	}))
	t.Cleanup(deviceSrv.Close)
	t.Cleanup(tokenSrv.Close)
	return deviceSrv, tokenSrv
}

// authServiceFixture builds an AuthService over a fresh file store in a temp
// workspace root.
func authServiceFixture(t *testing.T) (*AuthService, string) {
	t.Helper()
	root := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	return NewAuthService(store, root), root
}

func TestAuthServiceLoginDeviceFlowAndStatus(t *testing.T) {
	svc, _ := authServiceFixture(t)
	deviceSrv, tokenSrv := fakeDeviceAuthEndpoints(t)

	tok, err := svc.Login(context.Background(), LoginRequest{
		Config: map[string]string{
			"grant_type":               "device_code",
			"device_authorization_url": deviceSrv.URL,
			"token_url":                tokenSrv.URL,
			"client_id":                "core-client",
			"client_secret":            "core-secret",
		},
		Flow: "device_code",
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "core-device-tok" || tok.RefreshToken != "core-device-rt" {
		t.Fatalf("tok = %+v", tok)
	}

	status, err := svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	if len(status) != 1 {
		t.Fatalf("status = %d entries, want 1", len(status))
	}
	s := status[0]
	if s.GrantType != "device_code" || !s.HasRefresh || s.State != "cached" {
		t.Fatalf("status = %+v", s)
	}
	if strings.Contains(s.AccessToken, "core-device-tok") || s.AccessToken == "core-device-tok" {
		t.Fatalf("status leaked the full token: %q", s.AccessToken)
	}

	cleared, err := svc.Logout()
	if err != nil {
		t.Fatal(err)
	}
	if cleared != 1 {
		t.Fatalf("cleared %d, want 1", cleared)
	}
	status, err = svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	if len(status) != 0 {
		t.Fatalf("status after logout = %d entries, want 0", len(status))
	}
}

func TestAuthServiceLoginAutoDetectsDeviceFlow(t *testing.T) {
	svc, _ := authServiceFixture(t)
	deviceSrv, tokenSrv := fakeDeviceAuthEndpoints(t)

	tok, err := svc.Login(context.Background(), LoginRequest{
		Config: map[string]string{
			"device_authorization_url": deviceSrv.URL,
			"token_url":                tokenSrv.URL,
			"client_id":                "core-client",
			"client_secret":            "core-secret",
		},
		// Flow omitted: auto-detect from the config.
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "core-device-tok" {
		t.Fatalf("AccessToken = %q", tok.AccessToken)
	}
}

func TestAuthServiceLoginUnknownFlow(t *testing.T) {
	svc, _ := authServiceFixture(t)
	_, err := svc.Login(context.Background(), LoginRequest{
		Config: map[string]string{"client_id": "c"},
		Flow:   "bogus",
	})
	if err == nil || !strings.Contains(err.Error(), "unknown login flow") {
		t.Fatalf("err = %v, want unknown login flow", err)
	}
}

func TestAuthServiceStatusEmpty(t *testing.T) {
	svc, _ := authServiceFixture(t)
	status, err := svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	if len(status) != 0 {
		t.Fatalf("status = %d entries, want 0", len(status))
	}
	cleared, err := svc.Logout()
	if err != nil {
		t.Fatal(err)
	}
	if cleared != 0 {
		t.Fatalf("cleared %d on empty store, want 0", cleared)
	}
}

// TestAuthServiceLoginThenRequestReusesToken proves the desktop's login →
// request path: after Login, a cached request client reuses the token
// without re-running the flow.
func TestAuthServiceLoginThenRequestReusesToken(t *testing.T) {
	svc, root := authServiceFixture(t)
	deviceSrv, tokenSrv := fakeDeviceAuthEndpoints(t)
	cfg := map[string]string{
		"grant_type":               "device_code",
		"device_authorization_url": deviceSrv.URL,
		"token_url":                tokenSrv.URL,
		"client_id":                "core-client",
		"client_secret":            "core-secret",
	}
	if _, err := svc.Login(context.Background(), LoginRequest{Config: cfg, Flow: "device_code"}); err != nil {
		t.Fatal(err)
	}

	var gotAuth []string
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	svc2 := NewCachedRequestService(store, root)
	req := request.Request{
		Method: request.MethodGet,
		URL:    apiSrv.URL,
		Auth: request.Auth{
			Type:   "oauth2",
			Config: cfg,
		},
	}
	for i := 0; i < 2; i++ {
		if _, err := svc2.Send(context.Background(), req); err != nil {
			t.Fatalf("Send %d: %v", i, err)
		}
	}
	if len(gotAuth) != 2 || gotAuth[0] != "Bearer core-device-tok" || gotAuth[1] != "Bearer core-device-tok" {
		t.Fatalf("authorizations = %v, want Bearer core-device-tok on both requests", gotAuth)
	}
}

// TestAuthServiceLoginAuthCodeFlow runs the authorization_code grant with a
// caller-supplied Open that drives the callback like a browser would.
func TestAuthServiceLoginAuthCodeFlow(t *testing.T) {
	svc, _ := authServiceFixture(t)

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		cb := r.URL.Query().Get("redirect_uri")
		q := strings.SplitN(cb, "?", 2)[0] + "?code=core-code&state=" + state
		http.Redirect(w, r, q, http.StatusFound)
	}))
	defer authSrv.Close()
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"core-authcode-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"core-authcode-rt"}`))
	}))
	defer tokenSrv.Close()

	tok, err := svc.Login(context.Background(), LoginRequest{
		Config: map[string]string{
			"authorization_url": authSrv.URL,
			"token_url":         tokenSrv.URL,
			"client_id":         "core-client",
			"client_secret":     "core-secret",
		},
		Flow: "authorization_code",
		Open: func(_ context.Context, authorizationURL string) error {
			client := &http.Client{}
			resp, err := client.Get(authorizationURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "core-authcode-tok" || tok.RefreshToken != "core-authcode-rt" {
		t.Fatalf("tok = %+v", tok)
	}

	status, err := svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	if len(status) != 1 || status[0].GrantType != "authorization_code" || !status[0].HasRefresh {
		t.Fatalf("status = %+v", status)
	}
}
