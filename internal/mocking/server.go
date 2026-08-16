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

package mocking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Server serves a mock HTTP API derived from an OpenAPI document. It matches
// incoming requests against the document's paths and responds with generated
// example bodies from the matched operation's responses.
type Server struct {
	doc       *openapi3.T
	delay     time.Duration
	failEvery int
	logger    *log.Logger
}

// Option configures a mock Server.
type Option func(*Server)

// WithDelay adds a fixed artificial latency before every response.
func WithDelay(d time.Duration) Option {
	return func(s *Server) { s.delay = d }
}

// WithFailureRate makes the server return a 500 error for roughly every nth
// request, which is useful for exercising retry and error-handling paths.
// Values of n <= 1 disable failure simulation.
func WithFailureRate(everyN int) Option {
	return func(s *Server) {
		if everyN > 1 {
			s.failEvery = everyN
		}
	}
}

// WithLogger sets the logger used for per-request logging. Nil disables logging.
func WithLogger(l *log.Logger) Option {
	return func(s *Server) { s.logger = l }
}

// NewServer builds a mock Server from an OpenAPI document.
func NewServer(doc *openapi3.T, opts ...Option) (*Server, error) {
	if doc == nil {
		return nil, errors.New("mocking: nil OpenAPI document")
	}
	if err := doc.Validate(context.Background(), openapi3.DisableExamplesValidation()); err != nil {
		return nil, fmt.Errorf("mocking: invalid OpenAPI document: %w", err)
	}
	s := &Server{doc: doc}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

// ServeHTTP matches the request against the document and writes the mock
// response for the matched operation.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	s.maybeLog("%s %s", r.Method, r.URL.Path)

	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-r.Context().Done():
			return
		}
	}

	if s.failEvery > 0 {
		counter.inc()
		if counter.n > 0 && counter.n%s.failEvery == 0 {
			writeJSON(w, http.StatusInternalServerError, "internal", "simulated server error")
			s.maybeLog("→ 500 simulated (after %s)", time.Since(start))
			return
		}
	}

	route := s.match(r)
	if route == nil {
		writeJSON(w, http.StatusNotFound, "not_found", "no matching path in OpenAPI document")
		s.maybeLog("→ 404 (after %s)", time.Since(start))
		return
	}

	operation := route.Operation
	if operation == nil {
		writeJSON(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed for path")
		s.maybeLog("→ 405 (after %s)", time.Since(start))
		return
	}

	status := pickResponseStatus(operation.Responses)
	body, contentType := s.responseBody(operation, status)

	if contentType == "" {
		contentType = "application/json"
	}

	if body == nil {
		body = []byte("null")
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		s.maybeLog("→ write failed: %v", err)
		return
	}
	s.maybeLog("→ %d (%s)", status, time.Since(start))
}

type failureCountKey struct{}

type requestCounter struct {
	n int
}

func (c *requestCounter) inc() { c.n++ }

var counter = &requestCounter{}

// match locates the first document path whose pattern matches the request.
// Path templates use {name} segments, e.g. /users/{id}.
func (s *Server) match(r *http.Request) *matchResult {
	reqPath := cleanPath(r.URL.Path)

	var best *matchResult
	paths := make([]string, 0, len(s.doc.Paths.Map()))
	for path := range s.doc.Paths.Map() {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool {
		return templateSpecificity(paths[i]) > templateSpecificity(paths[j])
	})

	for _, tmpl := range paths {
		params, ok := matchTemplate(tmpl, reqPath)
		if !ok {
			continue
		}
		item := s.doc.Paths.Value(tmpl)
		op := item.GetOperation(r.Method)
		if op == nil {
			if best == nil {
				best = &matchResult{Path: tmpl}
			}
			continue
		}
		return &matchResult{Path: tmpl, Operation: op, Params: params}
	}
	return best
}

// matchResult holds a matched path and the operation for the request method,
// or just the path when the method is not allowed on it.
type matchResult struct {
	Path      string
	Operation *openapi3.Operation
	Params    map[string]string
}

func cleanPath(p string) string {
	if p == "" || p[0] != '/' {
		return "/" + p
	}
	return p
}

