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
	"fmt"
	"sync"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// RequestService exposes the request engine to front-ends. It is the shared
// application-service boundary used by the Desktop, CLI, and MCP. Front-ends
// should not depend on the transport directly.
type RequestService struct {
	client *request.Client
	// root is the workspace root this service is bound to ("" = unbound:
	// no token caching, no history recording). Run uses it for environment
	// resolution and history storage.
	root    string
	warning string

	histOnce   sync.Once
	historySvc *HistoryService
}

// NewRequestService returns a RequestService backed by a fresh request client
// without token caching.
func NewRequestService() *RequestService {
	return &RequestService{client: request.NewClient()}
}

// NewCachedRequestService returns a RequestService whose client caches OAuth
// tokens in store, scoped to root — the same store-backed wiring the CLI
// uses, so the desktop authenticates like the CLI. store may be nil, in which
// case no caching is configured.
func NewCachedRequestService(store secrets.Store, root string) *RequestService {
	client := request.NewClient()
	if store != nil {
		client = request.NewClient(request.WithTokenCache(store, root))
	}
	return &RequestService{client: client}
}

// SendResponse is the bridge-friendly result of executing a request. It uses
// plain JSON-friendly types so Wails and other front-ends can marshal it
// without special handling.
type SendResponse struct {
	StatusCode int                 `json:"statusCode"`
	StatusText string              `json:"statusText"`
	Proto      string              `json:"proto"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	DurationMS int64               `json:"durationMs"`
	Size       int64               `json:"size"`
	OK         bool                `json:"ok"`
	// Attempts reports how many sends the response took, including retries
	// (1 = no retry).
	Attempts int `json:"attempts"`
}

// Send executes the request and returns a SendResponse, or an error when the
// request could not be sent. An optional variable set is used for
// interpolation; when omitted, an empty set is used.
func (s *RequestService) Send(r request.Request, vars ...*variables.Set) (*SendResponse, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to send requests")
	}
	set := variables.NewSet()
	if len(vars) > 0 && vars[0] != nil {
		set = vars[0]
	}
	resp, err := s.client.Execute(context.Background(), &r, set)
	if err != nil {
		return nil, err
	}

	return &SendResponse{
		StatusCode: resp.StatusCode,
		StatusText: resp.StatusText,
		Proto:      resp.Proto,
		Headers:    resp.Headers,
		Body:       resp.Text(),
		DurationMS: resp.Duration.Milliseconds(),
		Size:       resp.Size,
		OK:         resp.OK(),
		Attempts:   resp.Attempts,
	}, nil
}
