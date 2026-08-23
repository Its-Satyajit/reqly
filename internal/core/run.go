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
	"os"
	"path/filepath"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/environments"
	"github.com/Its-Satyajit/reqly/internal/history"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// RunOptions carries the per-run inputs a front-end owns: selection hints and
// observers. Everything else — precedence, layering, masking, recording — is
// the pipeline's job (ADR 0025).
type RunRequestOptions struct {
	// EnvFlag is the environment named by --env (REQLY_ENV still wins).
	EnvFlag string
	// FileEnv is the request file's environment: field.
	FileEnv string
	// FileVars are the request/test file's own variables, layered between
	// the environment scope and runtime injection.
	FileVars *variables.Set
	// RuntimeVars are layered last (e.g. Bulk rows, Pagination state).
	RuntimeVars *variables.Set
	// RequestPath identifies the request in history entries ("none for scratchpad").
	RequestPath string
	// RecordHistory opts out of history recording; nil means record when a
	// workspace exists.
	RecordHistory *bool
	// AttachCookies opts out of Cookie Jar attachment (e.g. verbatim replay);
	// nil means attach the workspace jar's matching cookies when a history
	// store is available.
	AttachCookies *bool
	// OnRetry observes automatic retries before each backoff wait.
	OnRetry func(request.RetryEvent)
}

// RunResult is a masked execution result. Raw response bytes never leave the
// pipeline: secrets from environments, .env files, auth configs, and any
// acquired OAuth token are redacted from headers, body, and errors before
// return. History was recorded with raw bytes before masking.
type RunResult struct {
	Response *response.Response
	// Warning carries non-fatal notes (store fallback, recording failure).
	Warning string
}

// NewRunService returns a RequestService bound to a workspace root: token
// caching uses the shared backend policy (secrets.OpenForWorkspace) and
// history recording is enabled against <root>/.reqly/history.db.
func NewRunService(root string) *RequestService {
	return NewRunServiceForWorkspace(root, "file")
}

// NewRunServiceForWorkspace is NewRunService with an explicit default token
// backend ("file" or "keychain") — the desktop prefers the keychain, headless
// surfaces the file store. REQLY_TOKEN_STORE overrides both.
func NewRunServiceForWorkspace(root, defaultTokenBackend string) *RequestService {
	opened := secrets.OpenForWorkspace(root, defaultTokenBackend)
	s := &RequestService{root: root, warning: opened.Warning}
	if opened.Store != nil {
		s.client = request.NewClient(request.WithTokenCache(opened.Store, root))
	} else {
		s.client = request.NewClient()
	}
	return s
}

// History exposes the service's history store (lazily opened), or nil without
// a workspace. Cookie listing/deletion and history queries go through here so
// all front-ends share one store handle.
func (s *RequestService) History() *HistoryService { return s.recorder() }

// Warning returns the service-level non-fatal warning (store fallback), if any.
func (s *RequestService) Warning() string { return s.warning }

// Run executes one logical send through the full pipeline: environment
// resolution → variable layering → engine execution (with retry observation)
// → history recording (raw bytes) → masked output.
func (s *RequestService) Run(ctx context.Context, r request.Request, opts RunRequestOptions) (*RunResult, error) {
	if s == nil || s.client == nil {
		return nil, errNoWorkspace()
	}
	dir := s.root
	if dir == "" {
		dir = "."
	}

	envFlag := os.Getenv("REQLY_ENV")
	if envFlag == "" {
		envFlag = opts.EnvFlag
	}
	set, masker, err := environments.ResolveSet(dir, environments.Selection{
		EnvFlag:   envFlag,
		FileEnv:   opts.FileEnv,
		ConfigEnv: collections.WorkspaceEnvironment(dir),
	})
	if err != nil {
		return nil, err
	}
	layerScope(set, opts.FileVars)
	layerScope(set, opts.RuntimeVars)
	masker.Add(auth.MaskValues(r.Auth.Type, r.Auth.Config, set)...)

	// Cookie Jar: attach matching workspace cookies unless the caller opted
	// out (verbatim replay). The jar half that ingests Set-Cookie lives in
	// record() below — both halves of the jar live in this pipeline.
	if opts.AttachCookies == nil || *opts.AttachCookies {
		if h := s.recorder(); h != nil {
			attachJarCookies(&r, h, envName(envFlag, opts))
		}
	}

	var onRetry func(request.RetryEvent)
	if opts.OnRetry != nil {
		// Wrap so network-error strings (which can echo URLs) are redacted
		// before they reach the caller — secrets never leave the pipeline.
		onRetry = func(e request.RetryEvent) {
			if e.Err != nil {
				e.Err = &maskedError{msg: masker.Mask(e.Err.Error())}
			}
			opts.OnRetry(e)
		}
	}

	resp, err := s.client.ExecuteWithOnRetry(ctx, &r, set, onRetry)
	if err != nil {
		return nil, maskErr(masker, err)
	}

	warning := s.warning
	// The jar works independently of recording: Set-Cookie is ingested on
	// every send through a workspace, recorded or not.
	if h := s.recorder(); h != nil {
		h.IngestSetCookies(context.Background(), resp.Headers, envName(envFlag, opts))
	}
	if opts.RecordHistory == nil || *opts.RecordHistory {
		if warn := s.record(opts, r, *resp, envName(envFlag, opts)); warn != "" {
			warning = strings.TrimSpace(warning + " " + warn)
		}
	}

	maskAcquired(masker, resp.AuthToken)
	resp.Headers = maskHeaders(resp.Headers, masker)
	resp.Body = []byte(masker.Mask(string(resp.Body)))
	resp.AuthToken = ""

	return &RunResult{Response: resp, Warning: strings.TrimSpace(warning)}, nil
}

