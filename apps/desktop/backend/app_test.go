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
	"sync"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestWorkspaceLoadBridgeReturnsTree(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	if err := os.MkdirAll(filepath.Join(dir, "collections", "users", "auth"), 0o755); err != nil {
		t.Fatal(err)
	}
	for path, contents := range map[string]string{
		"collections/users/reqly.yaml":      "name: users\n",
		"collections/users/list-users.yaml": "request: {method: GET, url: users}\n",
		"collections/users/get-user.yaml":   "request: {method: GET, url: users/1}\n",
		"collections/users/auth/reqly.yaml": "name: auth\n",
		"collections/users/auth/login.yaml": "request: {method: POST, url: auth/login}\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)

	svc := NewAppService()
	tree, err := svc.WorkspaceLoad()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Collections) != 1 || tree.Collections[0].Name != "users" {
		t.Fatalf("collections = %+v", tree.Collections)
	}
	coll := tree.Collections[0]
	if coll.Path != "users" {
		t.Fatalf("collection path = %q, want users", coll.Path)
	}
	if len(coll.Requests) != 2 || coll.Requests[0].Path != "users/get-user" {
		t.Fatalf("requests = %+v", coll.Requests)
	}
	if len(coll.Folders) != 1 || coll.Folders[0].Path != "users/auth" {
		t.Fatalf("folders = %+v", coll.Folders)
	}
	if len(coll.Folders[0].Requests) != 1 || coll.Folders[0].Requests[0].Path != "users/auth/login" {
		t.Fatalf("folder requests = %+v", coll.Folders[0].Requests)
	}
}

func TestWorkspaceLoadBridgeWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml
	t.Chdir(dir)

	svc := NewAppService()
	if _, err := svc.WorkspaceLoad(); err == nil || !strings.Contains(err.Error(), "no workspace found") {
		t.Fatalf("WorkspaceLoad err = %v, want no-workspace error", err)
	}
}

