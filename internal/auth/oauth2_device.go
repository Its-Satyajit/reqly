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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Device flow (RFC 8628) — the OAuth 2.0 Device Authorization Grant. The
// client starts the flow at the device-authorization endpoint, shows the
// user a verification URI and code, then polls the token endpoint until the
// user approves (or the flow terminates). There is no browser callback, so
// the flow works headless and is fully testable with httptest.

// DeviceAuthorization is the response of the device-authorization endpoint
// (RFC 8628 §3.2).
type DeviceAuthorization struct {
	// DeviceCode is the opaque code the client sends while polling.
	DeviceCode string
	// UserCode is the code the user enters on the verification page.
	UserCode string
	// VerificationURI is the page the user opens to approve.
	VerificationURI string
	// VerificationURIComplete is an optional URI that already embeds the
	// user code, so the user only has to confirm.
	VerificationURIComplete string
	// Interval is the minimum seconds between polls (RFC 8628 §3.2); 0 when
	// the provider omits it (callers default to 5).
	Interval int
}

// StartDeviceAuthorization performs the device-authorization request
// (RFC 8628 §3.1): a form-encoded POST with HTTP Basic client auth.
func StartDeviceAuthorization(ctx context.Context, cfg map[string]string, vars Interpolator) (DeviceAuthorization, error) {
	daURL, err := requiredConfig(cfg, vars, "device_authorization_url")
	if err != nil {
		return DeviceAuthorization{}, err
	}
	clientID, err := requiredConfig(cfg, vars, "client_id")
	if err != nil {
		return DeviceAuthorization{}, err
	}
	clientSecret, err := requiredConfig(cfg, vars, "client_secret")
	if err != nil {
		return DeviceAuthorization{}, err
	}

	form := url.Values{}
	form.Set("client_id", clientID)
	if scope, err := vars.Interpolate(cfg["scope"]); err != nil {
		return DeviceAuthorization{}, fmt.Errorf("oauth2 scope: %w", err)
	} else if scope != "" {
		form.Set("scope", scope)
	}
	if audience, err := vars.Interpolate(cfg["audience"]); err != nil {
		return DeviceAuthorization{}, fmt.Errorf("oauth2 audience: %w", err)
	} else if audience != "" {
		form.Set("audience", audience)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, daURL, strings.NewReader(form.Encode()))
	if err != nil {
		return DeviceAuthorization{}, fmt.Errorf("oauth2: build device authorization request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return DeviceAuthorization{}, fmt.Errorf("oauth2: device authorization request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DeviceAuthorization{}, fmt.Errorf("oauth2: read device authorization response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return DeviceAuthorization{}, fmt.Errorf("oauth2: device authorization endpoint returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		Interval                int    `json:"interval"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return DeviceAuthorization{}, fmt.Errorf("oauth2: parse device authorization response: %w", err)
	}
	if payload.DeviceCode == "" || payload.UserCode == "" || payload.VerificationURI == "" {
		return DeviceAuthorization{}, fmt.Errorf("oauth2: device authorization response missing device_code/user_code/verification_uri")
	}
	return DeviceAuthorization{
		DeviceCode:              payload.DeviceCode,
		UserCode:                payload.UserCode,
		VerificationURI:         payload.VerificationURI,
		VerificationURIComplete: payload.VerificationURIComplete,
		Interval:                payload.Interval,
	}, nil
}

// DeviceCodeSource is a TokenSource implementing the Device Authorization
// grant (RFC 8628) end to end: it starts the flow, reports the verification
// URI and user code via Status, and polls until the user approves.
type DeviceCodeSource struct {
	// Status is called with a human-readable progress line while the flow
	// waits for approval (verification instructions, polling state). It may
	// be nil.
	Status func(status string)
}

// Token runs the device flow and returns the granted token.
func (s *DeviceCodeSource) Token(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error) {
	da, err := StartDeviceAuthorization(ctx, cfg, vars)
	if err != nil {
		return Token{}, err
	}
	uri := da.VerificationURI
	if da.VerificationURIComplete != "" {
		uri = da.VerificationURIComplete
	}
	s.status(fmt.Sprintf("open %s and enter code %s", uri, da.UserCode))
	return PollDeviceToken(ctx, cfg, vars, da.DeviceCode, da.Interval, s.status)
}

func (s *DeviceCodeSource) status(line string) {
	if s != nil && s.Status != nil {
		s.Status(line)
	}
}

// devicePollSlowDownExtra is added to the poll interval after a slow_down
// response (RFC 8628 §3.5). A variable so tests can shrink it.
var devicePollSlowDownExtra = 5 * time.Second

// PollDeviceToken polls the token endpoint with the device grant
// (RFC 8628 §3.4) until a token is granted, the flow terminates with a
// terminal error, or ctx is done. It waits the provider's interval before
// each poll and adds devicePollSlowDownExtra on a slow_down response
// (RFC 8628 §3.5).
func PollDeviceToken(ctx context.Context, cfg map[string]string, vars Interpolator, deviceCode string, intervalSeconds int, status func(string)) (Token, error) {
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

	interval := time.Duration(intervalSeconds) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	wait := time.NewTimer(interval)
	defer wait.Stop()
	for {
		select {
		case <-ctx.Done():
			return Token{}, fmt.Errorf("oauth2: device authorization: %w", ctx.Err())
		case <-wait.C:
		}

		tok, err := pollDeviceTokenOnce(ctx, tokenURL, clientID, clientSecret, deviceCode, tokenName(cfg))
		if err == nil {
			return tok, nil
		}
		var pollErr *devicePollError
		if !errors.As(err, &pollErr) {
			return Token{}, err
		}
		if pollErr.slowDown {
			interval += devicePollSlowDownExtra
		}
		if status != nil {
			status("waiting for authorization…")
		}
		wait.Reset(interval)
	}
}

// devicePollError marks a retryable poll response (RFC 8628 §3.5):
// authorization_pending (keep polling) and slow_down (poll at +5s).
type devicePollError struct {
	slowDown bool
	reason   string
}

func (e *devicePollError) Error() string { return e.reason }

// pollDeviceTokenOnce performs a single token-endpoint poll for the device
// grant. It returns a *devicePollError for retryable responses and a plain
// error for terminal or transport failures.
func pollDeviceTokenOnce(ctx context.Context, tokenURL, clientID, clientSecret, deviceCode, name string) (Token, error) {
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	form.Set("device_code", deviceCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, fmt.Errorf("oauth2: build device token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("oauth2: device token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Token{}, fmt.Errorf("oauth2: read device token response: %w", err)
	}

	// RFC 8628 §3.5: authorization_pending/slow_down (and terminal errors)
	// arrive as non-200 responses with a JSON error field.
	var errPayload struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(body, &errPayload)
	switch errPayload.Error {
	case "authorization_pending":
		return Token{}, &devicePollError{reason: "authorization pending"}
	case "slow_down":
		return Token{}, &devicePollError{slowDown: true, reason: "slow down"}
	case "access_denied":
		return Token{}, fmt.Errorf("oauth2: device authorization denied by the user")
	case "expired_token":
		return Token{}, fmt.Errorf("oauth2: device code expired; run reqly auth login again to restart the flow")
	}
	if errPayload.Error != "" {
		return Token{}, fmt.Errorf("oauth2: device authorization failed: %s", errPayload.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("oauth2: device token endpoint returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return parseTokenResponse(body, name)
}

// oauth2DeviceStatus reports progress from an automatically-acquired device
// flow (a first request with no cached token). The CLI installs a stderr
// printer so the verification URI is visible; the default is silent so
// library callers (and tests) never get unexpected output.
var oauth2DeviceStatus = func(string) {}

// SetOAuth2DeviceStatus installs the status reporter used for automatic
// device-flow acquisition, returning the previous reporter so callers (and
// tests) can restore it.
func SetOAuth2DeviceStatus(status func(line string)) func(line string) {
	prev := oauth2DeviceStatus
	if status != nil {
		oauth2DeviceStatus = status
	}
	return prev
}