// layerScope copies every variable of src into dst, preserving scopes so
// dst's precedence resolution orders them correctly.
// envName resolves the environment label used to partition jar entries and
// history rows: the same precedence everywhere.
func envName(envFlag string, opts RunRequestOptions) string {
	if envFlag != "" {
		return envFlag
	}
	return opts.FileEnv
}

// attachJarCookies splices the workspace jar's cookies matching req's URL
// into a Cookie header, preserving any Cookie header the request already
// carries.
func attachJarCookies(r *request.Request, h *HistoryService, env string) {
	cookies, err := h.Cookies(context.Background(), env)
	if err != nil || len(cookies) == 0 {
		return
	}
	isHTTPS := len(r.URL) >= 8 && r.URL[:8] == "https://"
	matched := history.FilterCookies(cookies, r.URL, isHTTPS)
	if len(matched) == 0 {
		return
	}
	parts := make([]string, 0, len(matched))
	for _, c := range matched {
		parts = append(parts, c.Name+"="+c.Value)
	}
	cookieVal := strings.Join(parts, "; ")
	for i, hd := range r.Headers {
		if hd.Key == "Cookie" || hd.Key == "cookie" {
			r.Headers[i].Value = hd.Value + "; " + cookieVal
			return
		}
	}
	r.Headers = append(r.Headers, request.Header{Key: "Cookie", Value: cookieVal})
}

func layerScope(dst, src *variables.Set) {
	if src == nil {
		return
	}
	for _, scope := range variables.Precedence() {
		src.Range(scope, func(key, value string) {
			dst.Set(scope, key, value)
		})
	}
}

func (s *RequestService) record(opts RunRequestOptions, r request.Request, resp response.Response, env string) string {
	if s == nil {
		return ""
	}
	h := s.recorder()
	if h == nil {
		return ""
	}
	reqHdrs := map[string][]string{}
	for _, hd := range r.Headers {
		reqHdrs[hd.Key] = append(reqHdrs[hd.Key], hd.Value)
	}
	e := &history.Entry{
		RequestPath: opts.RequestPath,
		Method:      string(r.Method),
		URL:         r.URL,
		Env:         env,
		Status:      resp.StatusCode,
		DurationMS:  resp.Duration.Milliseconds(),
		Size:        resp.Size,
		ReqHeaders:  reqHdrs,
		ReqBody:     []byte(r.Body),
		RespHeaders: resp.Headers,
		RespBody:    resp.Body,
		Attempts:    resp.Attempts,
	}
	if err := h.insert(context.Background(), e); err != nil {
		return "history recording failed: " + err.Error()
	}
	return ""
}

// recorder lazily opens the workspace history store on first use so services
// that never record pay nothing.
func (s *RequestService) recorder() *HistoryService {
	if s.root == "" {
		return nil
	}
	s.histOnce.Do(func() {
		store, err := history.NewStore(filepath.Join(s.root, ".reqly", "history.db"))
		if err != nil {
			return
		}
		s.historySvc = &HistoryService{store: store, client: s.client}
	})
	return s.historySvc
}

// Close releases resources held by the service (the history store, if opened).
func (s *RequestService) Close() error {
	if s != nil && s.historySvc != nil {
		return s.historySvc.Close()
	}
	return nil
}

func maskAcquired(masker *environments.Masker, token string) {
	if token != "" {
		masker.Add(token)
	}
}

func maskErr(masker *environments.Masker, err error) error {
	return &maskedError{msg: masker.Mask(err.Error()), err: err}
}

type maskedError struct {
	msg string
	err error
}

func (e *maskedError) Error() string { return e.msg }

// Unwrap preserves the original error chain so callers can classify failures
// (e.g. errors.Is(err, context.Canceled)) even after secret masking.
func (e *maskedError) Unwrap() error { return e.err }

func errNoWorkspace() error {
	return &noWorkspaceError{}
}

type noWorkspaceError struct{}

func (e *noWorkspaceError) Error() string {
	return "no workspace found: open a reqly workspace to send requests"
}

// SendResponseFrom maps a masked RunResult onto the bridge-friendly
// SendResponse DTO (Desktop/MCP surfaces).
func SendResponseFrom(rr *RunResult) *SendResponse {
	if rr == nil || rr.Response == nil {
		return nil
	}
	resp := rr.Response
	return &SendResponse{
		StatusCode: resp.StatusCode,
		StatusText: resp.StatusText,
		Proto:      resp.Proto,
		Headers:    resp.Headers,
		Body:       string(resp.Body),
		DurationMS: resp.Duration.Milliseconds(),
		Size:       resp.Size,
		OK:         resp.OK(),
		Attempts:   resp.Attempts,
	}
}

// NewRunServiceWithTokenStore binds a service with an explicit token store
// (may be nil). For front-ends that obtain the store through the shared
// secrets.OpenForWorkspace seam themselves — e.g. the desktop, which also
// hands the same store to its auth service.
func NewRunServiceWithTokenStore(root string, store secrets.Store) *RequestService {
	s := &RequestService{root: root}
	if store != nil {
		s.client = request.NewClient(request.WithTokenCache(store, root))
	} else {
		s.client = request.NewClient()
	}
	return s
}
