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

package request

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// Client executes Request values over HTTP and returns response.Response
// values. It is the shared engine used by the Desktop, CLI, and MCP.
type Client struct {
	http    *http.Client
	timeout time.Duration
	// tokens enables store-backed caching of tokens acquired by TokenSource
	// schemes (e.g. oauth2). When set, tokens are reused across requests
	// until they near expiry instead of being re-acquired each time.
	tokens *TokenCache
}

// TokenCache configures store-backed caching of tokens acquired by TokenSource
// schemes. Root scopes the cache keys so tokens from different workspaces do
// not collide.
type TokenCache struct {
	store secrets.Store
	root  string

	// mu serializes acquisition for a given cache key so concurrent requests
	// for the same config do not double-acquire.
	mu sync.Mutex
}

// Option configures a Client.
type Option func(*Client)

// WithTimeout sets a default timeout for all requests. Requests with an
// explicit Timeout field override it.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithHTTPClient replaces the underlying http.Client (useful for tests and
// for injecting custom transports).
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.http = client
	}
}

// WithTokenCache enables store-backed caching of tokens acquired by
// TokenSource schemes (e.g. oauth2 Client Credentials). Root scopes cache
// keys to the workspace. A nil store disables caching.
func WithTokenCache(store secrets.Store, root string) Option {
	return func(c *Client) {
		if store == nil {
			c.tokens = nil
			return
		}
		c.tokens = &TokenCache{store: store, root: root}
	}
}

