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

package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func encodeSegment(v any) string {
	b, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(b)
}

func makeToken(header, payload map[string]any, sig string) string {
	return encodeSegment(header) + "." + encodeSegment(payload) + "." + sig
}

func TestDecode_HS256(t *testing.T) {
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{"sub": "123", "name": "Reqly"}
	token := makeToken(header, payload, "sig")
	Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { Now = time.Now })
	tok, err := Decode(token)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if tok.Header["alg"] != "HS256" {
		t.Fatalf("alg: got %v", tok.Header["alg"])
	}
	if tok.Payload["sub"] != "123" {
		t.Fatalf("sub: got %v", tok.Payload["sub"])
	}
	if tok.Alg != "HS256" {
		t.Fatalf("Alg: got %q", tok.Alg)
	}
}

func TestDecode_AnyAlgNoVerify(t *testing.T) {
	for _, alg := range []string{"HS256", "RS256", "ES256", "none"} {
		header := map[string]any{"alg": alg}
		payload := map[string]any{"sub": "u1"}
		sig := "sig"
		if alg == "none" {
			sig = ""
		}
		token := makeToken(header, payload, sig)
		tok, err := Decode(token)
		if err != nil {
			t.Fatalf("alg %s: %v", alg, err)
		}
		if tok.Alg != alg {
			t.Fatalf("alg %s: got %q", alg, tok.Alg)
		}
	}
}

func TestDecode_BearerPrefixAndWhitespace(t *testing.T) {
	header := map[string]any{"alg": "HS256"}
	payload := map[string]any{"sub": "u1"}
	token := makeToken(header, payload, "sig")
	withBearer := "Bearer " + token
	tok, err := Decode("  " + withBearer + "  ")
	if err != nil {
		t.Fatalf("Bearer prefix: %v", err)
	}
	if tok.Payload["sub"] != "u1" {
		t.Fatalf("payload: got %v", tok.Payload)
	}
	// lower-case bearer
	tok2, err := Decode("bearer " + token)
	if err != nil {
		t.Fatalf("bearer lower: %v", err)
	}
	if tok2.Payload["sub"] != "u1" {
		t.Fatalf("bearer lower payload")
	}
}

func TestDecode_Expired(t *testing.T) {
	Now = func() time.Time { return time.Unix(2000, 0).UTC() }
	t.Cleanup(func() { Now = time.Now })
	header := map[string]any{"alg": "HS256"}
	payload := map[string]any{"sub": "u1", "exp": float64(1000)}
	token := makeToken(header, payload, "sig")
	tok, err := Decode(token)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if tok.Expiry.Status != "expired" {
		t.Fatalf("status: got %q want expired", tok.Expiry.Status)
	}
	if tok.Expiry.Remaining != -1000 {
		t.Fatalf("remaining: got %d want -1000", tok.Expiry.Remaining)
	}
	if tok.Expiry.Exp == nil || *tok.Expiry.Exp != 1000 {
		t.Fatalf("exp: got %v", tok.Expiry.Exp)
	}
}

func TestDecode_NotYetValid(t *testing.T) {
	Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { Now = time.Now })
	header := map[string]any{"alg": "HS256"}
	payload := map[string]any{"sub": "u1", "nbf": float64(2000)}
	token := makeToken(header, payload, "sig")
	tok, err := Decode(token)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if tok.Expiry.Status != "not_yet_valid" {
		t.Fatalf("status: got %q", tok.Expiry.Status)
	}
	if tok.Expiry.Remaining != 1000 {
		t.Fatalf("remaining: got %d", tok.Expiry.Remaining)
	}
}

func TestDecode_Valid(t *testing.T) {
	Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { Now = time.Now })
	header := map[string]any{"alg": "HS256"}
	payload := map[string]any{"exp": float64(2000)}
	token := makeToken(header, payload, "sig")
	tok, err := Decode(token)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if tok.Expiry.Status != "valid" {
		t.Fatalf("status: got %q", tok.Expiry.Status)
	}
	if tok.Expiry.Remaining != 1000 {
		t.Fatalf("remaining: got %d", tok.Expiry.Remaining)
	}
}

func TestDecode_NoExpiry(t *testing.T) {
	Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { Now = time.Now })
	header := map[string]any{"alg": "HS256"}
	payload := map[string]any{"sub": "u1"}
	token := makeToken(header, payload, "sig")
	tok, err := Decode(token)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if tok.Expiry.Status != "no_expiry" {
		t.Fatalf("status: got %q", tok.Expiry.Status)
	}
}

func TestDecode_IatInformational(t *testing.T) {
	Now = func() time.Time { return time.Unix(2000, 0).UTC() }
	t.Cleanup(func() { Now = time.Now })
	header := map[string]any{"alg": "HS256"}
	payload := map[string]any{"iat": float64(1000), "exp": float64(3000)}
	token := makeToken(header, payload, "sig")
	tok, err := Decode(token)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if tok.Expiry.Iat == nil || *tok.Expiry.Iat != 1000 {
		t.Fatalf("iat: got %v", tok.Expiry.Iat)
	}
	if tok.Expiry.Status != "valid" {
		t.Fatalf("status: got %q", tok.Expiry.Status)
	}
}

func TestDecode_NonNumericExpStillDecodes(t *testing.T) {
	Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { Now = time.Now })
	header := map[string]any{"alg": "HS256"}
	payload := map[string]any{"exp": "not-a-number", "sub": "u1"}
	token := makeToken(header, payload, "sig")
	tok, err := Decode(token)
	if err == nil {
		t.Fatalf("expected field error for non-numeric exp")
	}
	if !strings.Contains(err.Error(), "invalid exp") {
		t.Fatalf("err: got %v want invalid exp", err)
	}
	if tok == nil || tok.Payload["sub"] != "u1" {
		t.Fatalf("still expect decoded payload")
	}
}

func TestDecode_FloatExp(t *testing.T) {
	Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { Now = time.Now })
	header := map[string]any{"alg": "HS256"}
	payload := map[string]any{"exp": float64(1500.9)}
	token := makeToken(header, payload, "sig")
	tok, err := Decode(token)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if tok.Expiry.Exp == nil || *tok.Expiry.Exp != 1500 {
		t.Fatalf("exp float truncated: got %v", tok.Expiry.Exp)
	}
}

func TestDecode_BadSegments(t *testing.T) {
	_, err := Decode("only.two")
	if err == nil || !strings.Contains(err.Error(), "expected 3 segments") {
		t.Fatalf("2 segments: got %v", err)
	}
	_, err = Decode("")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty: got %v", err)
	}
}

func TestDecode_BadBase64(t *testing.T) {
	token := "!!!.eyJzdWIiOiIxIn0.sig"
	_, err := Decode(token)
	if err == nil || !strings.Contains(err.Error(), "invalid header: base64url") {
		t.Fatalf("bad base64 header: got %v", err)
	}
}

func TestDecode_NonJSONPayload(t *testing.T) {
	header := encodeSegment(map[string]any{"alg": "HS256"})
	badPayload := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
	token := header + "." + badPayload + ".sig"
	_, err := Decode(token)
	if err == nil || !strings.Contains(err.Error(), "invalid payload: not JSON") {
		t.Fatalf("non-json payload: got %v", err)
	}
}

func TestDecode_NonJSONHeader(t *testing.T) {
	badHeader := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
	payload := encodeSegment(map[string]any{"sub": "u1"})
	token := badHeader + "." + payload + ".sig"
	_, err := Decode(token)
	if err == nil || !strings.Contains(err.Error(), "invalid header: not JSON") {
		t.Fatalf("non-json header: got %v", err)
	}
}
