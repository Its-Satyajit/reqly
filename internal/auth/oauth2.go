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

// Token is an OAuth 2.0 token acquired from a token endpoint.
type Token struct {
	AccessToken string
	TokenType   string
	// ExpiresIn is the lifetime in seconds, 0 when the response omits it.
	ExpiresIn int64
	// Expiry is the absolute expiry time derived from ExpiresIn plus the
	// acquisition time. It is zero when ExpiresIn is unknown.
	Expiry time.Time
	// RefreshToken is the refresh token returned by the token endpoint,
	// empty for grants that do not issue one (e.g. Client Credentials).
	RefreshToken string
}

// TokenSource acquires an OAuth 2.0 token for a scheme. It is implemented
// optionally by schemes that need a token before the request can be applied;
// the request engine calls it ahead of Apply and injects the token.
type TokenSource interface {
	Token(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error)
}

// RefreshingTokenSource is optionally implemented by TokenSources that can
// renew an expired access token from a refresh token (RFC 6749 §6) without
// re-running the original grant. The token cache uses it so auth-code tokens
// renew without reopening the browser.
type RefreshingTokenSource interface {
	RefreshToken(ctx context.Context, cfg map[string]string, vars Interpolator, refreshToken string) (Token, error)
}

// oauth2Scheme acquires an OAuth 2.0 token and applies it as a Bearer
// header. Acquisition and application are split: Token() does the grant
// round trip (dispatched on grant_type), Apply() just sets the header from
// the resolved token that the engine injected under the "token" config key.
type oauth2Scheme struct{}

// Token dispatches on grant_type: client_credentials performs an RFC 6749
// §4.4 grant; authorization_code runs the RFC 6749 §4.1 + RFC 7636 flow
// (see oauth2_authcode.go). Any other grant type is rejected up front.
func (oauth2Scheme) Token(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error) {
	grantType, err := vars.Interpolate(cfg["grant_type"])
	if err != nil {
		return Token{}, fmt.Errorf("oauth2 grant_type: %w", err)
	}
	if grantType == "" {
		grantType = "client_credentials"
	}
	switch grantType {
	case "client_credentials":
		return tokenClientCredentials(ctx, cfg, vars)
	case "authorization_code":
		return (&AuthorizationCodeSource{Open: oauth2BrowserOpener}).Token(ctx, cfg, vars)
	default:
		return Token{}, fmt.Errorf("oauth2: unsupported grant_type %q", grantType)
	}
}

// tokenClientCredentials performs an RFC 6749 §4.4 Client Credentials grant:
// a form-encoded POST to token_url with HTTP Basic client auth.
func tokenClientCredentials(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error) {
	tokenURL, err := requiredConfig(cfg, vars, "token_url")
	if err != nil {
		return Token{}, err
	}
	clientID, err := requiredConfig(cfg, vars, "client_id")
	if err != nil {
		return Token{}, err
	}
	clientSecret, err := requiredConfig(cfg, vars, "client_secret")
	if err != nil {
		return Token{}, err
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
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

	return postTokenForm(ctx, tokenURL, clientID, clientSecret, form, tokenName(cfg))
}

// requiredConfig interpolates a required auth config value, returning a
// descriptive error when it is missing or empty.
func requiredConfig(cfg map[string]string, vars Interpolator, key string) (string, error) {
	value, err := vars.Interpolate(cfg[key])
	if err != nil {
		return "", fmt.Errorf("oauth2 %s: %w", key, err)
	}
	if value == "" {
		return "", fmt.Errorf("oauth2 auth requires %q", key)
	}
	return value, nil
}

// tokenName returns the JSON field holding the access token, defaulting to
// access_token.
func tokenName(cfg map[string]string) string {
	if name := cfg["token_name"]; name != "" {
		return name
	}
	return "access_token"
}

// postTokenForm posts form to tokenURL with HTTP Basic client auth and
// parses the JSON response into a Token.
func postTokenForm(ctx context.Context, tokenURL, clientID, clientSecret string, form url.Values, name string) (Token, error) {
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

	tok, err := parseTokenResponse(body, name)
	if err != nil {
		return Token{}, err
	}
	return tok, nil
}

// parseTokenResponse decodes an OAuth 2.0 token endpoint JSON response,
// honoring token_name (default access_token) and capturing refresh_token.
func parseTokenResponse(body []byte, tokenName string) (Token, error) {
	var payload struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Token{}, fmt.Errorf("oauth2: parse token response: %w", err)
	}

	if tokenName != "access_token" {
		// token_name selects an alternate field when the provider returns
		// the token under a non-standard key.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			return Token{}, fmt.Errorf("oauth2: parse token response: %w", err)
		}
		if rawVal, ok := raw[tokenName]; ok {
			var value string
			if err := json.Unmarshal(rawVal, &value); err != nil {
				return Token{}, fmt.Errorf("oauth2: token_name %q is not a string: %w", tokenName, err)
			}
			payload.AccessToken = value
		}
	}

	if payload.AccessToken == "" {
		return Token{}, fmt.Errorf("oauth2: token response missing %q", tokenName)
	}

	tok := Token{
		AccessToken:  payload.AccessToken,
		TokenType:    payload.TokenType,
		ExpiresIn:    payload.ExpiresIn,
		RefreshToken: payload.RefreshToken,
	}
	if payload.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}
	return tok, nil
}

// RefreshToken renews an expired access token via the refresh-token grant
// (RFC 6749 §6): a form-encoded POST with grant_type=refresh_token and the
// refresh token, with HTTP Basic client auth. The response may carry a new
// refresh token (rotation); the cache persists it when present.
func (oauth2Scheme) RefreshToken(ctx context.Context, cfg map[string]string, vars Interpolator, refreshToken string) (Token, error) {
	tokenURL, err := requiredConfig(cfg, vars, "token_url")
	if err != nil {
		return Token{}, err
	}
	clientID, err := requiredConfig(cfg, vars, "client_id")
	if err != nil {
		return Token{}, err
	}
	clientSecret, err := requiredConfig(cfg, vars, "client_secret")
	if err != nil {
		return Token{}, err
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	if scope, err := vars.Interpolate(cfg["scope"]); err != nil {
		return Token{}, fmt.Errorf("oauth2 scope: %w", err)
	} else if scope != "" {
		form.Set("scope", scope)
	}

	return postTokenForm(ctx, tokenURL, clientID, clientSecret, form, tokenName(cfg))
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
