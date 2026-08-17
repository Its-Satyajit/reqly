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
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// Interpolator resolves {{key}} placeholders in request fields. It is the
// variables.Set interface, kept small so the engine does not depend on the
// full variables package internals.
type Interpolator interface {
	Interpolate(input string) (string, error)
}

// Client executes Request values over HTTP and returns response.Response
// values. It is the shared engine used by the Desktop, CLI, and MCP.
type Client struct {
	http    *http.Client
	timeout time.Duration
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
func (c *Client) Execute(ctx context.Context, r *Request, vars Interpolator) (*response.Response, error) {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(r.Timeout)*time.Millisecond)
		defer cancel()
	}

	req, err := c.build(ctx, r, vars)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	// Digest (and other challenge-based schemes) respond to a 401
	// WWW-Authenticate challenge by computing credentials and retrying once.
	// The retry is bounded to a single challenge/response round-trip.
	if resp.StatusCode == http.StatusUnauthorized && r.Auth.Type != "" {
		if s, ok := auth.Lookup(r.Auth.Type); ok {
			if cs, ok := s.(auth.ChallengedScheme); ok {
				if challenge := resp.Header.Get("WWW-Authenticate"); challenge != "" {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					retryReq, err := c.build(ctx, r, vars)
					if err != nil {
						return nil, err
					}
					if err := cs.Challenge(retryReq, challenge, r.Auth.Config, vars); err != nil {
						return nil, err
					}
					resp, err = c.http.Do(retryReq)
					if err != nil {
						return nil, err
					}
				}
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
	}, nil
}

// build constructs a net/http Request from the model, applying interpolation,
// query parameters, headers, body, and authentication.
func (c *Client) build(ctx context.Context, r *Request, vars Interpolator) (*http.Request, error) {
	if vars == nil {
		vars = variables.NewSet()
	}

	method := r.Method
	if method == "" {
		method = MethodGet
	}

	rawURL, err := vars.Interpolate(r.URL)
	if err != nil {
		return nil, err
	}
	if rawURL == "" {
		return nil, errors.New("request has no URL")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	q := u.Query()
	for _, p := range r.Query {
		key, err := vars.Interpolate(p.Key)
		if err != nil {
			return nil, err
		}
		value, err := vars.Interpolate(p.Value)
		if err != nil {
			return nil, err
		}
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	var body io.Reader
	if r.Body != "" {
		interpolated, err := vars.Interpolate(r.Body)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(interpolated)
	}

	req, err := http.NewRequestWithContext(ctx, string(method), u.String(), body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", detectContentType(r.Body))
	}

	for _, h := range r.Headers {
		key, err := vars.Interpolate(h.Key)
		if err != nil {
			return nil, err
		}
		value, err := vars.Interpolate(h.Value)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(key, "Content-Type") {
			req.Header.Set(key, value)
			continue
		}
		req.Header.Add(key, value)
	}

	if err := auth.Apply(req, r.Auth.Type, r.Auth.Config, vars); err != nil {
		return nil, err
	}

	return req, nil
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
