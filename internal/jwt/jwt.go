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
	"fmt"
	"strings"
	"time"
)

// Token is the decoded JWT.
type Token struct {
	Header     map[string]any `json:"header"`
	Payload    map[string]any `json:"payload"`
	RawHeader  string         `json:"-"`
	RawPayload string         `json:"-"`
	Signature  string         `json:"signature"`
	Alg        string         `json:"alg"`
	Expiry     ExpiryStatus   `json:"expiry"`
}

// ExpiryStatus reports exp/nbf/iat-derived validity at decode time.
type ExpiryStatus struct {
	Status    string `json:"status"` // expired|not_yet_valid|valid|no_expiry
	Remaining int64  `json:"remaining"`
	Exp       *int64 `json:"exp,omitempty"`
	Nbf       *int64 `json:"nbf,omitempty"`
	Iat       *int64 `json:"iat,omitempty"`
}

// Now is injectable for deterministic tests (default time.Now).
var Now = time.Now

// Decode decodes a JWT without verification. It strips an optional
// "Bearer " prefix, trims whitespace, and expects exactly 3 dot-segments
// (signature may be empty for alg:none). Per-segment errors are explicit.
func Decode(token string) (*Token, error) {
	raw := strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		raw = strings.TrimSpace(raw[7:])
	}
	if raw == "" {
		return nil, fmt.Errorf("invalid token: empty")
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token: expected 3 segments, got %d", len(parts))
	}
	segHeader, segPayload, sig := parts[0], parts[1], parts[2]

	headerBytes, err := base64.RawURLEncoding.DecodeString(segHeader)
	if err != nil {
		return nil, fmt.Errorf("invalid header: base64url: %w", err)
	}
	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid header: not JSON: %w", err)
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(segPayload)
	if err != nil {
		return nil, fmt.Errorf("invalid payload: base64url: %w", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("invalid payload: not JSON: %w", err)
	}
	alg := ""
	if v, ok := header["alg"]; ok {
		if s, ok := v.(string); ok {
			alg = s
		}
	}
	expiry, fieldErr := computeExpiry(payload)
	tok := &Token{
		Header:     header,
		Payload:    payload,
		RawHeader:  segHeader,
		RawPayload: segPayload,
		Signature:  sig,
		Alg:        alg,
		Expiry:     expiry,
	}
	if fieldErr != nil {
		// Return decoded token plus field error — caller can still inspect
		// header/payload while surfacing the expiry problem.
		return tok, fieldErr
	}
	return tok, nil
}

func computeExpiry(payload map[string]any) (ExpiryStatus, error) {
	var st ExpiryStatus
	var fieldErr error
	now := Now().UTC().Unix()

	parseNumeric := func(key string) (*int64, error) {
		v, ok := payload[key]
		if !ok {
			return nil, nil
		}
		switch n := v.(type) {
		case float64:
			i := int64(n)
			return &i, nil
		case int:
			i := int64(n)
			return &i, nil
		case int64:
			return &n, nil
		case json.Number:
			f, err := n.Float64()
			if err != nil {
				return nil, fmt.Errorf("invalid %s: not numeric: %w", key, err)
			}
			i := int64(f)
			return &i, nil
		default:
			return nil, fmt.Errorf("invalid %s: not numeric (got %T)", key, v)
		}
	}

	exp, err := parseNumeric("exp")
	if err != nil {
		fieldErr = err
	} else {
		st.Exp = exp
	}
	nbf, err := parseNumeric("nbf")
	if err != nil {
		if fieldErr == nil {
			fieldErr = err
		}
	} else {
		st.Nbf = nbf
	}
	iat, err := parseNumeric("iat")
	if err != nil {
		if fieldErr == nil {
			fieldErr = err
		}
	} else {
		st.Iat = iat
	}

	// Determine status.
	if fieldErr != nil {
		// Keep status computable from valid exp/nbf where possible, but
		// prefer to surface the field error. Use available exp/nbf for status.
		// If exp is invalid, fall through to no_expiry vs error handling
		// — we still compute based on nbf if present.
	}
	if st.Exp == nil && st.Nbf == nil {
		st.Status = "no_expiry"
		st.Remaining = 0
		return st, fieldErr
	}
	if st.Nbf != nil && now < *st.Nbf {
		st.Status = "not_yet_valid"
		// remaining until nbf
		st.Remaining = *st.Nbf - now
		return st, fieldErr
	}
	if st.Exp != nil {
		remaining := *st.Exp - now
		st.Remaining = remaining
		if remaining < 0 {
			st.Status = "expired"
		} else {
			st.Status = "valid"
		}
		return st, fieldErr
	}
	// nbf present and not in future, but no exp
	st.Status = "valid"
	st.Remaining = 0
	return st, fieldErr
}
