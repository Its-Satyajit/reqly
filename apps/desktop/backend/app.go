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
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/environments"
	"github.com/Its-Satyajit/reqly/internal/history"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppService is a Wails v3 service exposing the Go core to the frontend.
// It should stay thin: business logic belongs in the Go core (internal/core).
type AppService struct {
	requests     *core.RequestService
	auth         *core.AuthService
	environments *core.EnvironmentService
	workspace    *core.WorkspaceService
	runs         *core.CollectionRunService
	history      *core.HistoryService
	// authBackend is the active token-store backend name ("file"/"keychain").
	authBackend string

	runMu      sync.Mutex
	runCancels map[string]context.CancelFunc
}

// NewAppService creates a new AppService. It resolves the workspace rooted at
// the app's working directory, opens the token store (keychain by default,
// falling back to the file store with a warning), and wires store-backed OAuth
// token caching so desktop requests authenticate exactly like the CLI. The
// reqly:// custom-scheme receiver is registered so auth-code logins can
// complete via deep links (feed them with DeliverCustomSchemeCallback).
func NewAppService() *AppService {
	root := collections.FindWorkspaceRoot(".")
	store, backend := openAppTokenStore(root)

	svc := &AppService{
		requests:     core.NewCachedRequestService(store, root),
		authBackend:  backend,
		environments: core.NewEnvironmentService(root),
		workspace:    core.NewWorkspaceService(root),
		runs:         core.NewCollectionRunService(root),
		runCancels:   make(map[string]context.CancelFunc),
	}
	if store != nil {
		svc.auth = core.NewAuthService(store, root)
	}
	// history is per-workspace SQLite (best-effort; no workspace → no history)
	if root != "" {
		if hStore, err := history.NewStore(filepath.Join(root, ".reqly", "history.db")); err == nil {
			svc.history = core.NewHistoryService(hStore, request.NewClient())
		}
	}
	auth.RegisterCustomSchemeReceiver("reqly")
	return svc
}

// SendOptions carries the per-send environment pill plus, for file-backed
// tabs, the request file's workspace-relative path. Env is the environment
// pill (a request file's environment: field); when empty, the app's active
// environment applies. When RequestPath is non-empty, the sent request is
// treated as the tab's live draft and the full inheritance chain (base URL,
// merged headers, inherited auth, variable scopes) is re-resolved against it
// at send time; when empty, the request is sent as-is (scratchpad).
type SendOptions struct {
	Env string `json:"env"`
	// RequestPath is the workspace-relative Request Path of the request file
	// this send belongs to (e.g. "users/auth/login").
	RequestPath string `json:"requestPath"`
}

// SendRequest executes an HTTP request through the core and returns the
// bridge-friendly response for the frontend to render. The active environment
// is resolved from the workspace rooted at the app's working directory; the
// SendOptions environment pill (if any) overrides it. For file-backed sends
// (RequestPath set) the request is re-resolved against the live draft so
// inherited fields always reflect the current container chain, with the
// request-file variable scopes layered above the environment; scratchpad
// sends interpolate with the environment scope alone.
func (s *AppService) SendRequest(r request.Request, opts SendOptions) (*core.SendResponse, error) {
	vars, err := resolveAppEnvironment(opts.Env)
	if err != nil {
		return nil, err
	}
	attachHistoryCookies := func(req *request.Request) {
		if s.history == nil {
			return
		}
		cookies, err := s.history.Cookies(context.Background(), opts.Env)
		if err != nil || len(cookies) == 0 {
			return
		}
		isHTTPS := len(req.URL) >= 8 && req.URL[:8] == "https://"
		matched := history.FilterCookies(cookies, req.URL, isHTTPS)
		if len(matched) == 0 {
			return
		}
		var parts []string
		for _, c := range matched {
			parts = append(parts, c.Name+"="+c.Value)
		}
		cookieVal := ""
		for i, p := range parts {
			if i == 0 {
				cookieVal = p
			} else {
				cookieVal += "; " + p
			}
		}
		for i, h := range req.Headers {
			if h.Key == "Cookie" || h.Key == "cookie" {
				req.Headers[i].Value = h.Value + "; " + cookieVal
				return
			}
		}
		req.Headers = append(req.Headers, request.Header{Key: "Cookie", Value: cookieVal})
	}
	var resp *core.SendResponse
	if opts.RequestPath != "" {
		resolved, err := s.workspace.ResolveSend(opts.RequestPath, r)
		if err != nil {
			return nil, err
		}
		layerVars(vars, resolved.Vars)
		attachHistoryCookies(&resolved.Request)
		resp, err = s.requests.Send(resolved.Request, vars)
		if err != nil {
			return nil, err
		}
		s.recordHistory(opts.RequestPath, resolved.Request, resp, opts.Env)
		return resp, nil
	}
	attachHistoryCookies(&r)
	resp, err = s.requests.Send(r, vars)
	if err != nil {
		return nil, err
	}
	s.recordHistory("", r, resp, opts.Env)
	return resp, nil
}

