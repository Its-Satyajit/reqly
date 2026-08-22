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

package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"hash"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

func decodeSegment(t *testing.T, s string) []byte {
	t.Helper()
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestJWTApplyHS256(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	err := jwtScheme{}.Apply(req, map[string]string{
		"secret":    "my-secret",
		"algorithm": "HS256",
		"claims":    `{"sub":"1234567890","name":"Reqly"}`,
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	hdr := req.Header.Get("Authorization")
	if !strings.HasPrefix(hdr, "Bearer ") {
		t.Fatalf("expected Bearer token, got %q", hdr)
	}
	token := strings.TrimPrefix(hdr, "Bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT segments, got %d", len(parts))
	}

	// Header must claim HS256.
	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(decodeSegment(t, parts[0]), &header); err != nil {
		t.Fatal(err)
	}
	if header.Alg != "HS256" || header.Typ != "JWT" {
		t.Fatalf("header: got %+v", header)
	}

	// Payload carries the configured claims.
	var payload map[string]any
	if err := json.Unmarshal(decodeSegment(t, parts[1]), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["sub"] != "1234567890" || payload["name"] != "Reqly" {
		t.Fatalf("payload: got %+v", payload)
	}

	// Signature verifies against the secret.
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, []byte("my-secret"))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		t.Fatal("signature does not verify")
	}
}

func TestJWTApplyAllAlgorithms(t *testing.T) {
	algos := map[string]func() hash.Hash{
		"HS256": sha256.New,
		"HS384": sha512.New384,
		"HS512": sha512.New,
	}
	for algo, newHash := range algos {
		t.Run(algo, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
			err := jwtScheme{}.Apply(req, map[string]string{
				"secret":    "s",
				"algorithm": algo,
				"claims":    `{"role":"admin"}`,
			}, variables.NewSet())
			if err != nil {
				t.Fatal(err)
			}
			token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
			parts := strings.Split(token, ".")
			sig, err := base64.RawURLEncoding.DecodeString(parts[2])
			if err != nil {
				t.Fatal(err)
			}
			mac := hmac.New(newHash, []byte("s"))
			mac.Write([]byte(parts[0] + "." + parts[1]))
			if !hmac.Equal(sig, mac.Sum(nil)) {
				t.Fatalf("%s signature does not verify", algo)
			}
		})
	}
}

func TestJWTApplyDefaultAlgorithm(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	err := jwtScheme{}.Apply(req, map[string]string{
		"secret": "s",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	parts := strings.Split(token, ".")
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(decodeSegment(t, parts[0]), &header); err != nil {
		t.Fatal(err)
	}
	if header.Alg != "HS256" {
		t.Fatalf("default algorithm: got %q, want HS256", header.Alg)
	}
}

func TestJWTApplyExpiresIn(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	err := jwtScheme{}.Apply(req, map[string]string{
		"secret":    "s",
		"claims":    `{"sub":"u1"}`,
		"expiresIn": "3600",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	parts := strings.Split(token, ".")
	var payload struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(decodeSegment(t, parts[1]), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Sub != "u1" {
		t.Fatalf("payload sub: got %q", payload.Sub)
	}
	// exp must be in the future (now + ~3600s).
	if payload.Exp <= 0 {
		t.Fatalf("expected exp claim, got %d", payload.Exp)
	}
}

func TestJWTApplyInterpolated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	vars := variables.NewSet()
	vars.Set(variables.ScopeRequest, "secret", "interpolated-secret")
	vars.Set(variables.ScopeRequest, "claims", `{"sub":"interp"}`)
	err := jwtScheme{}.Apply(req, map[string]string{
		"secret": "{{secret}}",
		"claims": "{{claims}}",
	}, vars)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	parts := strings.Split(token, ".")
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, []byte("interpolated-secret"))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		t.Fatal("signature does not verify with interpolated secret")
	}
}

func TestJWTApplyInterpolatesAlgorithmAndExpiry(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	vars := variables.NewSet()
	vars.Set(variables.ScopeRequest, "alg", "HS512")
	vars.Set(variables.ScopeRequest, "ttl", "120")
	err := jwtScheme{}.Apply(req, map[string]string{
		"secret":    "s",
		"algorithm": "{{alg}}",
		"expiresIn": "{{ttl}}",
	}, vars)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	parts := strings.Split(token, ".")
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(decodeSegment(t, parts[0]), &header); err != nil {
		t.Fatal(err)
	}
	if header.Alg != "HS512" {
		t.Fatalf("algorithm: got %q, want HS512", header.Alg)
	}
	var payload struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(decodeSegment(t, parts[1]), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Exp == 0 {
		t.Fatal("expected exp claim from interpolated expiresIn")
	}
}

func TestJWTApplyMissingSecret(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	s := jwtScheme{}
	if err := s.Apply(req, nil, variables.NewSet()); err == nil {
		t.Fatal("expected error when secret missing")
	}
}

func TestJWTApplyUnsupportedAlgorithm(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	s := jwtScheme{}
	err := s.Apply(req, map[string]string{
		"secret":    "s",
		"algorithm": "none",
	}, variables.NewSet())
	if err == nil {
		t.Fatal("expected error for unsupported algorithm")
	}
	if !strings.Contains(err.Error(), "none") {
		t.Fatalf("expected error to name the algorithm, got %v", err)
	}
}
