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

// Package sse provides a Server-Sent Events (SSE) client for consuming
// continuous streams of server-generated events over HTTP.
package sse

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Event is a single server-sent event.
type Event struct {
	// ID is the event id (event's "id:" field), if any.
	ID string
	// Name is the event type ("event:" field); "" means "message".
	Name string
	// Data is the concatenated data lines ("data:" field), joined with newlines.
	Data string
	// Retry is the suggested reconnection time, if the server sent "retry:".
	Retry time.Duration
	// At is when the event was received.
	At time.Time
}

// Status is the connection lifecycle state.
type Status string

const (
	// StatusDisconnected means no stream is open.
	StatusDisconnected Status = "disconnected"
	// StatusConnecting means Connect is in progress.
	StatusConnecting Status = "connecting"
	// StatusConnected means the stream is open.
	StatusConnected Status = "connected"
	// StatusClosing means Close was requested.
	StatusClosing Status = "closing"
)

// Client consumes a Server-Sent Events stream. Use Next to read events and
// Close to tear down the connection. Safe for concurrent use.
type Client struct {
	url     string
	headers http.Header
	client  *http.Client

	mu       sync.RWMutex
	resp     *http.Response
	reader   *bufio.Reader
	status   Status
	statusFn func(Status)
}

// Option configures a Client.
type Option func(*Client)

// WithHeaders sets extra headers for the request.
func WithHeaders(headers http.Header) Option {
	return func(c *Client) { c.headers = headers }
}

// WithHTTPClient replaces the underlying HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.client = client }
}

// WithStatusHandler receives every status transition.
func WithStatusHandler(fn func(Status)) Option {
	return func(c *Client) { c.statusFn = fn }
}

// NewClient returns an SSE client for the given URL.
func NewClient(url string, opts ...Option) *Client {
	c := &Client{
		url:     url,
		headers: http.Header{},
		client:  &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// URL returns the stream URL.
func (c *Client) URL() string { return c.url }

// Status returns the current connection state.
func (c *Client) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

func (c *Client) setStatus(s Status) {
	c.mu.Lock()
	c.status = s
	c.mu.Unlock()
	if c.statusFn != nil {
		c.statusFn(s)
	}
}

// Connect opens the stream. It returns as soon as the HTTP response headers
// arrive; events are read with Next.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.RLock()
	if c.status == StatusConnected {
		c.mu.RUnlock()
		return fmt.Errorf("sse: already connected")
	}
	c.mu.RUnlock()

	c.setStatus(StatusConnecting)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		c.setStatus(StatusDisconnected)
		return fmt.Errorf("sse: build request: %w", err)
	}
	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		c.setStatus(StatusDisconnected)
		return fmt.Errorf("sse connect %s: %w", c.url, err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		c.setStatus(StatusDisconnected)
		return fmt.Errorf("sse connect %s: unexpected status %d", c.url, resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/event-stream") {
		resp.Body.Close()
		c.setStatus(StatusDisconnected)
		return fmt.Errorf("sse connect %s: not an event stream (Content-Type %q)", c.url, ct)
	}

	c.mu.Lock()
	c.resp = resp
	c.reader = bufio.NewReader(resp.Body)
	c.mu.Unlock()
	c.setStatus(StatusConnected)
	return nil
}

// Next blocks until the next event is parsed, or returns an error when the
// stream ends or the context is cancelled.
func (c *Client) Next(ctx context.Context) (Event, error) {
	c.mu.RLock()
	reader := c.reader
	c.mu.RUnlock()
	if reader == nil {
		return Event{}, fmt.Errorf("sse: not connected")
	}

	var id, name string
	var data []string
	var retry time.Duration

	for {
		line, err := readLine(ctx, reader)
		if err != nil {
			return Event{}, err
		}
		if line == "" {
			// Blank line dispatches the event.
			if len(data) > 0 || name != "" {
				return Event{
					ID:    id,
					Name:  name,
					Data:  strings.Join(data, "\n"),
					Retry: retry,
					At:    time.Now(),
				}, nil
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			// Comment line, ignore.
			continue
		}

		field, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimPrefix(value, " ")

		switch field {
		case "id":
			id = value
		case "event":
			name = value
		case "data":
			data = append(data, value)
		case "retry":
			if ms, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				retry = time.Duration(ms) * time.Millisecond
			}
		}
	}
}

// readLine reads a line, honoring context cancellation. It treats EOF as io.EOF.
func readLine(ctx context.Context, r *bufio.Reader) (string, error) {
	type result struct {
		line string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		line, err := r.ReadString('\n')
		ch <- result{line: strings.TrimRight(line, "\r\n"), err: err}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-ch:
		return res.line, res.err
	}
}

// Close closes the underlying response body.
func (c *Client) Close() error {
	c.setStatus(StatusClosing)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.resp != nil {
		err := c.resp.Body.Close()
		c.resp = nil
		c.reader = nil
		c.status = StatusDisconnected
		return err
	}
	c.status = StatusDisconnected
	return nil
}