func TestWorkspaceOpenRequestBridgeResolvesRequest(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	if err := os.MkdirAll(filepath.Join(dir, "collections", "users"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "collections/users/reqly.yaml"), []byte("name: users\nbaseURL: https://api.example.com\nheaders:\n  - key: X-Collection\n    value: users\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "collections/users/list-users.yaml"), []byte("name: List Users\nenvironment: staging\nrequest: {method: GET, url: users}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	svc := NewAppService()
	opened, err := svc.WorkspaceOpenRequest("users/list-users")
	if err != nil {
		t.Fatal(err)
	}
	if opened.Path != "users/list-users" || opened.Name != "list-users" {
		t.Fatalf("path/name = %q/%q", opened.Path, opened.Name)
	}
	if opened.FileEnv != "staging" {
		t.Fatalf("fileEnv = %q, want staging", opened.FileEnv)
	}
	if opened.Request.URL != "https://api.example.com/users" {
		t.Fatalf("URL = %q, want https://api.example.com/users", opened.Request.URL)
	}
}

func TestWorkspaceOpenRequestBridgeMissingErrors(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if _, err := svc.WorkspaceOpenRequest("nope/thing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("WorkspaceOpenRequest err = %v, want not-found error", err)
	}
}

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

	set, err := resolveAppEnvironment("")
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

	set, err := resolveAppEnvironment("")
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

	set, err := resolveAppEnvironment("")
	if err != nil {
		t.Fatalf("expected no error without workspace, got %v", err)
	}
	if set == nil {
		t.Fatal("expected a variable set (with process-env scope)")
	}
}

func TestSendRequestEnvironmentPillOverrides(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	// Second environment not selected by the descriptor; the pill must win.
	if err := os.WriteFile(filepath.Join(dir, "environments", "prod.yaml"), []byte("variables:\n  REGION: eu-central-1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var gotRegion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRegion = r.Header.Get("X-Region")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "")

	svc := NewAppService()
	_, err := svc.SendRequest(request.Request{
		Method:  "GET",
		URL:     srv.URL,
		Headers: []request.Header{{Key: "X-Region", Value: "{{REGION}}"}},
	}, SendOptions{Env: "prod"})
	if err != nil {
		t.Fatal(err)
	}
	if gotRegion != "eu-central-1" {
		t.Fatalf("X-Region = %q, want eu-central-1 (pill override)", gotRegion)
	}
}

func TestSendRequestSnapshotVariablesWinOverEnvironment(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "")

	var got map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = map[string]string{
			"X-Region": r.Header.Get("X-Region"),
			"X-Shared": r.Header.Get("X-Shared"),
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := NewAppService()
	_, err := svc.SendRequest(request.Request{
		Method: "GET",
		URL:    srv.URL,
		Headers: []request.Header{
			{Key: "X-Region", Value: "{{REGION}}"},
			{Key: "X-Shared", Value: "{{SHARED}}"},
		},
	}, SendOptions{
		// Environment scope carries REGION=us-west-2; the snapshot overlays a
		// collection-scope REGION and a request-scope SHARED. Precedence must
		// keep request > collection > environment.
		Vars: []core.ResolvedVariable{
			{Name: "REGION", Value: "snapshot-collection", Scope: string(variables.ScopeCollection)},
			{Name: "SHARED", Value: "snapshot-request", Scope: string(variables.ScopeRequest)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["X-Region"] != "snapshot-collection" {
		t.Fatalf("X-Region = %q, want snapshot-collection (collection wins over env)", got["X-Region"])
	}
	if got["X-Shared"] != "snapshot-request" {
		t.Fatalf("X-Shared = %q, want snapshot-request (request wins)", got["X-Shared"])
	}
}

func TestSendRequestMissingPillEnvironmentErrors(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)
	t.Setenv("REQLY_ENV", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := NewAppService()
	_, err := svc.SendRequest(request.Request{Method: "GET", URL: srv.URL}, SendOptions{Env: "nope"})
	if err == nil {
		t.Fatal("expected an error for a missing pill environment, got nil")
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

func writeRunWorkspace(t *testing.T, root, baseURL string) {
	t.Helper()
	writeWorkspace(t, root)
	if err := os.MkdirAll(filepath.Join(root, "collections", "users"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "collections", "users", "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"collections/users/reqly.yaml":       "name: users\nbaseURL: " + baseURL + "\n",
		"collections/users/a.yaml":           "name: A\nrequest: {method: GET, url: /a}\n",
		"collections/users/b.yaml":           "name: B\nrequest: {method: GET, url: /b}\n",
		"collections/users/tests/c.yaml":     "name: C\nrequest: {method: GET, url: /c}\n",
		"collections/users/tests/reqly.yaml": "name: tests\n",
	}
	for path, contents := range files {
		if err := os.WriteFile(filepath.Join(root, path), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// waitForRunFinish polls until the run goroutine has fully finished: the
// service is idle and the run id has been dropped from the cancel registry
// (which happens only after its final event is emitted). Guarding on the
// registry prevents the run goroutine's last emit from racing test teardown,
// which would otherwise restore the real emitter and panic without an app.
func waitForRunFinish(t *testing.T, svc *AppService, id string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for svc.runs.Active() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for collection run to finish")
		}
		time.Sleep(10 * time.Millisecond)
	}
	svc.runMu.Lock()
	_, stillTracked := svc.runCancels[id]
	svc.runMu.Unlock()
	for stillTracked {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for run goroutine cleanup")
		}
		time.Sleep(10 * time.Millisecond)
		svc.runMu.Lock()
		_, stillTracked = svc.runCancels[id]
		svc.runMu.Unlock()
	}
}

func TestWorkspaceRunCollectionStreamsEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	dir := t.TempDir()
	writeRunWorkspace(t, dir, srv.URL)
	t.Chdir(dir)

	svc := NewAppService()
	var mu sync.Mutex
	var names []string
	var steps []core.RunStep
	var done *core.RunReport
	orig := emitRunEvent
	emitRunEvent = func(name string, data any) {
		mu.Lock()
		defer mu.Unlock()
		names = append(names, name)
		switch v := data.(type) {
		case core.RunStep:
			steps = append(steps, v)
		case *core.RunReport:
			done = v
		}
	}
	defer func() { emitRunEvent = orig }()

	id, err := svc.WorkspaceRunCollection("users", "dev", false)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("expected a run id, got empty")
	}
	waitForRunFinish(t, svc, id)

	mu.Lock()
	defer mu.Unlock()
	if len(names) != 4 {
		t.Fatalf("events = %v, want 3 step + 1 done", names)
	}
	if len(steps) != 3 {
		t.Fatalf("steps = %+v, want 3", steps)
	}
	for _, ev := range names {
		if !strings.HasPrefix(ev, "reqly.run."+id+".") {
			t.Fatalf("event %q does not belong to run %q", ev, id)
		}
	}
	if steps[0].RequestPath != "users/a" || steps[1].RequestPath != "users/b" || steps[2].RequestPath != "users/tests/c" {
		t.Fatalf("step paths = %q/%q/%q, want users/a users/b users/tests/c", steps[0].RequestPath, steps[1].RequestPath, steps[2].RequestPath)
	}
	if !steps[0].Passed || !steps[1].Passed || !steps[2].Passed {
		t.Fatalf("steps not passed: %+v", steps)
	}
	if done == nil {
		t.Fatal("no done event")
	}
	if done.Total != 3 || done.Passed != 3 || done.Failed != 0 || !done.OK {
		t.Fatalf("report = %+v", done)
	}
}

func TestWorkspaceRunCollectionFolderScope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	dir := t.TempDir()
	writeRunWorkspace(t, dir, srv.URL)
	t.Chdir(dir)

	svc := NewAppService()
	var mu sync.Mutex
	var steps []core.RunStep
	orig := emitRunEvent
	emitRunEvent = func(name string, data any) {
		if !strings.HasSuffix(name, ".step") {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		if v, ok := data.(core.RunStep); ok {
			steps = append(steps, v)
		}
	}
	defer func() { emitRunEvent = orig }()

	id, err := svc.WorkspaceRunCollection("users/tests", "dev", false)
	if err != nil {
		t.Fatal(err)
	}
	waitForRunFinish(t, svc, id)

	mu.Lock()
	defer mu.Unlock()
	if len(steps) != 1 || steps[0].RequestPath != "users/tests/c" {
		t.Fatalf("steps = %+v, want only users/tests/c", steps)
	}
}

func TestWorkspaceRunCollectionWithoutWorkspaceErrors(t *testing.T) {
	dir := t.TempDir() // no reqly.yaml
	t.Chdir(dir)

	svc := NewAppService()
	if _, err := svc.WorkspaceRunCollection("users", "dev", false); err == nil || !strings.Contains(err.Error(), "no workspace found") {
		t.Fatalf("err = %v, want no-workspace error", err)
	}
}

func TestWorkspaceRunCollectionSingleFlight(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	defer srv.Close()

	dir := t.TempDir()
	writeRunWorkspace(t, dir, srv.URL)
	t.Chdir(dir)

	svc := NewAppService()
	orig := emitRunEvent
	emitRunEvent = func(string, any) {}
	defer func() { emitRunEvent = orig }()

	id, err := svc.WorkspaceRunCollection("users", "dev", false)
	if err != nil {
		t.Fatal(err)
	}
	defer svc.WorkspaceRunCancel(id)

	if _, err := svc.WorkspaceRunCollection("users", "dev", false); err == nil || !strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("second run err = %v, want single-flight error", err)
	}
	close(release)
	waitForRunFinish(t, svc, id)
}

func TestWorkspaceRunCancelAbortsRun(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	defer func() {
		close(release)
	}()
	defer srv.Close()

	dir := t.TempDir()
	writeRunWorkspace(t, dir, srv.URL)
	t.Chdir(dir)

	svc := NewAppService()
	var mu sync.Mutex
	var sawDone bool
	var sawErrorEvent string
	orig := emitRunEvent
	emitRunEvent = func(name string, data any) {
		mu.Lock()
		defer mu.Unlock()
		if strings.HasSuffix(name, ".done") {
			sawDone = true
		}
		if strings.HasSuffix(name, ".error") {
			if s, ok := data.(string); ok {
				sawErrorEvent = s
			}
		}
	}
	defer func() { emitRunEvent = orig }()

	id, err := svc.WorkspaceRunCollection("users", "dev", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.WorkspaceRunCancel(id); err != nil {
		t.Fatalf("cancel err = %v", err)
	}
	// The blocking server would hang the run until release is closed in
	// teardown, so a prompt finish proves the in-flight request was aborted.
	waitForRunFinish(t, svc, id)

	mu.Lock()
	defer mu.Unlock()
	if !sawDone {
		t.Fatalf("expected the run to finish after cancel, error event = %q", sawErrorEvent)
	}
}

func TestWorkspaceRunCancelUnknownIDErrors(t *testing.T) {
	dir := t.TempDir()
	writeWorkspace(t, dir)
	t.Chdir(dir)

	svc := NewAppService()
	if err := svc.WorkspaceRunCancel("nope"); err == nil || !strings.Contains(err.Error(), "no active collection run") {
		t.Fatalf("err = %v, want unknown-id error", err)
	}
}
