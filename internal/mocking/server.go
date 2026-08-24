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
	// routes are manual overrides checked before spec matching; nil doc with
	// routes only is valid (spec-less mock server).
	routes []Route
}

// Route is a manually defined mock response, matched before the OpenAPI
// document. Path is matched literally (no templates); Method "" matches any
// method. Disabled routes are skipped.
type Route struct {
	Method  string            `json:"method,omitempty" yaml:"method,omitempty"`
	Path    string            `json:"path" yaml:"path"`
	Status  int               `json:"status" yaml:"status"`
	Body    string            `json:"body,omitempty" yaml:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Enabled bool              `json:"enabled" yaml:"enabled"`
}

// WithRoutes adds manual routes served before the OpenAPI document. A nil
// document combined with routes builds a spec-less mock server.
func WithRoutes(routes []Route) Option {
	return func(s *Server) { s.routes = append(s.routes, routes...) }
}

// Option configures a mock Server.
type Option func(*Server)

// WithDelay adds a fixed artificial latency before every response.
func WithDelay(d time.Duration) Option {
	return func(s *Server) { s.delay = d }
}

// WithFailureRate makes the server return a 500 error for roughly every nth
// request, which is useful for exercising retry and error-handling paths.
// Values of everyN less than or equal to 1 leave failure simulation unchanged.
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

// NewServer builds a mock server from an OpenAPI document. The document may
// be nil when manual routes are supplied via WithRoutes.
func NewServer(doc *openapi3.T, opts ...Option) (*Server, error) {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	if doc == nil {
		if len(s.routes) == 0 {
			return nil, errors.New("mocking: nil OpenAPI document without manual routes")
		}
	} else {
		if err := doc.Validate(context.Background(), openapi3.DisableExamplesValidation()); err != nil {
			return nil, fmt.Errorf("mocking: invalid OpenAPI document: %w", err)
		}
		s.doc = doc
	}
	return s, nil
}

// serveRoute writes one manual route response; ok is false when no enabled
// route matches the request's path and method.
func (s *Server) serveRoute(w http.ResponseWriter, r *http.Request) bool {
	for _, rt := range s.routes {
		if !rt.Enabled {
			continue
		}
		if rt.Path != r.URL.Path {
			continue
		}
		if rt.Method != "" && !strings.EqualFold(rt.Method, r.Method) {
			continue
		}
		status := rt.Status
		if status == 0 {
			status = http.StatusOK
		}
		contentType := "application/json"
		for k, v := range rt.Headers {
			if strings.EqualFold(k, "Content-Type") {
				contentType = v
				continue
			}
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rt.Body))
		return true
	}
	return false
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

// cleanPath ensures a path begins with a slash.
func cleanPath(p string) string {
	if p == "" || p[0] != '/' {
		return "/" + p
	}
	return p
}

// matchTemplate reports whether an OpenAPI path template matches a concrete
// request path, returning the captured path parameters. Non-template segments
// matchTemplate matches a path against a template and captures values from parameter segments.
// It returns the captured parameters and whether the path matches the template.
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
// templateSpecificity scores a path template based on its literal segments.
// Each literal segment contributes 10 points.
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
// pickResponseStatus selects the HTTP status code for an OpenAPI response set.
// It prefers 200, then the lowest 2xx status, then the lowest numeric status, and
// defaults to 200 when no responses or numeric status codes are defined.
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

// writeJSON writes a JSON error response with the specified status, error code, and message.
func writeJSON(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "message": message})
}
