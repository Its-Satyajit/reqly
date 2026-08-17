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

package core

import (
	"context"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// RequestService exposes the request engine to front-ends. It is the shared
// application-service boundary used by the Desktop, CLI, and MCP. Front-ends
// should not depend on the transport directly.
type RequestService struct {
	client *request.Client
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
}

// Send executes the request and returns a SendResponse, or an error when the
// request could not be sent. An optional variable set is used for
// interpolation; when omitted, an empty set is used.
func (s *RequestService) Send(r request.Request, vars ...*variables.Set) (*SendResponse, error) {
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
	}, nil
}
