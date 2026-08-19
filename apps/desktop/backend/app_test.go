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

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeWorkspace(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "environments"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "environments", "dev.yaml"), []byte(`
variables:
  REGION: us-west-2
secrets:
  API_KEY: dev-secret
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reqly.yaml"), []byte("environment: dev\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveAppEnvironmentFromWorkspaceDescriptor(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "")

	set, err := resolveAppEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := set.Resolve("REGION"); !ok || got != "us-west-2" {
		t.Fatalf("expected environment variable REGION=us-west-2, got %q (ok=%v)", got, ok)
	}
	if got, ok := set.Resolve("API_KEY"); !ok || got != "dev-secret" {
		t.Fatalf("expected environment secret API_KEY, got %q (ok=%v)", got, ok)
	}
}

func TestResolveAppEnvironmentFromREQLYEnv(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	// Second environment not selected by the descriptor.
	if err := os.WriteFile(filepath.Join(dir, "environments", "prod.yaml"), []byte("variables:\n  REGION: eu-central-1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "prod")

	set, err := resolveAppEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := set.Resolve("REGION"); !ok || got != "eu-central-1" {
		t.Fatalf("expected REQLY_ENV to win, got %q (ok=%v)", got, ok)
	}
}

func TestResolveAppEnvironmentWithoutWorkspaceIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "")

	set, err := resolveAppEnvironment()
	if err != nil {
		t.Fatalf("expected no error without workspace, got %v", err)
	}
	if set == nil {
		t.Fatal("expected a variable set (with process-env scope)")
	}
}

// fakeAppDeviceEndpoints returns fake device-flow endpoints for the bridge
// tests: the device endpoint answers with a verification URI + code, and the
// token endpoint answers authorization_pending once, then grants.
func fakeAppDeviceEndpoints(t *testing.T) (*httptest.Server, *httptest.Server) {
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
		w.Write([]byte(`{"access_token":"desktop-device-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"desktop-device-rt"}`))
	}))
	t.Cleanup(deviceSrv.Close)
	t.Cleanup(tokenSrv.Close)
	return deviceSrv, tokenSrv
}

func TestAuthLoginDeviceFlowPersistsAndStatus(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)
	t.Setenv("REQLY_TOKEN_STORE", "file") // hermetic: never touch the OS keychain

	deviceSrv, tokenSrv := fakeAppDeviceEndpoints(t)
	svc := NewAppService()

	cfg := map[string]string{
		"grant_type":               "device_code",
		"device_authorization_url": deviceSrv.URL,
		"token_url":                tokenSrv.URL,
		"client_id":                "desktop-client",
		"client_secret":            "desktop-secret",
	}
	tok, err := svc.AuthLogin(cfg, "device_code")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "desktop-device-tok" || tok.RefreshToken != "desktop-device-rt" {
		t.Fatalf("tok = %+v", tok)
	}

	status, err := svc.AuthStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.Backend != "file" {
		t.Fatalf("backend = %q, want file", status.Backend)
	}
	if len(status.Tokens) != 1 {
		t.Fatalf("status tokens = %d, want 1", len(status.Tokens))
	}
	s := status.Tokens[0]
	if s.GrantType != "device_code" || !s.HasRefresh || s.State != "cached" {
		t.Fatalf("token status = %+v", s)
	}
	if strings.Contains(s.AccessToken, "desktop-device-tok") {
		t.Fatalf("status leaked the full token: %q", s.AccessToken)
	}

	cleared, err := svc.AuthLogout()
	if err != nil {
		t.Fatal(err)
	}
	if cleared != 1 {
		t.Fatalf("cleared %d, want 1", cleared)
	}
	status, err = svc.AuthStatus()
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Tokens) != 0 {
		t.Fatalf("status tokens after logout = %d, want 0", len(status.Tokens))
	}
}