func (s *AppService) recordHistory(requestPath string, req request.Request, resp *core.SendResponse, env string) {
	if s.history == nil || resp == nil {
		return
	}
	// headers as maps
	reqHdrs := map[string][]string{}
	for _, h := range req.Headers {
		reqHdrs[h.Key] = append(reqHdrs[h.Key], h.Value)
	}
	respHdrs := map[string][]string{}
	for k, vals := range resp.Headers {
		respHdrs[k] = vals
	}
	e := &history.Entry{
		RequestPath: requestPath,
		Method:      string(req.Method),
		URL:         req.URL,
		Env:         env,
		Status:      resp.StatusCode,
		DurationMS:  resp.DurationMS,
		Size:        resp.Size,
		ReqHeaders:  reqHdrs,
		ReqBody:     []byte(req.Body),
		RespHeaders: respHdrs,
		RespBody:    []byte(resp.Body),
	}
	_ = s.history.Record(context.Background(), e)
}

// layerVars overlays every scope of overlay onto base, preserving the
// low → high scope precedence (global < collection < folder < request).
func layerVars(base, overlay *variables.Set) {
	for _, scope := range []variables.Scope{
		variables.ScopeGlobal,
		variables.ScopeCollection,
		variables.ScopeFolder,
		variables.ScopeRequest,
	} {
		overlay.Range(scope, func(key, value string) {
			base.Set(scope, key, value)
		})
	}
}

// AuthLogin runs an interactive OAuth 2.0 login (authorization_code or
// device_code) and caches the token. For authorization_code flows the
// system browser is opened at the provider's authorization page; for
// device_code flows the caller surfaces the verification URI from the
// returned token acquisition steps (the core reports it via the token
// endpoint poll, so this bridge method returns once approved).
func (s *AppService) AuthLogin(config map[string]string, flow string) (*auth.Token, error) {
	if s.auth == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to log in")
	}
	tok, err := s.auth.Login(context.Background(), core.LoginRequest{
		Config: config,
		Flow:   flow,
		Open: func(_ context.Context, authorizationURL string) error {
			return launchAppBrowser(authorizationURL)
		},
	})
	if err != nil {
		return nil, err
	}
	return &tok, nil
}

// AuthStatusResponse is the bridge-friendly auth status: the active store
// backend plus the masked cached tokens.
type AuthStatusResponse struct {
	Backend string                 `json:"backend"`
	Tokens  []core.AuthTokenStatus `json:"tokens"`
}

// AuthStatus lists the cached OAuth tokens for the app's workspace with
// masked values, plus the active store backend.
func (s *AppService) AuthStatus() (*AuthStatusResponse, error) {
	if s.auth == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to see auth status")
	}
	tokens, err := s.auth.Status()
	if err != nil {
		return nil, err
	}
	return &AuthStatusResponse{Backend: s.authBackend, Tokens: tokens}, nil
}

// AuthLogout clears every cached OAuth token for the app's workspace and
// returns how many were removed.
func (s *AppService) AuthLogout() (int, error) {
	if s.auth == nil {
		return 0, fmt.Errorf("no workspace found: nothing to clear")
	}
	return s.auth.Logout()
}

// DeliverCustomSchemeCallback feeds a reqly:// (or other registered) deep
// link into the waiting authorization-code flow. The host calls this when the
// OS delivers a registered URL to the app.
func (s *AppService) DeliverCustomSchemeCallback(uri string) error {
	return auth.DeliverCustomSchemeCallback(uri)
}

// EnvList lists the workspace's environments plus the active one. Secret
// values never cross the bridge — only their names are returned.
func (s *AppService) EnvList() (*core.EnvListResponse, error) {
	return s.environments.List()
}

// EnvRead returns a single environment by name, without secret values.
func (s *AppService) EnvRead(name string) (*core.Environment, error) {
	return s.environments.Read(name)
}

// EnvCreate writes a new environment (name + optional description +
// variables) to disk. Errors on an empty/unsafe name or a duplicate.
func (s *AppService) EnvCreate(name, description string, variables map[string]string) error {
	return s.environments.Create(name, description, variables)
}

// EnvUpdate rewrites an existing environment's description and variables,
// preserving its secrets on disk.
func (s *AppService) EnvUpdate(name, description string, variables map[string]string) error {
	return s.environments.Update(name, description, variables)
}

