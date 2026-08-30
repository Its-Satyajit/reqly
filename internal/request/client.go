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

package request

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
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
	// onRetry, when set, is notified before each automatic retry wait.
	onRetry func(RetryEvent)
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

// WithOnRetry registers a callback invoked before each automatic retry wait,
// giving callers visibility into transient failures without parsing output.
func WithOnRetry(fn func(RetryEvent)) Option {
	return func(c *Client) {
		c.onRetry = fn
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

// ExecuteWithOnRetry behaves like Execute but additionally notifies the given
// observer of automatic retries. The observer travels with the call, so
// concurrent Executes sharing one Client never race over it; a client-level
// observer set via WithOnRetry fires too.
func (c *Client) ExecuteWithOnRetry(ctx context.Context, r *Request, vars auth.Interpolator, onRetry func(RetryEvent)) (*response.Response, error) {
	return c.execute(ctx, r, vars, onRetry)
}

// Execute runs a Request and returns the response. Variables are interpolated
// into the URL, headers, query parameters, and body before sending.
//
// When the request declares a Retry policy, failed attempts — network errors,
// or responses whose status is in the policy's retry set — are automatically
// re-sent with computed backoff (or the server's Retry-After) until Count is
// exhausted. Each attempt includes any auth challenge/refresh round-trips;
// those never consume retry budget. The returned response reports how many
// attempts it took in Attempts.
func (c *Client) Execute(ctx context.Context, r *Request, vars auth.Interpolator) (*response.Response, error) {
	return c.execute(ctx, r, vars, nil)
}

func (c *Client) execute(ctx context.Context, r *Request, vars auth.Interpolator, onRetry func(RetryEvent)) (*response.Response, error) {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(r.Timeout)*time.Millisecond)
		defer cancel()
	}

	var policy *Retry
	totalAttempts := 1
	if r.Retry != nil && r.Retry.Count > 0 {
		policy = r.Retry.normalized()
		totalAttempts = policy.Count + 1
	}

	for attempt := 1; ; attempt++ {
		resp, err := c.sendOnce(ctx, r, vars)
		if resp != nil {
			resp.Attempts = attempt
		}
		if err == nil && (policy == nil || !policy.retryable(resp.StatusCode)) {
			return resp, nil
		}
		if err != nil {
			if isContextErr(err) || policy == nil {
				return nil, err
			}
		}
		if attempt >= totalAttempts {
			return resp, err
		}

		delay := policy.delayFor(resp, attempt, time.Now())
		notify := onRetry
		if notify == nil {
			notify = c.onRetry
		}
		if notify != nil {
			event := RetryEvent{
				Attempt:       attempt,
				TotalAttempts: totalAttempts,
				Delay:         delay,
			}
			if err != nil {
				event.Err = err
			} else {
				event.StatusCode = resp.StatusCode
			}
			notify(event)
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return resp, ctx.Err()
		case <-timer.C:
		}
	}
}

