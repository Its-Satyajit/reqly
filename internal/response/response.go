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