// EnvUpdateSecrets changes an environment's secrets: `values` holds only the
// secrets the user changed; `remove` names secrets to delete. Existing secret
// values never leave the disk.
func (s *AppService) EnvUpdateSecrets(name string, values map[string]string, remove []string) error {
	return s.environments.UpdateSecrets(name, values, remove)
}

// EnvDelete removes an environment file, clearing the descriptor's active
// selection when the deleted environment was active.
func (s *AppService) EnvDelete(name string) error {
	return s.environments.Delete(name)
}

// EnvSetActive persists name as the workspace's active environment in the
// descriptor. An empty name clears the selection.
func (s *AppService) EnvSetActive(name string) error {
	return s.environments.SetActive(name)
}

// WorkspaceLoad returns the workspace's collection tree (collections →
// folders → requests, all name-sorted) with workspace-relative Request Paths.
// A workspace without a collections/ directory yields an empty tree.
func (s *AppService) WorkspaceLoad() (*core.WorkspaceTree, error) {
	return s.workspace.Load()
}

// WorkspaceOpenRequest resolves a request file by its workspace-relative
// Request Path into its fully resolved form (effective URL, merged headers,
// inherited auth, variable chain, file environment), ready for the editor.
func (s *AppService) WorkspaceOpenRequest(path string) (*core.OpenedRequest, error) {
	return s.workspace.OpenRequest(path)
}

// WorkspaceRunCollection starts a collection run for the collection or folder
// at the workspace-relative Request Path. env names the environment pill for
// the run (empty → the workspace descriptor's active environment). The run
// executes on a background goroutine: per-step results stream to the frontend
// as reqly.run.<id>.step events and the final report as reqly.run.<id>.done.
// Only one run may be active at a time.
func (s *AppService) WorkspaceRunCollection(path, env string, failFast bool) (string, error) {
	if s.runs == nil || s.runs.Root() == "" {
		return "", fmt.Errorf("no workspace found: open a reqly workspace to run collections")
	}
	id := newRunID()
	ctx, cancel := context.WithCancel(context.Background())
	s.runMu.Lock()
	if len(s.runCancels) > 0 {
		s.runMu.Unlock()
		cancel()
		return "", fmt.Errorf("a collection run is already in progress")
	}
	s.runCancels[id] = cancel
	s.runMu.Unlock()

	go func() {
		defer func() {
			cancel()
			s.runMu.Lock()
			delete(s.runCancels, id)
			s.runMu.Unlock()
		}()
		report, err := s.runs.Run(ctx, path, core.RunOptions{
			Env:      env,
			FailFast: failFast,
			OnStep: func(step core.RunStep) {
				emitRunEvent("reqly.run."+id+".step", step)
			},
		})
		if err != nil {
			emitRunEvent("reqly.run."+id+".error", err.Error())
			return
		}
		emitRunEvent("reqly.run."+id+".done", report)
	}()
	return id, nil
}

// WorkspaceRunCancel aborts the active run by id. It returns an error when
// no run with that id is in flight.
func (s *AppService) WorkspaceRunCancel(id string) error {
	s.runMu.Lock()
	cancel, ok := s.runCancels[id]
	s.runMu.Unlock()
	if !ok {
		return fmt.Errorf("no active collection run with id %q", id)
	}
	cancel()
	return nil
}

// runCounter sequences run ids so concurrent timestamps never collide.
var runCounter atomic.Int64

// newRunID returns a unique id for a collection run.
func newRunID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), runCounter.Add(1))
}

// emitRunEvent sends a Wails custom event to the frontend. It is a package
// variable so tests can capture emitted events without a running app. When no
// app is initialized (unit tests, early startup) the event is dropped.
var emitRunEvent = func(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.EmitEvent(&application.CustomEvent{Name: name, Data: data})
	}
}

// WorkspaceSaveRequest persists a request file's editable builder fields back
// to disk, preserving format and every non-editable field. expectedVersion
// must match the file's on-disk fingerprint from OpenedRequest.Version; a
// mismatch surfaces as a changed-on-disk conflict instead of clobbering the
// external edit. The returned string is the new baseline version for the tab.
func (s *AppService) WorkspaceSaveRequest(path string, draft request.Request, expectedVersion string) (string, error) {
	return s.workspace.SaveRequest(path, draft, expectedVersion)
}

// HistoryList returns masked history entries.
func (s *AppService) HistoryList(limit int, offset int, status string, env string) ([]history.Entry, error) {
	if s.history == nil {
		return nil, nil
	}
	var filter *int
	if status != "" {
		filter = statusToFilterHistory(status)
	}
	entries, err := s.history.List(context.Background(), limit, offset, filter)
	if err != nil {
		return nil, err
	}
	if env != "" {
		var out []history.Entry
		for _, e := range entries {
			if e.Env == env {
				out = append(out, e)
			}
		}
		return out, nil
	}
	return entries, nil
}