// transportForRequest clones the client's transport for per-request proxy/TLS (M47).
func (c *Client) transportForRequest(r *Request, vars auth.Interpolator) (*http.Transport, error) {
	base := http.DefaultTransport.(*http.Transport).Clone()
	// Proxy
	if r.Proxy != "" {
		proxyStr, err := vars.Interpolate(r.Proxy)
		if err != nil {
			return nil, fmt.Errorf("proxy: %w", err)
		}
		if proxyStr != "" {
			u, err := url.Parse(proxyStr)
			if err != nil {
				return nil, fmt.Errorf("proxy: invalid URL %q: %w", proxyStr, err)
			}
			base.Proxy = http.ProxyURL(u)
		}
	}
	// TLS
	if r.TLS != nil {
		tlsCfg := &tls.Config{}
		if r.TLS.InsecureSkipVerify {
			tlsCfg.InsecureSkipVerify = true
		}
		if r.TLS.CAFile != "" {
			caPath, err := vars.Interpolate(r.TLS.CAFile)
			if err != nil {
				return nil, fmt.Errorf("tls: caFile: %w", err)
			}
			pem, err := os.ReadFile(caPath)
			if err != nil {
				return nil, fmt.Errorf("tls: caFile %q: %w", caPath, err)
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(pem) {
				return nil, fmt.Errorf("tls: caFile %q: failed to parse PEM", caPath)
			}
			tlsCfg.RootCAs = pool
		}
		if tlsCfg.InsecureSkipVerify || tlsCfg.RootCAs != nil {
			base.TLSClientConfig = tlsCfg
		}
	}
	// DisableKeepAlives
	if r.DisableKeepAlives {
		base.DisableKeepAlives = true
	}
	// HTTPVersion ALPN / transport preferences
	if r.HTTPVersion != "" && r.HTTPVersion != "auto" {
		if base.TLSClientConfig == nil {
			base.TLSClientConfig = &tls.Config{}
		}
		switch strings.ToLower(r.HTTPVersion) {
		case "http1.1":
			base.TLSClientConfig.NextProtos = []string{"http/1.1"}
			base.ForceAttemptHTTP2 = false
		case "http2":
			base.TLSClientConfig.NextProtos = []string{"h2", "http/1.1"}
			base.ForceAttemptHTTP2 = true
		}
	}
	return base, nil
}

// sendOnce performs exactly one send of the request: build, transport,
// challenge-based auth handling, and body read. It is the unit a retry
// policy counts; auth re-sends inside it stay within one attempt.
func (c *Client) sendOnce(ctx context.Context, r *Request, vars auth.Interpolator) (*response.Response, error) {
	req, authToken, err := c.build(ctx, r, vars)
	if err != nil {
		return nil, err
	}

	// Per-request transport (M47/M56) — when proxy/TLS/HTTPVersion/DisableKeepAlives set to avoid alloc.
	var httpClient *http.Client = c.http
	needsTransport := r.Proxy != "" || r.TLS != nil || r.DisableKeepAlives || (r.HTTPVersion != "" && r.HTTPVersion != "auto")
	needsNoFollow := r.FollowRedirects != nil && !*r.FollowRedirects
	if needsTransport || needsNoFollow {
		tr, err := c.transportForRequest(r, vars)
		if err != nil && needsTransport {
			return nil, err
		}
		if !needsTransport {
			// Clone current transport when only redirect behavior changes.
			if c.http.Transport != nil {
				if base, ok := c.http.Transport.(*http.Transport); ok {
					tr = base.Clone()
				}
			}
			if tr == nil {
				tr = http.DefaultTransport.(*http.Transport).Clone()
			}
		}
		httpClient = &http.Client{Transport: tr, Timeout: c.http.Timeout}
		if needsNoFollow {
			httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
	}

	start := time.Now()
	var traceMu sync.Mutex
	var dnsStart, dnsDone, connectStart, connectDone, tlsStart, tlsDone, gotConn, wroteRequest, firstByte time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			traceMu.Lock()
			dnsStart = time.Now()
			traceMu.Unlock()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			traceMu.Lock()
			dnsDone = time.Now()
			traceMu.Unlock()
		},
		ConnectStart: func(network, addr string) {
			traceMu.Lock()
			connectStart = time.Now()
			traceMu.Unlock()
		},
		ConnectDone: func(network, addr string, err error) {
			traceMu.Lock()
			connectDone = time.Now()
			traceMu.Unlock()
		},
		TLSHandshakeStart: func() {
			traceMu.Lock()
			tlsStart = time.Now()
			traceMu.Unlock()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			traceMu.Lock()
			tlsDone = time.Now()
			traceMu.Unlock()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			traceMu.Lock()
			gotConn = time.Now()
			traceMu.Unlock()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			traceMu.Lock()
			wroteRequest = time.Now()
			traceMu.Unlock()
		},
		GotFirstResponseByte: func() {
			traceMu.Lock()
			firstByte = time.Now()
			traceMu.Unlock()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	resp, err := httpClient.Do(req)
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
				resp, err = httpClient.Do(retryReq)
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
				resp, err = httpClient.Do(retryReq)
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

	duration := time.Since(start)
	// httptrace synthesis per exporter/har timings
	var timings *response.Timings
	traceMu.Lock()
	if !dnsStart.IsZero() || !connectStart.IsZero() || !firstByte.IsZero() {
		timings = &response.Timings{}
		if !dnsStart.IsZero() && !dnsDone.IsZero() {
			timings.DNS = dnsDone.Sub(dnsStart).Milliseconds()
		}
		if !connectStart.IsZero() && !connectDone.IsZero() {
			timings.Connect = connectDone.Sub(connectStart).Milliseconds()
		}
		if !tlsStart.IsZero() && !tlsDone.IsZero() {
			timings.TLS = tlsDone.Sub(tlsStart).Milliseconds()
		}
		if !gotConn.IsZero() && !wroteRequest.IsZero() {
			timings.Request = wroteRequest.Sub(gotConn).Milliseconds()
		}
		if !wroteRequest.IsZero() && !firstByte.IsZero() {
			timings.Server = firstByte.Sub(wroteRequest).Milliseconds()
		}
		if !firstByte.IsZero() {
			timings.Response = duration.Milliseconds() - firstByte.Sub(start).Milliseconds()
			timings.Transfer = duration.Milliseconds() - timings.DNS - timings.Connect - timings.TLS - timings.Request - timings.Server - timings.Response
			if timings.Transfer < 0 {
				timings.Transfer = 0
			}
		}
	}
	traceMu.Unlock()
	return &response.Response{
		StatusCode: resp.StatusCode,
		StatusText: http.StatusText(resp.StatusCode),
		Proto:      resp.Proto,
		Headers:    resp.Header,
		Body:       body,
		Duration:   duration,
		Size:       int64(len(body)),
		AuthToken:  authToken,
		Timings:    timings,
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
