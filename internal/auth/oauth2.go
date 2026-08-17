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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Token is an OAuth 2.0 access token acquired from a token endpoint.
type Token struct {
	AccessToken string
	TokenType   string
	// ExpiresIn is the lifetime in seconds, 0 when the response omits it.
	ExpiresIn int64
	// Expiry is the absolute expiry time derived from ExpiresIn plus the
	// acquisition time. It is zero when ExpiresIn is unknown.
	Expiry time.Time
}

// TokenSource acquires an OAuth 2.0 token for a scheme. It is implemented
// optionally by schemes that need a token before the request can be applied;
// the request engine calls it ahead of Apply and injects the token.
type TokenSource interface {
	Token(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error)
}

// oauth2Scheme acquires a Client Credentials token and applies it as a
// Bearer header. Acquisition and application are split: Token() does the
// token-endpoint round trip, Apply() just sets the header from the resolved
// token that the engine injected under the "token" config key.
type oauth2Scheme struct{}

// Token performs an RFC 6749 §4.4 Client Credentials grant: a form-encoded
// POST to token_url with HTTP Basic client auth.
func (oauth2Scheme) Token(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error) {
	tokenURL, err := vars.Interpolate(cfg["token_url"])
	if err != nil {
		return Token{}, fmt.Errorf("oauth2 token_url: %w", err)
	}
	if tokenURL == "" {
		return Token{}, fmt.Errorf("oauth2 auth requires a token_url")
	}
	clientID, err := vars.Interpolate(cfg["client_id"])
	if err != nil {
		return Token{}, fmt.Errorf("oauth2 client_id: %w", err)
	}
	if clientID == "" {
		return Token{}, fmt.Errorf("oauth2 auth requires a client_id")
	}
	clientSecret, err := vars.Interpolate(cfg["client_secret"])
	if err != nil {
		return Token{}, fmt.Errorf("oauth2 client_secret: %w", err)
	}
	if clientSecret == "" {
		return Token{}, fmt.Errorf("oauth2 auth requires a client_secret")
	}

	form := url.Values{}
	grantType, err := vars.Interpolate(cfg["grant_type"])
	if err != nil {
		return Token{}, fmt.Errorf("oauth2 grant_type: %w", err)
	}
	if grantType == "" {
		grantType = "client_credentials"
	}
	form.Set("grant_type", grantType)
	if scope, err := vars.Interpolate(cfg["scope"]); err != nil {
		return Token{}, fmt.Errorf("oauth2 scope: %w", err)
	} else if scope != "" {
		form.Set("scope", scope)
	}
	if audience, err := vars.Interpolate(cfg["audience"]); err != nil {
		return Token{}, fmt.Errorf("oauth2 audience: %w", err)
	} else if audience != "" {
		form.Set("audience", audience)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, fmt.Errorf("oauth2: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("oauth2: token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Token{}, fmt.Errorf("oauth2: read token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("oauth2: token endpoint returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Token{}, fmt.Errorf("oauth2: parse token response: %w", err)
	}

	tokenName := cfg["token_name"]
	if tokenName == "" {
		tokenName = "access_token"
	}
	// token_name selects an alternate field when the provider returns the
	// token under a non-standard key.
	if tokenName != "access_token" {
		var alt struct {
			Value string `json:"-"`
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			return Token{}, fmt.Errorf("oauth2: parse token response: %w", err)
		}
		if rawVal, ok := raw[tokenName]; ok {
			if err := json.Unmarshal(rawVal, &alt.Value); err != nil {
				return Token{}, fmt.Errorf("oauth2: token_name %q is not a string: %w", tokenName, err)
			}
			payload.AccessToken = alt.Value
		}
	}

	if payload.AccessToken == "" {
		return Token{}, fmt.Errorf("oauth2: token response missing %q", tokenName)
	}

	tok := Token{
		AccessToken: payload.AccessToken,
		TokenType:   payload.TokenType,
		ExpiresIn:   payload.ExpiresIn,
	}
	if payload.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}
	return tok, nil
}

// Apply sets Authorization: Bearer <token>. The token is expected to have
// been injected into config under "token" by the request engine via the
// TokenSource pre-flight; it is not re-acquired here.
func (oauth2Scheme) Apply(req *http.Request, cfg map[string]string, _ Interpolator) error {
	token := cfg["token"]
	if token == "" {
		return fmt.Errorf("oauth2 auth requires a token; was a token acquired before apply?")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// SecretKeys reports the client secret as the sensitive static config value.
func (oauth2Scheme) SecretKeys() []string { return []string{"client_secret"} }

func init() {
	Register("oauth2", oauth2Scheme{})
}
