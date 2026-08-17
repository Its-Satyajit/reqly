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

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// writeCachedToken seeds a token store in the workspace root with one cached
// token for a synthetic config.
func writeCachedToken(t *testing.T, root, token string, expiry time.Time) {
	t.Helper()
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := json.Marshal(map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"endpoint":     "https://auth.example.com/token",
		"expiry":       expiry,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set("some-workspace:some-config", string(blob)); err != nil {
		t.Fatal(err)
	}
}

// chdirWorkspace moves the process into root for the duration of the test.
func chdirWorkspace(t *testing.T, root string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAuthStatusShowsCachedToken(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(1*time.Hour))
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{
		"https://auth.example.com/token",
		"very",   // masked prefix visible
		"oken",   // masked suffix visible
		"cached", // state
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	if strings.Contains(output, "very-secret-access-token") {
		t.Fatalf("status leaked the full token:\n%s", output)
	}
	if !strings.Contains(output, "****************") {
		t.Fatalf("expected masked stars, got:\n%s", output)
	}
}

func TestAuthStatusEmpty(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "no cached tokens") {
		t.Fatalf("expected 'no cached tokens', got:\n%s", out.String())
	}
}

func TestAuthStatusExpired(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(-1*time.Hour))
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "expired") {
		t.Fatalf("expected 'expired' state, got:\n%s", out.String())
	}
}

func TestAuthLogoutClearsTokens(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(1*time.Hour))
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "logout"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "cleared 1 cached token(s)") {
		t.Fatalf("expected 'cleared 1 cached token(s)', got:\n%s", out.String())
	}

	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := store.Keys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 0 {
		t.Fatalf("store still has keys after logout: %v", keys)
	}
}

func TestAuthLogoutEmpty(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "logout"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "cleared 0 cached token(s)") {
		t.Fatalf("expected 'cleared 0 cached token(s)', got:\n%s", out.String())
	}
}

// fakeAuthCodeProvider returns a fake provider: an authorization endpoint
// that redirects to the flow's callback with a code and the echoed state,
// plus a token endpoint that issues an access + refresh token.
func fakeAuthCodeProvider(t *testing.T) (*httptest.Server, *httptest.Server) {
	t.Helper()
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		cb, err := url.Parse(r.URL.Query().Get("redirect_uri"))
		if err != nil {
			http.Error(w, "bad redirect_uri", http.StatusBadRequest)
			return
		}
		q := cb.Query()
		q.Set("code", "cli-code")
		q.Set("state", state)
		cb.RawQuery = q.Encode()
		http.Redirect(w, r, cb.String(), http.StatusFound)
	}))
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok-auth-code","token_type":"Bearer","expires_in":3600,"refresh_token":"cli-rt"}`))
	}))
	t.Cleanup(authSrv.Close)
	t.Cleanup(tokenSrv.Close)
	return authSrv, tokenSrv
}

// writeAuthConfig writes a flat OAuth config file under dir and returns its
// path.
func writeAuthConfig(t *testing.T, dir string, cfg map[string]string) string {
	t.Helper()
	var b strings.Builder
	for _, k := range []string{"authorization_url", "device_authorization_url", "token_url", "client_id", "client_secret", "redirect_uri", "scope"} {
		if v, ok := cfg[k]; ok {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	path := filepath.Join(dir, "auth-config.yaml")
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// fakeLaunchBrowser drives the callback like a real browser would.
func fakeLaunchBrowser(authorizationURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(authorizationURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func TestAuthLoginCompletesFlowAndPersists(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	authSrv, tokenSrv := fakeAuthCodeProvider(t)
	cfgPath := writeAuthConfig(t, root, map[string]string{
		"authorization_url": authSrv.URL,
		"token_url":         tokenSrv.URL,
		"client_id":         "cli-client",
		"client_secret":     "cli-secret",
	})

	oldBrowser := launchBrowser
	launchBrowser = fakeLaunchBrowser
	t.Cleanup(func() { launchBrowser = oldBrowser })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "login", cfgPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{"login complete", "tok", "yes"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	if strings.Contains(output, "cli-secret") || strings.Contains(output, "tok-auth-code") {
		t.Fatalf("login output leaked a secret:\n%s", output)
	}

	// The token must be cached with the refresh token and grant type so
	// later requests reuse it and status reports it.
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := store.Keys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 {
		t.Fatalf("cached %d tokens, want 1", len(keys))
	}
	raw, err := store.Get(keys[0])
	if err != nil {
		t.Fatal(err)
	}
	tok, err := auth.ParseCachedToken(raw)
	if err != nil {
		t.Fatalf("ParseCachedToken: %v", err)
	}
	if tok.RefreshToken != "cli-rt" {
		t.Fatalf("RefreshToken = %q, want cli-rt", tok.RefreshToken)
	}
	if tok.GrantType != "authorization_code" {
		t.Fatalf("GrantType = %q, want authorization_code", tok.GrantType)
	}
	if tok.AccessToken != "tok-auth-code" {
		t.Fatalf("AccessToken = %q, want tok-auth-code", tok.AccessToken)
	}
}

func TestAuthLoginTimeout(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	authSrv, tokenSrv := fakeAuthCodeProvider(t)
	cfgPath := writeAuthConfig(t, root, map[string]string{
		"authorization_url": authSrv.URL,
		"token_url":         tokenSrv.URL,
		"client_id":         "cli-client",
		"client_secret":     "cli-secret",
	})

	oldBrowser := launchBrowser
	launchBrowser = func(string) error { return nil } // never hits the callback
	t.Cleanup(func() { launchBrowser = oldBrowser })

	oldTimeout := authLoginTimeoutSeconds
	authLoginTimeoutSeconds = 1
	t.Cleanup(func() { authLoginTimeoutSeconds = oldTimeout })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "login", cfgPath})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "authorization callback") {
		t.Fatalf("err = %v, want authorization callback wait failure", err)
	}
}

func TestAuthLoginValidatesConfig(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	cfgPath := writeAuthConfig(t, root, map[string]string{
		"token_url":     "https://token.example.com",
		"client_id":     "cli-client",
		"client_secret": "cli-secret",
		// authorization_url intentionally missing
	})

	called := false
	oldBrowser := launchBrowser
	launchBrowser = func(string) error { called = true; return nil }
	t.Cleanup(func() { launchBrowser = oldBrowser })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "login", cfgPath})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "authorization_url") {
		t.Fatalf("err = %v, want authorization_url validation error", err)
	}
	if called {
		t.Fatal("browser launched despite invalid config")
	}
}