// matchTemplate reports whether an OpenAPI path template matches a concrete
// request path, returning the captured path parameters. Non-template segments
// must match literally; {name} segments match any single non-empty segment.
func matchTemplate(tmpl, path string) (map[string]string, bool) {
	ts := strings.Split(tmpl, "/")
	ps := strings.Split(path, "/")
	if len(ts) != len(ps) {
		return nil, false
	}
	params := map[string]string{}
	for i := range ts {
		seg := ts[i]
		if seg == "" && ps[i] == "" {
			continue
		}
		if len(seg) >= 2 && seg[0] == '{' && seg[len(seg)-1] == '}' {
			name := seg[1 : len(seg)-1]
			if ps[i] == "" {
				return nil, false
			}
			params[name] = ps[i]
			continue
		}
		if seg != ps[i] {
			return nil, false
		}
	}
	return params, true
}

// templateSpecificity orders templates so that literal paths are tried before
// templated ones, and templates with more literal segments come first.
func templateSpecificity(tmpl string) int {
	score := 0
	for _, seg := range strings.Split(tmpl, "/") {
		if seg != "" && !(len(seg) >= 2 && seg[0] == '{' && seg[len(seg)-1] == '}') {
			score += 10
		}
	}
	return score
}

// pickResponseStatus selects which response to serve: the first 2xx code,
// falling back to "default" then the first defined status.
func pickResponseStatus(responses *openapi3.Responses) int {
	if responses == nil {
		return http.StatusOK
	}
	if code := responses.Status(http.StatusOK); code != nil {
		return http.StatusOK
	}
	codes := make([]int, 0, len(responses.Map()))
	for code := range responses.Map() {
		if c, err := strconv.Atoi(code); err == nil {
			codes = append(codes, c)
		}
	}
	sort.Ints(codes)
	if len(codes) == 0 {
		return http.StatusOK
	}
	for _, c := range codes {
		if c >= 200 && c < 300 {
			return c
		}
	}
	return codes[0]
}

// responseBody builds a mock response body and content type from an operation's
// chosen response. It prefers application/json and an explicit example, then a
// schema-generated example, and falls back to the raw content of the response.
func (s *Server) responseBody(operation *openapi3.Operation, status int) ([]byte, string) {
	responses := operation.Responses
	if responses == nil {
		return []byte("{}"), "application/json"
	}
	resp := responses.Status(status)
	if resp == nil || resp.Value == nil {
		if dflt := responses.Default(); dflt != nil {
			resp = dflt
		}
	}
	if resp == nil || resp.Value == nil {
		return []byte("{}"), "application/json"
	}

	content := resp.Value.Content
	if len(content) == 0 {
		return []byte("{}"), "application/json"
	}

	// Prefer application/json, otherwise the first declared media type.
	var mediaType string
	if _, ok := content["application/json"]; ok {
		mediaType = "application/json"
	} else {
		for mt := range content {
			mediaType = mt
			break
		}
	}
	mt := content[mediaType]

	if example := s.exampleValue(mt); example != nil {
		data, err := json.Marshal(example)
		if err == nil {
			return data, mediaType
		}
	}
	if mt.Schema != nil && mt.Schema.Value != nil {
		if data, err := json.Marshal(generateExample(mt.Schema.Value)); err == nil {
			return data, mediaType
		}
	}
	return []byte("{}"), mediaType
}

// exampleValue returns an explicit example from a media type if one is defined.
func (s *Server) exampleValue(mt *openapi3.MediaType) any {
	if mt == nil {
		return nil
	}
	if mt.Example != nil {
		return mt.Example
	}
	if len(mt.Examples) > 0 {
		keys := make([]string, 0, len(mt.Examples))
		for k := range mt.Examples {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if ref := mt.Examples[keys[0]]; ref != nil && ref.Value != nil {
			return ref.Value.Value
		}
	}
	return nil
}

func (s *Server) maybeLog(format string, args ...any) {
	if s.logger != nil {
		s.logger.Printf(format, args...)
	}
}

func writeJSON(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "message": message})
}
