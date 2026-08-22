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

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

func newVars(t *testing.T, set func(*variables.Set)) *variables.Set {
	t.Helper()
	v := variables.NewSet()
	if set != nil {
		set(v)
	}
	return v
}

func TestApplyBearer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	err := auth.Apply(req, "bearer", map[string]string{"token": "abc123"}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer abc123" {
		t.Fatalf("expected 'Bearer abc123', got %q", got)
	}
}

func TestApplyBearerInterpolated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	vars := newVars(t, func(v *variables.Set) {
		v.Set(variables.ScopeRequest, "token", "tok-xyz")
	})
	err := auth.Apply(req, "bearer", map[string]string{"token": "{{token}}"}, vars)
	if err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok-xyz" {
		t.Fatalf("expected 'Bearer tok-xyz', got %q", got)
	}
}

func TestApplyBasic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	err := auth.Apply(req, "basic", map[string]string{
		"username": "user",
		"password": "pass",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	user, pass, ok := req.BasicAuth()
	if !ok || user != "user" || pass != "pass" {
		t.Fatalf("expected user/pass, got %q/%q ok=%v", user, pass, ok)
	}
}

func TestApplyAPIKeyInHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	err := auth.Apply(req, "apikey", map[string]string{
		"key":   "X-API-Key",
		"value": "k-123",
		"in":    "header",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("X-API-Key"); got != "k-123" {
		t.Fatalf("expected 'k-123', got %q", got)
	}
}

func TestApplyAPIKeyInQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/?existing=1", nil)
	err := auth.Apply(req, "apikey", map[string]string{
		"key":   "api_key",
		"value": "q-456",
		"in":    "query",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()
	if got := q.Get("api_key"); got != "q-456" {
		t.Fatalf("expected 'q-456', got %q", got)
	}
	if got := q.Get("existing"); got != "1" {
		t.Fatalf("expected existing query param preserved, got %q", got)
	}
}

func TestApplyNone(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	req.Header.Set("Authorization", "Bearer inherited")
	err := auth.Apply(req, "none", nil, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer inherited" {
		t.Fatalf("none must not add or remove auth, got %q", got)
	}
}

func TestApplyEmptyTypeIsNoOp(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err := auth.Apply(req, "", nil, variables.NewSet()); err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatalf("expected no Authorization header, got %q", got)
	}
}

func TestApplyUnknownTypeErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	err := auth.Apply(req, "ntlm", nil, variables.NewSet())
	if err == nil {
		t.Fatal("expected error for unknown auth type")
	}
	if !strings.Contains(err.Error(), "ntlm") {
		t.Fatalf("expected error to name the unknown type, got %v", err)
	}
}

func TestApplyBearerMissingTokenErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err := auth.Apply(req, "bearer", nil, variables.NewSet()); err == nil {
		t.Fatal("expected error when bearer token missing")
	}
}

func TestApplyBasicMissingCredentialsErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err := auth.Apply(req, "basic", nil, variables.NewSet()); err == nil {
		t.Fatal("expected error when basic username/password missing")
	}
	if err := auth.Apply(req, "basic", map[string]string{"username": "u"}, variables.NewSet()); err == nil {
		t.Fatal("expected error when basic password missing")
	}
}

func TestApplyAPIKeyMissingValueErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err := auth.Apply(req, "apikey", map[string]string{"key": "X-Key"}, variables.NewSet()); err == nil {
		t.Fatal("expected error when apikey value missing")
	}
}

func TestApplyAPIKeyInvalidLocationErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err := auth.Apply(req, "apikey", map[string]string{"key": "X-Key", "value": "v", "in": "cookie"}, variables.NewSet()); err == nil {
		t.Fatal("expected error when apikey in is not header or query")
	}
}
