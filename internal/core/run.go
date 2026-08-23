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
	opened := secrets.OpenForWorkspace(root, "file")
	s := &RequestService{root: root, warning: opened.Warning}
	if opened.Store != nil {
		s.client = request.NewClient(request.WithTokenCache(opened.Store, root))
	} else {
		s.client = request.NewClient()
	}
	return s
}

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
	if opts.RecordHistory == nil || *opts.RecordHistory {
		if warn := s.record(opts, r, resp); warn != "" {
			warning = strings.TrimSpace(warning + " " + warn)
		}
	}

	maskAcquired(masker, resp.AuthToken)
	resp.Headers = maskHeaders(resp.Headers, masker)
	resp.Body = []byte(masker.Mask(string(resp.Body)))
	resp.AuthToken = ""

	return &RunResult{Response: resp, Warning: strings.TrimSpace(warning)}, nil
}

func layerScope(dst, src *variables.Set) {
	if src == nil {
		return
	}
	src.Range(variables.ScopeRuntime, func(key, value string) {
		dst.Set(variables.ScopeRuntime, key, value)
	})
}

func (s *RequestService) record(opts RunRequestOptions, r request.Request, resp *response.Response) string {
	h := s.recorder()
	if h == nil {
		return ""
	}
	reqHdrs := map[string][]string{}
	for _, hd := range r.Headers {
		reqHdrs[hd.Key] = append(reqHdrs[hd.Key], hd.Value)
	}
	env := os.Getenv("REQLY_ENV")
	if env == "" {
		env = opts.EnvFlag
	}
	if env == "" {
		env = opts.FileEnv
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
	if err := h.Record(context.Background(), e); err != nil {
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
	return &maskedError{msg: masker.Mask(err.Error())}
}

type maskedError struct{ msg string }

func (e *maskedError) Error() string { return e.msg }

func errNoWorkspace() error {
	return &noWorkspaceError{}
}

type noWorkspaceError struct{}

func (e *noWorkspaceError) Error() string {
	return "no workspace found: open a reqly workspace to send requests"
}
