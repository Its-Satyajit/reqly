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
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/jwt"
)

// JwtClaim is one decoded claim rendered as a display string.
type JwtClaim struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// JwtTokenView is the inspector payload: claims flattened to display strings
// (nested values are JSON-encoded) so the frontend never handles raw any.
type JwtTokenView struct {
	Header    []JwtClaim       `json:"header"`
	Payload   []JwtClaim       `json:"payload"`
	Signature string           `json:"signature"`
	Alg       string           `json:"alg"`
	Expiry    jwt.ExpiryStatus `json:"expiry"`
}

// scalarizeClaims renders each claim as a display string, sorted by key so
// the inspector lists claims deterministically.
func scalarizeClaims(claims map[string]any) []JwtClaim {
	keys := make([]string, 0, len(claims))
	for k := range claims {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]JwtClaim, 0, len(keys))
	for _, k := range keys {
		v := claims[k]
		var rendered string
		switch t := v.(type) {
		case string:
			rendered = t
		case float64:
			rendered = strconv.FormatFloat(t, 'f', -1, 64)
		case bool:
			rendered = strconv.FormatBool(t)
		case nil:
			rendered = "null"
		default:
			if data, err := json.Marshal(t); err == nil {
				rendered = string(data)
			} else {
				rendered = fmt.Sprintf("%v", t)
			}
		}
		out = append(out, JwtClaim{Key: k, Value: rendered})
	}
	return out
}

// JwtDecode decodes a JWT offline (no verification) — identical fidelity to
// `reqly jwt decode`, including expiry status computed at decode time.
func (s *AppService) JwtDecode(token string) (*JwtTokenView, error) {
	tok, err := jwt.Decode(token)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &JwtTokenView{
		Header:    scalarizeClaims(tok.Header),
		Payload:   scalarizeClaims(tok.Payload),
		Signature: tok.Signature,
		Alg:       tok.Alg,
		Expiry:    tok.Expiry,
	}, nil
}

// JwtFromAuthHeader extracts a bearer token from an Authorization header
// value ("Bearer eyJ…") — used by the inspector's auto-capture.
func (s *AppService) JwtFromAuthHeader(value string) (string, bool) {
	v := strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(v), "bearer ") {
		token := strings.TrimSpace(v[7:])
		return token, token != ""
	}
	return "", false
}
