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

package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"time"
)

// jwtScheme signs a JSON Web Token per request and sends it as a Bearer
// token. HS256, HS384, and HS512 are supported, hand-rolled over crypto/hmac.
type jwtScheme struct{}

// Apply signs a JWT from secret/algorithm/claims/expiresIn and sets
// Authorization: Bearer <token>.
func (jwtScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	secret, err := vars.Interpolate(cfg["secret"])
	if err != nil {
		return fmt.Errorf("jwt secret: %w", err)
	}
	if secret == "" {
		return fmt.Errorf("jwt auth requires a secret")
	}

	algorithm, err := vars.Interpolate(cfg["algorithm"])
	if err != nil {
		return fmt.Errorf("jwt algorithm: %w", err)
	}
	if algorithm == "" {
		algorithm = "HS256"
	}
	newHash, ok := jwtHashes[algorithm]
	if !ok {
		return fmt.Errorf("jwt auth: unsupported algorithm %q", algorithm)
	}

	claims, err := vars.Interpolate(cfg["claims"])
	if err != nil {
		return fmt.Errorf("jwt claims: %w", err)
	}

	expiresIn, err := vars.Interpolate(cfg["expiresIn"])
	if err != nil {
		return fmt.Errorf("jwt expiresIn: %w", err)
	}

	token, err := signJWT(newHash, []byte(secret), algorithm, claims, expiresIn)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// jwtHashes maps JWT algorithm names to their hash constructors.
var jwtHashes = map[string]func() hash.Hash{
	"HS256": sha256.New,
	"HS384": sha512.New384,
	"HS512": sha512.New,
}

// SecretKeys reports the signing secret as the sensitive config value.
func (jwtScheme) SecretKeys() []string { return []string{"secret"} }

// signJWT builds a compact JWS: base64url(header).base64url(payload).signature.
func signJWT(newHash func() hash.Hash, secret []byte, algorithm, claims, expiresIn string) (string, error) {
	header, err := json.Marshal(map[string]string{
		"alg": algorithm,
		"typ": "JWT",
	})
	if err != nil {
		return "", fmt.Errorf("jwt: encoding header: %w", err)
	}

	payload := claims
	if claims == "" {
		payload = "{}"
	}
	if expiresIn != "" {
		seconds, err := strconv.ParseInt(expiresIn, 10, 64)
		if err != nil {
			return "", fmt.Errorf("jwt: expiresIn must be a number of seconds: %w", err)
		}
		var claimsObj map[string]any
		if err := json.Unmarshal([]byte(payload), &claimsObj); err != nil {
			return "", fmt.Errorf("jwt: claims must be a JSON object: %w", err)
		}
		claimsObj["exp"] = time.Now().Unix() + seconds
		encoded, err := json.Marshal(claimsObj)
		if err != nil {
			return "", fmt.Errorf("jwt: encoding claims: %w", err)
		}
		payload = string(encoded)
	} else if payload != "{}" {
		// Validate that standalone claims parse as JSON so broken configs
		// surface before the request is sent.
		var m map[string]any
		if err := json.Unmarshal([]byte(payload), &m); err != nil {
			return "", fmt.Errorf("jwt: claims must be a JSON object: %w", err)
		}
	}

	seg1 := base64.RawURLEncoding.EncodeToString(header)
	seg2 := base64.RawURLEncoding.EncodeToString([]byte(payload))

	mac := hmac.New(newHash, secret)
	mac.Write([]byte(seg1 + "." + seg2))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return seg1 + "." + seg2 + "." + sig, nil
}

func init() {
	Register("jwt", jwtScheme{})
}