func TestAuthLoginAuthCodeOpensBrowser(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)
	t.Setenv("REQLY_TOKEN_STORE", "file")

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		cb := r.URL.Query().Get("redirect_uri")
		q := strings.SplitN(cb, "?", 2)[0] + "?code=desktop-code&state=" + state
		http.Redirect(w, r, q, http.StatusFound)
	}))
	defer authSrv.Close()
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"desktop-authcode-tok","token_type":"Bearer","expires_in":3600,"refresh_token":"desktop-authcode-rt"}`))
	}))
	defer tokenSrv.Close()

	// The Open hook must launch the browser at the authorization URL; the
	// fake drives the callback like a browser would.
	var opened string
	oldLaunch := launchAppBrowser
	launchAppBrowser = func(url string) error {
		opened = url
		client := &http.Client{}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	}
	t.Cleanup(func() { launchAppBrowser = oldLaunch })

	svc := NewAppService()
	tok, err := svc.AuthLogin(map[string]string{
		"authorization_url": authSrv.URL,
		"token_url":         tokenSrv.URL,
		"client_id":         "desktop-client",
		"client_secret":     "desktop-secret",
	}, "authorization_code")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "desktop-authcode-tok" {
		t.Fatalf("AccessToken = %q", tok.AccessToken)
	}
	if !strings.Contains(opened, authSrv.URL) {
		t.Fatalf("browser opened at %q, want the authorization URL", opened)
	}
}

func TestAuthStatusWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml
	t.Chdir(dir)

	svc := NewAppService()
	if _, err := svc.AuthStatus(); err == nil || !strings.Contains(err.Error(), "no workspace found") {
		t.Fatalf("AuthStatus err = %v, want no-workspace error", err)
	}
	if _, err := svc.AuthLogin(map[string]string{}, "device_code"); err == nil {
		t.Fatal("AuthLogin without workspace: err = nil, want error")
	}
	if _, err := svc.AuthLogout(); err == nil {
		t.Fatal("AuthLogout without workspace: err = nil, want error")
	}
}

func TestDeliverCustomSchemeCallbackBridge(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)
	t.Setenv("REQLY_TOKEN_STORE", "file")

	// No flow is waiting: the bridge surfaces the auth package's error.
	svc := NewAppService()
	err := svc.DeliverCustomSchemeCallback("reqly://callback?code=x&state=y")
	if err == nil || !strings.Contains(err.Error(), "no authorization flow waiting") {
		t.Fatalf("err = %v, want no-waiting-flow error", err)
	}
}

func TestEnvListBridgeReturnsEnvironmentsAndActive(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	list, err := svc.EnvList()
	if err != nil {
		t.Fatal(err)
	}
	if list.Active != "dev" {
		t.Fatalf("active = %q, want dev", list.Active)
	}
	if len(list.Environments) != 1 {
		t.Fatalf("environments = %d, want 1", len(list.Environments))
	}
	env := list.Environments[0]
	if env.Name != "dev" || env.Variables["REGION"] != "us-west-2" {
		t.Fatalf("env = %+v", env)
	}
	if len(env.Secrets) != 1 || env.Secrets[0] != "API_KEY" {
		t.Fatalf("secrets = %v, want [API_KEY]", env.Secrets)
	}
}

func TestEnvReadBridgeReturnsEnvironmentWithoutSecretValues(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	env, err := svc.EnvRead("dev")
	if err != nil {
		t.Fatal(err)
	}
	if env.Name != "dev" || env.Variables["REGION"] != "us-west-2" {
		t.Fatalf("env = %+v", env)
	}
	if len(env.Secrets) != 1 || env.Secrets[0] != "API_KEY" {
		t.Fatalf("secrets = %v, want [API_KEY]", env.Secrets)
	}
}

func TestEnvSetActiveBridgePersists(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	// Second environment not selected by the descriptor.
	if err := os.WriteFile(filepath.Join(dir, "environments", "prod.yaml"), []byte("variables:\n  REGION: eu-central-1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.EnvSetActive("prod"); err != nil {
		t.Fatal(err)
	}
	list, err := svc.EnvList()
	if err != nil {
		t.Fatal(err)
	}
	if list.Active != "prod" {
		t.Fatalf("active = %q, want prod", list.Active)
	}
}

func TestEnvSetActiveBridgeClears(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.EnvSetActive(""); err != nil {
		t.Fatal(err)
	}
	list, err := svc.EnvList()
	if err != nil {
		t.Fatal(err)
	}
	if list.Active != "" {
		t.Fatalf("active = %q, want cleared", list.Active)
	}
}

func TestEnvListBridgeWithoutWorkspaceIsEmpty(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml
	t.Chdir(dir)

	svc := NewAppService()
	list, err := svc.EnvList()
	if err != nil {
		t.Fatal(err)
	}
	if list.Active != "" || len(list.Environments) != 0 {
		t.Fatalf("got active=%q envs=%d, want empty", list.Active, len(list.Environments))
	}
}

func TestEnvCreateBridgeWritesEnvironment(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.EnvCreate("staging", "Staging server", map[string]string{"REGION": "ap-south-1"}); err != nil {
		t.Fatal(err)
	}

	env, err := svc.EnvRead("staging")
	if err != nil {
		t.Fatal(err)
	}
	if env.Name != "staging" || env.Description != "Staging server" {
		t.Fatalf("env = %+v", env)
	}
	if env.Variables["REGION"] != "ap-south-1" {
		t.Fatalf("variables = %v", env.Variables)
	}
	if len(env.Secrets) != 0 {
		t.Fatalf("secrets = %v, want none", env.Secrets)
	}
}

func TestEnvCreateBridgeDuplicateErrors(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.EnvCreate("dev", "Duplicate", nil); err == nil {
		t.Fatal("expected duplicate error, got nil")
	}
}

func TestEnvUpdateBridgePersistsChanges(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.EnvUpdate("dev", "Updated", map[string]string{"REGION": "eu-west-1"}); err != nil {
		t.Fatal(err)
	}

	env, err := svc.EnvRead("dev")
	if err != nil {
		t.Fatal(err)
	}
	if env.Description != "Updated" || env.Variables["REGION"] != "eu-west-1" {
		t.Fatalf("env = %+v", env)
	}
}

func TestEnvUpdateBridgeMissingEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.EnvUpdate("nope", "x", nil); err == nil {
		t.Fatal("expected missing-environment error, got nil")
	}
}

func TestEnvUpdateSecretsBridgePersists(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.EnvUpdateSecrets("dev", map[string]string{"API_KEY": "new-key"}, nil); err != nil {
		t.Fatal(err)
	}

	env, err := svc.EnvRead("dev")
	if err != nil {
		t.Fatal(err)
	}
	// API_KEY name still present; value never exposed.
	if len(env.Secrets) != 1 || env.Secrets[0] != "API_KEY" {
		t.Fatalf("secrets = %v, want [API_KEY]", env.Secrets)
	}
}

func TestEnvDeleteBridgeRemovesFileAndClearsActive(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.EnvDelete("dev"); err != nil {
		t.Fatal(err)
	}

	list, err := svc.EnvList()
	if err != nil {
		t.Fatal(err)
	}
	if list.Active != "" || len(list.Environments) != 0 {
		t.Fatalf("got active=%q envs=%d, want empty", list.Active, len(list.Environments))
	}
}