// writeCachedAuthCodeToken seeds a token store entry with an auth-code token
// (grant type + refresh token) for the auth-code status assertions.
func writeCachedAuthCodeToken(t *testing.T, root, token string, expiry time.Time) {
	t.Helper()
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := json.Marshal(map[string]any{
		"access_token":  token,
		"token_type":    "Bearer",
		"endpoint":      "https://auth.example.com/token",
		"expiry":        expiry,
		"refresh_token": "rt-secret",
		"grant_type":    "authorization_code",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set("some-workspace:some-config", string(blob)); err != nil {
		t.Fatal(err)
	}
}

// fakeDeviceProvider returns fake device-flow endpoints: the device
// authorization endpoint answers with a verification URI + code, and the
// token endpoint answers authorization_pending once, then grants.
func fakeDeviceProvider(t *testing.T) (*httptest.Server, *httptest.Server) {
	t.Helper()
	deviceSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"device_code":"dev-code","user_code":"AB-1234","verification_uri":"https://idp.example.com/device","verification_uri_complete":"https://idp.example.com/device?user_code=AB-1234","interval":1}`))
	}))
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pending" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"authorization_pending"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok-device","token_type":"Bearer","expires_in":3600,"refresh_token":"dev-rt"}`))
	}))
	t.Cleanup(deviceSrv.Close)
	t.Cleanup(tokenSrv.Close)
	return deviceSrv, tokenSrv
}

func TestAuthLoginDeviceFlowAuto(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	deviceSrv, tokenSrv := fakeDeviceProvider(t)
	cfgPath := writeAuthConfig(t, root, map[string]string{
		"device_authorization_url": deviceSrv.URL,
		"token_url":                tokenSrv.URL,
		"client_id":                "dev-client",
		"client_secret":            "dev-secret",
	})

	// The device flow must not open a browser.
	called := false
	oldBrowser := launchBrowser
	launchBrowser = func(string) error { called = true; return nil }
	t.Cleanup(func() { launchBrowser = oldBrowser })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "login", cfgPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("device login opened a browser")
	}

	output := out.String()
	for _, want := range []string{"open https://idp.example.com/device?user_code=AB-1234", "AB-1234", "login complete", "tok", "yes"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	if strings.Contains(output, "dev-secret") || strings.Contains(output, "tok-device") {
		t.Fatalf("login output leaked a secret:\n%s", output)
	}

	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := store.Keys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 {
		t.Fatalf("cached %d tokens, want 1", len(keys))
	}
	raw, err := store.Get(keys[0])
	if err != nil {
		t.Fatal(err)
	}
	tok, err := auth.ParseCachedToken(raw)
	if err != nil {
		t.Fatalf("ParseCachedToken: %v", err)
	}
	if tok.GrantType != "device_code" {
		t.Fatalf("GrantType = %q, want device_code", tok.GrantType)
	}
	if tok.AccessToken != "tok-device" || tok.RefreshToken != "dev-rt" {
		t.Fatalf("token = %+v", tok)
	}
}

func TestAuthLoginDeviceFlowExplicitFlag(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	deviceSrv, tokenSrv := fakeDeviceProvider(t)
	// Both URLs present: auto would pick authorization_code, so the explicit
	// --flow device flag must win.
	cfgPath := writeAuthConfig(t, root, map[string]string{
		"authorization_url":        "https://idp.example.com/authorize",
		"device_authorization_url": deviceSrv.URL,
		"token_url":                tokenSrv.URL,
		"client_id":                "dev-client",
		"client_secret":            "dev-secret",
	})

	oldFlow := authLoginFlow
	t.Cleanup(func() { authLoginFlow = oldFlow })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "login", "--flow", "device_code", cfgPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "login complete") {
		t.Fatalf("expected login complete, got:\n%s", out.String())
	}
}

