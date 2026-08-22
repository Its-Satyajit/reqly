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

package response

import "time"

// Response models an executed API response: status, headers, body, timing,
// and size. It is shared by the Desktop, CLI, and MCP front-ends.
type Response struct {
	// StatusCode is the HTTP status code (e.g. 200).
	StatusCode int
	// StatusText is the status text (e.g. "OK").
	StatusText string
	// Proto is the HTTP protocol version (e.g. "HTTP/1.1").
	Proto string
	// Headers are the response headers.
	Headers map[string][]string
	// Body is the raw response body.
	Body []byte
	// Duration is the total request duration.
	Duration time.Duration
	// Size is the body size in bytes.
	Size int64
	// AuthToken is the resolved access token used for this request, when an
	// auth scheme acquired one (e.g. oauth2). It exists so callers can mask
	// it in output; it is never serialized.
	AuthToken string `json:"-"`
}

// OK reports whether the response indicates success (2xx).
func (r *Response) OK() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// Text returns the response body as a string.
func (r *Response) Text() string {
	return string(r.Body)
}