// HistoryShow returns one masked entry.
func (s *AppService) HistoryShow(id string) (*history.Entry, error) {
	if s.history == nil {
		return nil, fmt.Errorf("no workspace")
	}
	e, err := s.history.Show(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// HistorySearch searches via FTS.
func (s *AppService) HistorySearch(query string, limit int) ([]history.Entry, error) {
	if s.history == nil {
		return nil, nil
	}
	return s.history.Search(context.Background(), query, limit)
}

// HistoryClear clears history.
func (s *AppService) HistoryClear(env *string) error {
	if s.history == nil {
		return nil
	}
	return s.history.Clear(context.Background(), env)
}

// HistoryReplay replays a stored entry verbatim.
func (s *AppService) HistoryReplay(id string) (*core.SendResponse, error) {
	if s.history == nil {
		return nil, fmt.Errorf("no workspace")
	}
	// load raw entry for faithful replay (store shows masked but bodies exact)
	e, err := s.history.Show(context.Background(), id)
	if err != nil {
		return nil, err
	}
	// need raw headers/bodies; use store directly via HistoryService raw? For now use masked entry (bodies exact)
	req := request.Request{Method: request.Method(e.Method), URL: e.URL, Headers: headersFromMapHistory(e.ReqHeaders), Body: string(e.ReqBody)}
	return s.requests.Send(req, nil)
}

// CookieList lists cookies for env.
func (s *AppService) CookieList(env string) ([]history.Cookie, error) {
	if s.history == nil {
		return nil, nil
	}
	return s.history.Cookies(context.Background(), env)
}

// CookieDelete deletes one cookie.
func (s *AppService) CookieDelete(name, domain, path, env string) error {
	if s.history == nil {
		return nil
	}
	return s.history.DeleteCookie(context.Background(), name, domain, path, env)
}

// CookieClear clears cookies.
func (s *AppService) CookieClear(env *string) error {
	if s.history == nil {
		return nil
	}
	return s.history.ClearCookies(context.Background(), env)
}

func statusToFilterHistory(s string) *int {
	switch s {
	case "2xx":
		v := 200
		return &v
	case "4xx":
		v := 400
		return &v
	case "5xx":
		v := 500
		return &v
	default:
		var code int
		if _, err := fmt.Sscan(s, &code); err == nil {
			return &code
		}
		return nil
	}
}

func headersFromMapHistory(m map[string][]string) []request.Header {
	var out []request.Header
	for k, vals := range m {
		for _, v := range vals {
			out = append(out, request.Header{Key: k, Value: v})
		}
	}
	return out
}

// resolveAppEnvironment loads the process-env scope plus the environment
// selected for execution. env is a per-send override (a request file's
// environment pill); when empty, REQLY_ENV or the workspace descriptor's
// environment: field applies. When no workspace or environment is present,
// the returned set still carries the process-env scope so OS variables always
// interpolate.
func resolveAppEnvironment(env string) (*variables.Set, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve working directory: %w", err)
	}
	envFlag := os.Getenv("REQLY_ENV")
	if env != "" {
		envFlag = env
	}
	sel := environments.Selection{
		EnvFlag:   envFlag,
		ConfigEnv: collections.WorkspaceEnvironment(dir),
	}
	set, _, err := environments.ResolveSet(dir, sel)
	if err != nil {
		return nil, err
	}
	return set, nil
}

// openAppTokenStore opens the token store for a workspace root. The desktop
// defaults to the OS keychain (REQLY_TOKEN_STORE overrides), falling back to
// the file store with a warning when no keychain is available. Without a
// workspace, a nil store is returned.
func openAppTokenStore(root string) (secrets.Store, string) {
	if root == "" {
		return nil, ""
	}
	backend := os.Getenv("REQLY_TOKEN_STORE")
	if backend == "" {
		backend = "keychain"
	}
	switch backend {
	case "keychain":
		store, err := secrets.NewKeychainStore("reqly", filepath.Join(root, ".reqly", "keychain.index"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v; falling back to the file store\n", err)
			backend = "file"
			break
		}
		return store, "keychain"
	case "file":
	default:
		fmt.Fprintf(os.Stderr, "warning: unknown token store %q; using the file store\n", backend)
		backend = "file"
	}
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v; token caching disabled\n", err)
		return nil, ""
	}
	return store, backend
}

// launchAppBrowser opens url in the system default browser. It is a package
// variable so tests can substitute a fake driver.
var launchAppBrowser = func(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