// NewClient returns a Client with the given options.
func NewClient(opts ...Option) *Client {
	c := &Client{http: &http.Client{}}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Execute runs a Request and returns the response. Variables are interpolated
// into the URL, headers, query parameters, and body before sending.
func (c *Client) Execute(ctx context.Context, r *Request, vars auth.Interpolator) (*response.Response, error) {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(r.Timeout)*time.Millisecond)
		defer cancel()
	}

	req, authToken, err := c.build(ctx, r, vars)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	// Digest (and other challenge-based schemes) respond to a 401 Digest
	// WWW-Authenticate challenge by computing credentials and retrying once.
	// The retry is bounded to a single challenge/response round-trip and only
	// fires for a matching Digest challenge; other 401s return as-is.
	if resp.StatusCode == http.StatusUnauthorized && r.Auth.Type != "" {
		challenge := resp.Header.Get("WWW-Authenticate")
		if strings.HasPrefix(challenge, "Digest") {
			retryReq, _, err := c.build(ctx, r, vars)
			if err != nil {
				return nil, err
			}
			retried, err := auth.Challenge(retryReq, r.Auth.Type, challenge, r.Auth.Config, vars)
			if err != nil {
				return nil, err
			}
			if retried {
				// Drain and close the 401 body so the connection can be
				// reused for the retry. A drain/close error is irrelevant
				// here: this response is being abandoned, so the errors are
				// explicitly not propagated.
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				resp, err = c.http.Do(retryReq)
				if err != nil {
					return nil, err
				}
			}
		} else if s, ok := auth.Lookup(r.Auth.Type); ok {
			// Reactive refresh: a 401 on a TokenSource scheme with caching
			// enabled forces the cached token out and renews it — via the
			// refresh-token grant when a refresh token is cached (no second
			// browser flow), otherwise by re-acquiring — then retries exactly
			// once. A second 401 is returned as-is (no retry loop).
			if ts, isTokenSource := s.(auth.TokenSource); isTokenSource && c.tokens != nil {
				if err := c.forceRefreshToken(ctx, ts, r.Auth.Config, vars); err != nil {
					// Drain and close the 401 body before bailing so the
					// connection can be reused; the response is abandoned.
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					return nil, err
				}
				retryReq, refreshedToken, err := c.build(ctx, r, vars)
				if err != nil {
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					return nil, err
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				resp, err = c.http.Do(retryReq)
				if err != nil {
					return nil, err
				}
				authToken = refreshedToken
			}
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &response.Response{
		StatusCode: resp.StatusCode,
		StatusText: http.StatusText(resp.StatusCode),
		Proto:      resp.Proto,
		Headers:    resp.Header,
		Body:       body,
		Duration:   time.Since(start),
		Size:       int64(len(body)),
		AuthToken:  authToken,
	}, nil
}

// build constructs a net/http Request from the model, applying interpolation,
// query parameters, headers, body, and authentication. It returns the built
// request plus the resolved auth token (empty when no token was acquired).
func (c *Client) build(ctx context.Context, r *Request, vars auth.Interpolator) (*http.Request, string, error) {
	if vars == nil {
		vars = variables.NewSet()
	}

	method := r.Method
	if method == "" {
		method = MethodGet
	}

	rawURL, err := vars.Interpolate(r.URL)
	if err != nil {
		return nil, "", err
	}
	if rawURL == "" {
		return nil, "", errors.New("request has no URL")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	q := u.Query()
	for _, p := range r.Query {
		key, err := vars.Interpolate(p.Key)
		if err != nil {
			return nil, "", err
		}
		value, err := vars.Interpolate(p.Value)
		if err != nil {
			return nil, "", err
		}
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	var body io.Reader
	if r.Body != "" {
		interpolated, err := vars.Interpolate(r.Body)
		if err != nil {
			return nil, "", err
		}
		body = strings.NewReader(interpolated)
	}

	req, err := http.NewRequestWithContext(ctx, string(method), u.String(), body)
	if err != nil {
		return nil, "", err
	}

	if body != nil {
		req.Header.Set("Content-Type", detectContentType(r.Body))
	}

	for _, h := range r.Headers {
		key, err := vars.Interpolate(h.Key)
		if err != nil {
			return nil, "", err
		}
		value, err := vars.Interpolate(h.Value)
		if err != nil {
			return nil, "", err
		}
		if strings.EqualFold(key, "Content-Type") {
			req.Header.Set(key, value)
			continue
		}
		req.Header.Add(key, value)
	}

	authToken := ""
	cfg := r.Auth.Config
	if r.Auth.Type != "" {
		// Schemes implementing TokenSource (e.g. oauth2) acquire a token
		// before Apply; the resolved token is injected into a copy of the
		// config so the request's own config is never mutated.
		if s, ok := auth.Lookup(r.Auth.Type); ok {
			if ts, ok := s.(auth.TokenSource); ok {
				tok, err := c.acquireToken(ctx, ts, r.Auth.Config, vars)
				if err != nil {
					return nil, "", err
				}
				authToken = tok.AccessToken
				cfg = make(map[string]string, len(r.Auth.Config)+1)
				for k, v := range r.Auth.Config {
					cfg[k] = v
				}
				cfg["token"] = authToken
			}
		}
	}

	if err := auth.Apply(req, r.Auth.Type, cfg, vars); err != nil {
		return nil, "", err
	}

	return req, authToken, nil
}

// acquireToken resolves a token for a TokenSource scheme. When caching is
// enabled, the cache lock serializes the lookup+acquire+store critical section
// so concurrent requests for the same config do not double-acquire; the defer
// guarantees the lock is released even if acquisition panics or grows an early
// return later.
func (c *Client) acquireToken(ctx context.Context, s auth.TokenSource, cfg map[string]string, vars auth.Interpolator) (auth.Token, error) {
	if c.tokens == nil {
		return s.Token(ctx, cfg, vars)
	}
	c.tokens.mu.Lock()
	defer c.tokens.mu.Unlock()
	ts := auth.NewCachedTokenSource(s, c.tokens.store, auth.TokenCacheKey(c.tokens.root, cfg))
	return ts.Token(ctx, cfg, vars)
}

// forceRefreshToken invalidates and renews a cached token under the same lock
// as acquireToken, so a reactive 401 refresh cannot race a concurrent acquire
// on the same store entry.
func (c *Client) forceRefreshToken(ctx context.Context, s auth.TokenSource, cfg map[string]string, vars auth.Interpolator) error {
	if c.tokens == nil {
		return nil
	}
	c.tokens.mu.Lock()
	defer c.tokens.mu.Unlock()
	ts := auth.NewCachedTokenSource(s, c.tokens.store, auth.TokenCacheKey(c.tokens.root, cfg))
	_, err := ts.ForceRefresh(ctx, cfg, vars)
	return err
}

// detectContentType returns a best-effort content type for the request body
// when the user did not set one explicitly.
func detectContentType(body string) string {
	trimmed := strings.TrimSpace(body)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if json.Valid([]byte(trimmed)) {
			return "application/json"
		}
	}
	return "text/plain"
}
