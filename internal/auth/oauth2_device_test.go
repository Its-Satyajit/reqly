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

package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

// deviceConfig returns a minimal device-flow config with live test endpoints.
func deviceConfig(t *testing.T, deviceURL, tokenURL string) map[string]string {
	t.Helper()
	return map[string]string{
		"grant_type":               "device_code",
		"device_authorization_url": deviceURL,
		"token_url":                tokenURL,
		"client_id":                "dev-client",
		"client_secret":            "dev-secret",
		"scope":                    "read write",
	}
}

// fakeDeviceEndpoint asserts the device-authorization request shape and
// returns a scripted response.
func fakeDeviceEndpoint(t *testing.T, check func(r *http.Request, form url.Values), body string, status int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "dev-client" || pass != "dev-secret" {
			http.Error(w, "bad basic auth", http.StatusUnauthorized)
			return
		}
		if check != nil {
			check(r, r.PostForm)
		}
		if status != http.StatusOK {
			http.Error(w, body, status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// scriptedTokenEndpoint returns a token endpoint that replays responses in
// order (each one 200 unless given a status), recording poll timestamps.
func scriptedTokenEndpoint(t *testing.T, responses []string, statuses []int, times *[]time.Time, mu *sync.Mutex) *httptest.Server {
	t.Helper()
	var i int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mu != nil {
			mu.Lock()
			defer mu.Unlock()
		}
		if times != nil {
			*times = append(*times, time.Now())
		}
		idx := i
		if idx < len(responses)-1 {
			i++
		}
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if got := r.PostForm.Get("grant_type"); got != "urn:ietf:params:oauth:grant-type:device_code" {
			http.Error(w, "bad grant_type "+got, http.StatusBadRequest)
			return
		}
		if got := r.PostForm.Get("device_code"); got != "dev-code" {
			http.Error(w, "bad device_code "+got, http.StatusBadRequest)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "dev-client" || pass != "dev-secret" {
			http.Error(w, "bad basic auth", http.StatusUnauthorized)
			return
		}
		status := http.StatusOK
		if statuses != nil && idx < len(statuses) {
			status = statuses[idx]
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(responses[idx]))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestStartDeviceAuthorization(t *testing.T) {
	var gotForm url.Values
	srv := fakeDeviceEndpoint(t, func(_ *http.Request, form url.Values) {
		gotForm = form
	}, `{"device_code":"dev-code","user_code":"AB-1234","verification_uri":"https://idp.example.com/device","verification_uri_complete":"https://idp.example.com/device?user_code=AB-1234","interval":7}`, http.StatusOK)

	da, err := StartDeviceAuthorization(context.Background(), deviceConfig(t, srv.URL, ""), variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if da.DeviceCode != "dev-code" || da.UserCode != "AB-1234" {
		t.Fatalf("da = %+v", da)
	}
	if da.VerificationURI != "https://idp.example.com/device" {
		t.Fatalf("VerificationURI = %q", da.VerificationURI)
	}
	if da.VerificationURIComplete != "https://idp.example.com/device?user_code=AB-1234" {
		t.Fatalf("VerificationURIComplete = %q", da.VerificationURIComplete)
	}
	if da.Interval != 7 {
		t.Fatalf("Interval = %d, want 7", da.Interval)
	}

	// RFC 8628 §3.1: client_id, scope, audience in the form.
	if gotForm.Get("client_id") != "dev-client" {
		t.Fatalf("client_id = %q", gotForm.Get("client_id"))
	}
	if gotForm.Get("scope") != "read write" {
		t.Fatalf("scope = %q", gotForm.Get("scope"))
	}
}

func TestStartDeviceAuthorizationValidation(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(tokenSrv.Close)

	base := map[string]string{"client_id": "c", "client_secret": "s", "token_url": tokenSrv.URL}
	for _, missing := range []string{"device_authorization_url", "client_id", "client_secret"} {
		cfg := map[string]string{}
		for k, v := range base {
			cfg[k] = v
		}
		cfg["device_authorization_url"] = tokenSrv.URL
		delete(cfg, missing)
		_, err := StartDeviceAuthorization(context.Background(), cfg, variables.NewSet())
		if err == nil || !strings.Contains(err.Error(), missing) {
			t.Fatalf("missing %s: err = %v, want validation error naming %q", missing, err, missing)
		}
	}
}

func TestStartDeviceAuthorizationNon200(t *testing.T) {
	srv := fakeDeviceEndpoint(t, nil, "rate limited", http.StatusTooManyRequests)
	_, err := StartDeviceAuthorization(context.Background(), deviceConfig(t, srv.URL, ""), variables.NewSet())
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("err = %v, want 429 error", err)
	}
}

func TestStartDeviceAuthorizationMissingFields(t *testing.T) {
	srv := fakeDeviceEndpoint(t, nil, `{"user_code":"AB-1234"}`, http.StatusOK)
	_, err := StartDeviceAuthorization(context.Background(), deviceConfig(t, srv.URL, ""), variables.NewSet())
	if err == nil || !strings.Contains(err.Error(), "missing device_code/user_code/verification_uri") {
		t.Fatalf("err = %v, want missing-fields error", err)
	}
}

func TestPollDeviceTokenImmediateGrant(t *testing.T) {
	body := `{"access_token":"dev-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"dev-rt"}`
	srv := scriptedTokenEndpoint(t, []string{body}, nil, nil, nil)

	tok, err := PollDeviceToken(context.Background(), deviceConfig(t, "", srv.URL), variables.NewSet(), "dev-code", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "dev-tok" || tok.RefreshToken != "dev-rt" {
		t.Fatalf("tok = %+v", tok)
	}
	if tok.Expiry.IsZero() {
		t.Fatal("Expiry not set")
	}
}

func TestPollDeviceTokenPendingThenGrant(t *testing.T) {
	srv := scriptedTokenEndpoint(t, []string{
		`{"error":"authorization_pending"}`,
		`{"access_token":"dev-tok","token_type":"Bearer","expires_in":3600}`,
	}, []int{http.StatusBadRequest, http.StatusOK}, nil, nil)

	start := time.Now()
	tok, err := PollDeviceToken(context.Background(), deviceConfig(t, "", srv.URL), variables.NewSet(), "dev-code", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "dev-tok" {
		t.Fatalf("AccessToken = %q", tok.AccessToken)
	}
	// Two polls at interval=1s: at least ~2s elapse.
	if elapsed := time.Since(start); elapsed < 2*time.Second-100*time.Millisecond {
		t.Fatalf("elapsed = %v, want >= ~2s (interval honored)", elapsed)
	}
}

func TestPollDeviceTokenSlowDownRetriesAndAddsInterval(t *testing.T) {
	oldExtra := devicePollSlowDownExtra
	devicePollSlowDownExtra = 150 * time.Millisecond
	t.Cleanup(func() { devicePollSlowDownExtra = oldExtra })

	var mu sync.Mutex
	var polls []time.Time
	srv := scriptedTokenEndpoint(t, []string{
		`{"error":"slow_down"}`,
		`{"access_token":"dev-tok","token_type":"Bearer","expires_in":3600}`,
	}, []int{http.StatusBadRequest, http.StatusOK}, &polls, &mu)

	tok, err := PollDeviceToken(context.Background(), deviceConfig(t, "", srv.URL), variables.NewSet(), "dev-code", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "dev-tok" {
		t.Fatalf("AccessToken = %q", tok.AccessToken)
	}
	// The second poll must wait interval (1s) + the slow_down extra, not
	// just the interval. Floor assertion — never flakes on fast machines.
	mu.Lock()
	defer mu.Unlock()
	if len(polls) != 2 {
		t.Fatalf("polled %d times, want 2", len(polls))
	}
	if delta := polls[1].Sub(polls[0]); delta < time.Second+devicePollSlowDownExtra-100*time.Millisecond {
		t.Fatalf("second poll %v after slow_down, want >= ~1.15s", delta)
	}
}

func TestPollDeviceTokenTerminalErrors(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		status   int
		wantPart string
	}{
		{"access_denied", `{"error":"access_denied"}`, http.StatusBadRequest, "denied by the user"},
		{"expired_token", `{"error":"expired_token"}`, http.StatusBadRequest, "device code expired"},
		{"unknown_error", `{"error":"unsupported_grant_type"}`, http.StatusBadRequest, "unsupported_grant_type"},
		{"plain_non200", "internal error", http.StatusInternalServerError, "500"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := scriptedTokenEndpoint(t, []string{tc.body}, []int{tc.status}, nil, nil)
			_, err := PollDeviceToken(context.Background(), deviceConfig(t, "", srv.URL), variables.NewSet(), "dev-code", 1, nil)
			if err == nil {
				t.Fatal("err = nil, want terminal error")
			}
			if !strings.Contains(err.Error(), tc.wantPart) {
				t.Fatalf("err = %q, want it to contain %q", err, tc.wantPart)
			}
		})
	}
}

func TestPollDeviceTokenContextCancel(t *testing.T) {
	srv := scriptedTokenEndpoint(t, []string{`{"access_token":"dev-tok"}`}, nil, nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := PollDeviceToken(ctx, deviceConfig(t, "", srv.URL), variables.NewSet(), "dev-code", 1, nil)
	if err == nil || !strings.Contains(err.Error(), "context") {
		t.Fatalf("err = %v, want context cancellation error", err)
	}
}

func TestDeviceCodeSourceToken(t *testing.T) {
	deviceSrv := fakeDeviceEndpoint(t, nil,
		`{"device_code":"dev-code","user_code":"AB-1234","verification_uri":"https://idp.example.com/device","verification_uri_complete":"https://idp.example.com/device?user_code=AB-1234","interval":1}`, http.StatusOK)
	tokenSrv := scriptedTokenEndpoint(t, []string{
		`{"error":"authorization_pending"}`,
		`{"access_token":"dev-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"dev-rt"}`,
	}, []int{http.StatusBadRequest, http.StatusOK}, nil, nil)

	var lines []string
	src := &DeviceCodeSource{Status: func(line string) { lines = append(lines, line) }}
	tok, err := src.Token(context.Background(), deviceConfig(t, deviceSrv.URL, tokenSrv.URL), variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "dev-tok" || tok.RefreshToken != "dev-rt" {
		t.Fatalf("tok = %+v", tok)
	}

	// The user must see the verification URI and the code; the complete URI
	// is preferred when the provider sends it.
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"https://idp.example.com/device?user_code=AB-1234", "AB-1234", "waiting for authorization"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("status lines missing %q, got:\n%s", want, joined)
		}
	}
}

func TestOAuth2SchemeDispatchesDeviceCode(t *testing.T) {
	deviceSrv := fakeDeviceEndpoint(t, nil,
		`{"device_code":"dev-code","user_code":"AB-1234","verification_uri":"https://idp.example.com/device","interval":1}`, http.StatusOK)
	tokenSrv := scriptedTokenEndpoint(t, []string{
		`{"access_token":"dev-tok","token_type":"Bearer","expires_in":3600}`,
	}, nil, nil, nil)

	cfg := deviceConfig(t, deviceSrv.URL, tokenSrv.URL)
	tok, err := (oauth2Scheme{}).Token(context.Background(), cfg, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "dev-tok" {
		t.Fatalf("AccessToken = %q", tok.AccessToken)
	}
}
