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

package importer

import (
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
)

func TestParseCurlSimpleGet(t *testing.T) {
	req, err := ParseCurl(`curl https://api.example.com/users`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != request.MethodGet {
		t.Fatalf("method: got %q", req.Method)
	}
	if req.URL != "https://api.example.com/users" {
		t.Fatalf("url: got %q", req.URL)
	}
}

func TestParseCurlMethodAndHeaders(t *testing.T) {
	req, err := ParseCurl(`curl -X POST https://api.example.com/users -H 'Content-Type: application/json' -H "Authorization: Bearer abc"`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != request.MethodPost {
		t.Fatalf("method: got %q", req.Method)
	}
	if len(req.Headers) != 2 {
		t.Fatalf("headers: got %d", len(req.Headers))
	}
	if req.Headers[0].Key != "Content-Type" || req.Headers[0].Value != "application/json" {
		t.Fatalf("header[0]: got %+v", req.Headers[0])
	}
	if req.Headers[1].Key != "Authorization" || req.Headers[1].Value != "Bearer abc" {
		t.Fatalf("header[1]: got %+v", req.Headers[1])
	}
}

func TestParseCurlDataInfersPost(t *testing.T) {
	req, err := ParseCurl(`curl -d '{"name":"reqly"}' -H 'Content-Type: application/json' https://api.example.com/users`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != request.MethodPost {
		t.Fatalf("method: got %q, want POST inferred from -d", req.Method)
	}
	if req.Body != `{"name":"reqly"}` {
		t.Fatalf("body: got %q", req.Body)
	}
}

func TestParseCurlDataWithExplicitPut(t *testing.T) {
	req, err := ParseCurl(`curl -X PUT -d '{"name":"x"}' https://api.example.com/users/1`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != request.MethodPut {
		t.Fatalf("method: got %q", req.Method)
	}
	if req.Body != `{"name":"x"}` {
		t.Fatalf("body: got %q", req.Body)
	}
}

func TestParseCurlGetDataAsQuery(t *testing.T) {
	req, err := ParseCurl(`curl -G -d 'page=2&limit=10' https://api.example.com/users`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != request.MethodGet {
		t.Fatalf("method: got %q", req.Method)
	}
	if req.Body != "" {
		t.Fatalf("body: got %q, want empty", req.Body)
	}
	// Query params extracted and appended to URL.
	if len(req.Query) != 2 {
		t.Fatalf("query: got %d params", len(req.Query))
	}
	for _, p := range req.Query {
		switch p.Key {
		case "page":
			if p.Value != "2" {
				t.Fatalf("page: got %q", p.Value)
			}
		case "limit":
			if p.Value != "10" {
				t.Fatalf("limit: got %q", p.Value)
			}
		default:
			t.Fatalf("unexpected query param %q", p.Key)
		}
	}
}

func TestParseCurlBasicAuth(t *testing.T) {
	req, err := ParseCurl(`curl -u 'admin:secret' https://api.example.com/secure`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Auth.Type != "basic" {
		t.Fatalf("auth type: got %q", req.Auth.Type)
	}
	if req.Auth.Config["username"] != "admin" || req.Auth.Config["password"] != "secret" {
		t.Fatalf("auth config: got %+v", req.Auth.Config)
	}
}

func TestParseCurlUserAgentAndCookie(t *testing.T) {
	req, err := ParseCurl(`curl -A 'reqly/1.0' -b 'session=xyz' https://api.example.com/`)
	if err != nil {
		t.Fatal(err)
	}
	headers := map[string]string{}
	for _, h := range req.Headers {
		headers[h.Key] = h.Value
	}
	if headers["User-Agent"] != "reqly/1.0" {
		t.Fatalf("user-agent: got %q", headers["User-Agent"])
	}
	if headers["Cookie"] != "session=xyz" {
		t.Fatalf("cookie: got %q", headers["Cookie"])
	}
}

func TestParseCurlHead(t *testing.T) {
	req, err := ParseCurl(`curl -I https://api.example.com/status`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != request.MethodHead {
		t.Fatalf("method: got %q", req.Method)
	}
}

func TestParseCurlIgnoreNoiseFlags(t *testing.T) {
	req, err := ParseCurl(`curl -s -L -k --compressed -o out.txt https://api.example.com/`)
	if err != nil {
		t.Fatal(err)
	}
	if req.URL != "https://api.example.com/" {
		t.Fatalf("url: got %q", req.URL)
	}
}

func TestParseCurlUnsupportedFlag(t *testing.T) {
	if _, err := ParseCurl(`curl -F 'file=@a.txt' https://api.example.com/`); err == nil {
		t.Fatal("expected error for unsupported -F flag")
	}
}

func TestParseCurlUnterminatedQuote(t *testing.T) {
	if _, err := ParseCurl(`curl -H 'Content-Type: application/json https://api.example.com/`); err == nil {
		t.Fatal("expected error for unterminated quote")
	}
}

func TestParseCurlMissingValue(t *testing.T) {
	if _, err := ParseCurl(`curl -H`); err == nil {
		t.Fatal("expected error for missing flag value")
	}
}

func TestParseCurlNoURL(t *testing.T) {
	if _, err := ParseCurl(`curl -X GET`); err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestParseCurlMultipleDataJoined(t *testing.T) {
	req, err := ParseCurl(`curl -X POST -d 'a=1' -d 'b=2' https://api.example.com/`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Body != "a=1&b=2" {
		t.Fatalf("body: got %q", req.Body)
	}
}

func TestParseCurlQueryParamsFromURL(t *testing.T) {
	req, err := ParseCurl(`curl 'https://api.example.com/users?role=admin&active=true'`)
	if err != nil {
		t.Fatal(err)
	}
	if len(req.Query) != 2 {
		t.Fatalf("query: got %d", len(req.Query))
	}
}