func TestAuthLoginUnknownFlow(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)

	deviceSrv, tokenSrv := fakeDeviceProvider(t)
	cfgPath := writeAuthConfig(t, root, map[string]string{
		"device_authorization_url": deviceSrv.URL,
		"token_url":                tokenSrv.URL,
		"client_id":                "dev-client",
		"client_secret":            "dev-secret",
	})

	oldFlow := authLoginFlow
	t.Cleanup(func() { authLoginFlow = oldFlow })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "login", "--flow", "bogus", cfgPath})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown login flow") {
		t.Fatalf("err = %v, want unknown login flow error", err)
	}
}

func TestAuthStatusShowsBackend(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(1*time.Hour))
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "token store: file") {
		t.Fatalf("expected 'token store: file', got:\n%s", out.String())
	}
}

func TestAuthStatusKeychainFallback(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(1*time.Hour))
	chdirWorkspace(t, root)
	t.Setenv("REQLY_TOKEN_STORE", "keychain")

	oldStore := newKeychainStore
	newKeychainStore = func(_, _ string) (*secrets.KeychainStore, error) {
		return nil, errors.New("keychain unavailable (test)")
	}
	t.Cleanup(func() { newKeychainStore = oldStore })

	var warnBuf bytes.Buffer
	oldWarn := warnf
	warnf = func(format string, args ...any) { fmt.Fprintf(&warnBuf, format, args...) }
	t.Cleanup(func() { warnf = oldWarn })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(warnBuf.String(), "falling back to the file store") {
		t.Fatalf("expected fallback warning, got:\n%s", warnBuf.String())
	}
	if !strings.Contains(out.String(), "token store: file") {
		t.Fatalf("expected 'token store: file' after fallback, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "cached") {
		t.Fatalf("expected cached token after fallback, got:\n%s", out.String())
	}
}

func TestAuthStatusUnknownStore(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)
	t.Setenv("REQLY_TOKEN_STORE", "bogus")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown token store") {
		t.Fatalf("err = %v, want unknown token store error", err)
	}
}

func TestAuthStatusFlagOverridesEnv(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedToken(t, root, "very-secret-access-token", time.Now().Add(1*time.Hour))
	chdirWorkspace(t, root)
	t.Setenv("REQLY_TOKEN_STORE", "bogus")

	oldStore := authStoreFlag
	t.Cleanup(func() { authStoreFlag = oldStore })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status", "--store", "file"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("--store file should override a bogus env: %v", err)
	}
	if !strings.Contains(out.String(), "token store: file") {
		t.Fatalf("expected 'token store: file', got:\n%s", out.String())
	}
}

func TestAuthLoginDeviceFlowKeychainFallback(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	chdirWorkspace(t, root)
	t.Setenv("REQLY_TOKEN_STORE", "keychain")

	oldStore := newKeychainStore
	newKeychainStore = func(_, _ string) (*secrets.KeychainStore, error) {
		return nil, errors.New("keychain unavailable (test)")
	}
	t.Cleanup(func() { newKeychainStore = oldStore })

	deviceSrv, tokenSrv := fakeDeviceProvider(t)
	cfgPath := writeAuthConfig(t, root, map[string]string{
		"device_authorization_url": deviceSrv.URL,
		"token_url":                tokenSrv.URL,
		"client_id":                "dev-client",
		"client_secret":            "dev-secret",
	})

	var warnBuf bytes.Buffer
	oldWarn := warnf
	warnf = func(format string, args ...any) { fmt.Fprintf(&warnBuf, format, args...) }
	t.Cleanup(func() { warnf = oldWarn })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "login", cfgPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(warnBuf.String(), "falling back to the file store") {
		t.Fatalf("expected fallback warning, got:\n%s", warnBuf.String())
	}
	if !strings.Contains(out.String(), "login complete") {
		t.Fatalf("expected login complete, got:\n%s", out.String())
	}

	// The token must land in the fallback file store.
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := store.Keys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 {
		t.Fatalf("cached %d tokens in fallback store, want 1", len(keys))
	}
}

func TestAuthStatusShowsGrantAndRefresh(t *testing.T) {
	root := makeTestWorkspace(t, "http://example.com")
	writeCachedAuthCodeToken(t, root, "very-secret-access-token", time.Now().Add(1*time.Hour))
	chdirWorkspace(t, root)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{
		"authorization_code", // grant type
		"yes",                // refresh token cached
		"cached",             // state
		"very",               // masked prefix
		"oken",               // masked suffix
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	if strings.Contains(output, "rt-secret") || strings.Contains(output, "very-secret-access-token") {
		t.Fatalf("status leaked a secret:\n%s", output)
	}
}
