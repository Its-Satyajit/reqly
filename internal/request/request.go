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

// Request defines a single API request. The engine layer (transport, protocol
// dispatch, authentication, variable interpolation, scripting) builds on this
// definition.
//
// The transport remains abstract so additional protocols (WebSocket, gRPC,
// SSE, MQTT, ...) can be added without changing the application architecture.
type Request struct {
	ID     string `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Method Method `json:"method,omitempty" yaml:"method,omitempty"`
	URL    string `json:"url,omitempty" yaml:"url,omitempty"`

	Headers []Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Query   []Parameter `json:"query,omitempty" yaml:"query,omitempty"`
	Body    string      `json:"body,omitempty" yaml:"body,omitempty"`

	Auth       Auth       `json:"auth,omitempty" yaml:"auth,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty" yaml:"pagination,omitempty"`
	Timeout    int64      `json:"timeout,omitempty" yaml:"timeout,omitempty"` // milliseconds; 0 means "no explicit timeout"
}

// Pagination configures iterative fetching for a paginated endpoint.
type Pagination struct {
	Strategy      string `json:"strategy" yaml:"strategy"` // page|offset|cursor|link-header
	PageParam     string `json:"pageParam,omitempty" yaml:"pageParam,omitempty"`
	PageSizeParam string `json:"pageSizeParam,omitempty" yaml:"pageSizeParam,omitempty"`
	OffsetParam   string `json:"offsetParam,omitempty" yaml:"offsetParam,omitempty"`
	LimitParam    string `json:"limitParam,omitempty" yaml:"limitParam,omitempty"`
	CursorParam   string `json:"cursorParam,omitempty" yaml:"cursorParam,omitempty"`
	NextPath      string `json:"nextPath,omitempty" yaml:"nextPath,omitempty"`
	MaxPages      int    `json:"maxPages,omitempty" yaml:"maxPages,omitempty"`
	PageSize      int    `json:"pageSize,omitempty" yaml:"pageSize,omitempty"`
	Limit         int    `json:"limit,omitempty" yaml:"limit,omitempty"`
}

// Method is an HTTP method.
type Method string

const (
	MethodGet     Method = "GET"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodPatch   Method = "PATCH"
	MethodDelete  Method = "DELETE"
	MethodHead    Method = "HEAD"
	MethodOptions Method = "OPTIONS"
)

// Header is a single request header.
type Header struct {
	Key   string `json:"key" yaml:"key,omitempty"`
	Value string `json:"value" yaml:"value,omitempty"`
}

// Parameter is a query or path parameter.
type Parameter struct {
	Key   string `json:"key" yaml:"key,omitempty"`
	Value string `json:"value" yaml:"value,omitempty"`
}

// Auth describes the authentication configuration attached to a request.
// Implementations live in the auth package and are dispatched by type.
type Auth struct {
	Type   string            `json:"type,omitempty" yaml:"type,omitempty"`
	Config map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
}
