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

package main

import (
	"testing"
)

// tokenFixture is an unsigned HS256-shaped JWT with exp far in the future.
const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.sig"

func TestJwtDecodeReturnsHeaderPayloadExpiry(t *testing.T) {
	svc := &AppService{}
	tok, err := svc.JwtDecode(validToken)
	if err != nil {
		t.Fatalf("JwtDecode: %v", err)
	}
	if tok.Alg != "HS256" {
		t.Errorf("Alg = %q", tok.Alg)
	}
	claimValue := func(claims []JwtClaim, key string) (string, bool) {
		for _, c := range claims {
			if c.Key == key {
				return c.Value, true
			}
		}
		return "", false
	}
	if typ, ok := claimValue(tok.Header, "typ"); !ok || typ != "JWT" {
		t.Errorf("header typ = %q (found %v)", typ, ok)
	}
	if sub, ok := claimValue(tok.Payload, "sub"); !ok || sub != "123" {
		t.Errorf("payload sub = %q (found %v)", sub, ok)
	}
}

func TestJwtDecodeRejectsGarbage(t *testing.T) {
	svc := &AppService{}
	if _, err := svc.JwtDecode("not-a-jwt"); err == nil {
		t.Fatal("garbage accepted")
	}
}

func TestJwtFromAuthHeader(t *testing.T) {
	svc := &AppService{}
	got, ok := svc.JwtFromAuthHeader("Bearer " + validToken)
	if !ok || got != validToken {
		t.Fatalf("got=%q ok=%v", got, ok)
	}
	if _, ok := svc.JwtFromAuthHeader("Basic abc"); ok {
		t.Error("non-bearer header accepted")
	}
}

func TestJwtDecodeExpiryStatusShape(t *testing.T) {
	svc := &AppService{}
	tok, err := svc.JwtDecode(validToken)
	if err != nil {
		t.Fatal(err)
	}
	if tok.Expiry.Status != "no_expiry" {
		t.Errorf("Status = %q, want no_expiry (fixture has no exp)", tok.Expiry.Status)
	}
}
