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
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

// newCNonce returns a fresh client nonce for digest challenges. It is a var so
// tests can pin it.
var newCNonce = func() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("digest: generating cnonce: %v", err))
	}
	return hex.EncodeToString(b)
}

// ChallengedScheme is implemented by schemes that must respond to a server
// challenge (e.g. digest) before they can produce credentials. The request
// client checks for this capability on a 401 response.
type ChallengedScheme interface {
	// Challenge applies the scheme to req given the WWW-Authenticate header
	// value from a 401 response.
	Challenge(req *http.Request, challenge string, cfg map[string]string, vars Interpolator) error
}

// digestScheme authenticates per RFC 2617 (MD5) and RFC 7616 (SHA-256)
// digest access authentication.
type digestScheme struct{}

// Apply validates required config up front. The actual credentials cannot be
// computed until the server sends its nonce in a 401 challenge.
func (digestScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	username, err := vars.Interpolate(cfg["username"])
	if err != nil {
		return fmt.Errorf("digest username: %w", err)
	}
	if username == "" {
		return fmt.Errorf("digest auth requires a username")
	}
	password, err := vars.Interpolate(cfg["password"])
	if err != nil {
		return fmt.Errorf("digest password: %w", err)
	}
	if password == "" {
		return fmt.Errorf("digest auth requires a password")
	}
	return nil
}

// Challenge parses the WWW-Authenticate challenge, computes the digest
// response, and sets the Authorization header on req.
func (s digestScheme) Challenge(req *http.Request, challenge string, cfg map[string]string, vars Interpolator) error {
	params, err := parseDigestChallenge(challenge)
	if err != nil {
		return fmt.Errorf("digest challenge: %w", err)
	}

	username, err := vars.Interpolate(cfg["username"])
	if err != nil {
		return fmt.Errorf("digest username: %w", err)
	}
	password, err := vars.Interpolate(cfg["password"])
	if err != nil {
		return fmt.Errorf("digest password: %w", err)
	}

	realm := params["realm"]
	algorithm := strings.ToUpper(params["algorithm"])
	if algorithm == "" {
		algorithm = "MD5"
	}
	if algorithm != "MD5" && algorithm != "SHA-256" {
		return fmt.Errorf("digest auth: unsupported algorithm %q", params["algorithm"])
	}

	qop := ""
	if params["qop"] != "" {
		qop = "auth"
	}
	cnonce := newCNonce()
	nc := "00000001"

	bodyHash := ""
	if params["qop"] == "auth-int" {
		h, err := digestBodyHash(req, algorithm)
		if err != nil {
			return err
		}
		bodyHash = h
		qop = "auth-int"
	}

	uri := req.URL.RequestURI()
	response, err := computeDigestResponse(req.Method, uri, username, password, realm,
		algorithm, params["nonce"], cnonce, nc, qop, bodyHash)
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("Digest ")
	fmt.Fprintf(&b, "username=%q", username)
	if realm != "" {
		fmt.Fprintf(&b, ", realm=%q", realm)
	}
	if nonce := params["nonce"]; nonce != "" {
		fmt.Fprintf(&b, ", nonce=%q", nonce)
	}
	fmt.Fprintf(&b, ", uri=%q", uri)
	if qop != "" {
		fmt.Fprintf(&b, ", qop=%s", qop)
		fmt.Fprintf(&b, ", nc=%s", nc)
		fmt.Fprintf(&b, ", cnonce=%q", cnonce)
	}
	if algorithm != "MD5" {
		fmt.Fprintf(&b, ", algorithm=%s", algorithm)
	}
	fmt.Fprintf(&b, ", response=%q", response)
	if opaque := params["opaque"]; opaque != "" {
		fmt.Fprintf(&b, ", opaque=%q", opaque)
	}
	req.Header.Set("Authorization", b.String())
	return nil
}

// digestBodyHash returns H(entity-body) for qop=auth-int. The body is read via
// GetBody so the request body is not consumed.
func digestBodyHash(req *http.Request, algorithm string) (string, error) {
	body, err := req.GetBody()
	if err != nil {
		return "", fmt.Errorf("digest auth-int: reading body: %w", err)
	}
	defer body.Close()
	raw := make([]byte, 0, 256)
	buf := make([]byte, 512)
	for {
		n, err := body.Read(buf)
		raw = append(raw, buf[:n]...)
		if err != nil {
			break
		}
	}
	return hashHex(algorithm, string(raw)), nil
}

// computeDigestResponse implements the RFC 2617/7616 response calculation.
func computeDigestResponse(method, uri, username, password, realm, algorithm,
	nonce, cnonce, nc, qop, bodyHash string) (string, error) {
	HA1 := hashHex(algorithm, username+":"+realm+":"+password)
	HA2Input := method + ":" + uri
	if qop == "auth-int" {
		HA2Input += ":" + bodyHash
	}
	HA2 := hashHex(algorithm, HA2Input)

	if qop != "" {
		if nonce == "" || cnonce == "" || nc == "" {
			return "", fmt.Errorf("digest auth: challenge missing nonce/qop parameters")
		}
		return hashHex(algorithm, HA1+":"+nonce+":"+nc+":"+cnonce+":"+qop+":"+HA2), nil
	}
	if nonce == "" {
		return "", fmt.Errorf("digest auth: challenge missing nonce")
	}
	return hashHex(algorithm, HA1+":"+nonce+":"+HA2), nil
}

func md5Hex(s string) string { return hashHex("MD5", s) }

func hashHex(algorithm, s string) string {
	switch algorithm {
	case "SHA-256":
		sum := sha256.Sum256([]byte(s))
		return hex.EncodeToString(sum[:])
	default:
		sum := md5.Sum([]byte(s))
		return hex.EncodeToString(sum[:])
	}
}

// parseDigestChallenge parses a WWW-Authenticate value like
// `Digest realm="x", nonce="y", qop="auth", opaque="z"` into its parameters.
func parseDigestChallenge(challenge string) (map[string]string, error) {
	params := make(map[string]string)
	challenge = strings.TrimSpace(challenge)
	if !strings.HasPrefix(challenge, "Digest") {
		return nil, fmt.Errorf("not a Digest challenge")
	}
	rest := strings.TrimSpace(challenge[len("Digest"):])
	rest = strings.TrimPrefix(rest, "=")

	i := 0
	for i < len(rest) {
		for i < len(rest) && (rest[i] == ',' || rest[i] == ' ' || rest[i] == '\t') {
			i++
		}
		if i >= len(rest) {
			break
		}
		start := i
		for i < len(rest) && rest[i] != '=' && rest[i] != ',' {
			i++
		}
		key := strings.TrimSpace(rest[start:i])
		if i >= len(rest) || rest[i] == ',' {
			continue
		}
		i++ // skip '='
		value := ""
		if i < len(rest) && rest[i] == '"' {
			i++
			vs := i
			for i < len(rest) && rest[i] != '"' {
				i++
			}
			value = rest[vs:i]
			if i < len(rest) {
				i++
			}
		} else {
			vs := i
			for i < len(rest) && rest[i] != ',' {
				i++
			}
			value = strings.TrimSpace(rest[vs:i])
		}
		if key != "" {
			params[key] = value
		}
	}
	return params, nil
}

func init() {
	Register("digest", digestScheme{})
}
