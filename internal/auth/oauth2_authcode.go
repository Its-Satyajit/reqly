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
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Authorization Code + PKCE (RFC 6749 §4.1, RFC 7636). The flow is split so
// the browser-driving caller (CLI login, auto-login) can launch the
// authorization page and the loopback callback can be tested without one:
// StartAuthorizationFlow validates config and starts the one-shot listener,
// WaitCode blocks until the redirect delivers a code (or an error), and
// ExchangeCode trades the code for a token at the token endpoint.

// PKCEVerifier returns a code_verifier for RFC 7636: 32 random bytes
// base64url-encoded (43 characters, within the required 43–128 range).
func PKCEVerifier() (string, error) {
	return randToken(32)
}

// PKCEChallenge returns the S256 code_challenge for verifier (RFC 7636 §4.2).
func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// randToken returns n cryptographically random bytes as a base64url string.
func randToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("oauth2: generate random: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// BuildAuthorizationURL assembles the provider authorization URL for an
// authorization_code grant (RFC 6749 §4.1.1) with PKCE (RFC 7636 §4.3) and
// the caller-supplied state.
func BuildAuthorizationURL(cfg map[string]string, vars Interpolator, redirectURI, verifier, state string) (string, error) {
	authURL, err := requiredConfig(cfg, vars, "authorization_url")
	if err != nil {
		return "", err
	}
	clientID, err := requiredConfig(cfg, vars, "client_id")
	if err != nil {
		return "", err
	}

	u, err := url.Parse(authURL)
	if err != nil {
		return "", fmt.Errorf("oauth2 invalid authorization_url %q: %w", authURL, err)
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("code_challenge", PKCEChallenge(verifier))
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	if scope, err := vars.Interpolate(cfg["scope"]); err != nil {
		return "", fmt.Errorf("oauth2 scope: %w", err)
	} else if scope != "" {
		q.Set("scope", scope)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// AuthorizationFlow is an in-progress authorization_code flow: the
// authorization URL the user must approve, plus the one-shot callback
// transport waiting for the provider redirect — the loopback HTTP listener
// (CLI default) or a registered custom scheme receiving deep links
// (desktop).
type AuthorizationFlow struct {
	// AuthorizationURL is the page the user approves in their browser.
	AuthorizationURL string
	// RedirectURI is the exact redirect_uri sent in the authorization
	// request and the token exchange.
	RedirectURI string

	state        string
	verifier     string
	server       *http.Server
	ln           net.Listener
	customScheme string
	result       chan callbackResult
	handled      sync.Once
	closed       sync.Once
}

type callbackResult struct {
	code string
	err  error
}

// StartAuthorizationFlow validates the authorization_code config, generates
// PKCE (S256) and a state value, and starts the one-shot loopback callback
// listener. The redirect URI defaults to an ephemeral 127.0.0.1 port; a
// config-provided redirect_uri must also be loopback (127.0.0.1/localhost).
func StartAuthorizationFlow(cfg map[string]string, vars Interpolator) (*AuthorizationFlow, error) {
	for _, key := range []string{"authorization_url", "token_url", "client_id", "client_secret"} {
		if _, err := requiredConfig(cfg, vars, key); err != nil {
			return nil, err
		}
	}

	verifier, err := PKCEVerifier()
	if err != nil {
		return nil, err
	}
	state, err := randToken(16)
	if err != nil {
		return nil, err
	}

	flow := &AuthorizationFlow{
		AuthorizationURL: "",
		state:            state,
		verifier:         verifier,
		result:           make(chan callbackResult, 1),
	}
	if err := flow.startCallback(cfg, vars); err != nil {
		return nil, err
	}

	authURL, err := BuildAuthorizationURL(cfg, vars, flow.RedirectURI, verifier, state)
	if err != nil {
		flow.Close()
		return nil, err
	}
	flow.AuthorizationURL = authURL
	return flow, nil
}

// startCallback wires the callback transport for the flow. Loopback
// redirect URIs (the default: an ephemeral 127.0.0.1 port) start the one-shot
// HTTP listener; a custom scheme (e.g. reqly://callback) is accepted only
// when a receiver for that scheme has been registered (the desktop app) and
// completes via DeliverCustomSchemeCallback.
func (f *AuthorizationFlow) startCallback(cfg map[string]string, vars Interpolator) error {
	raw := cfg["redirect_uri"]
	if raw == "" {
		return f.startLoopback("http://127.0.0.1:0/callback", "127.0.0.1:0")
	}
	interpolated, err := vars.Interpolate(raw)
	if err != nil {
		return fmt.Errorf("oauth2 redirect_uri: %w", err)
	}
	u, err := url.Parse(interpolated)
	if err != nil {
		return fmt.Errorf("oauth2 invalid redirect_uri %q: %w", interpolated, err)
	}
	if host := u.Hostname(); host == "127.0.0.1" || host == "localhost" {
		if u.Port() == "" {
			return fmt.Errorf("oauth2 redirect_uri %q must include a port", interpolated)
		}
		return f.startLoopback(interpolated, u.Host)
	}

	// Non-loopback: only a registered custom-scheme receiver can deliver it.
	customSchemeMu.Lock()
	registered := customSchemeReceivers[u.Scheme]
	customSchemeMu.Unlock()
	if !registered {
		return fmt.Errorf("oauth2 redirect_uri %q uses scheme %q with no registered receiver; loopback (127.0.0.1/localhost) callbacks work by default, custom schemes require the desktop app", interpolated, u.Scheme)
	}
	f.RedirectURI = interpolated
	f.customScheme = u.Scheme
	customSchemeFlows.Store(u.Scheme, f)
	return nil
}

// startLoopback starts the one-shot HTTP callback listener and resolves the
// redirect URI (substituting the actual port when the URI asked for port 0).
func (f *AuthorizationFlow) startLoopback(redirectURI, listenAddr string) error {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("oauth2: start callback listener on %s: %w", listenAddr, err)
	}
	if u, err := url.Parse(redirectURI); err == nil && u.Port() == "0" {
		port := ln.Addr().(*net.TCPAddr).Port
		u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(port))
		redirectURI = u.String()
	}
	f.RedirectURI = redirectURI
	f.ln = ln
	f.server = &http.Server{Handler: http.HandlerFunc(f.serveHTTP)}
	go func() {
		// The listener serves a single flow; ServeHTTP delivers the result
		// and the deferred Close in Token() shuts the server down.
		if err := f.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case f.result <- callbackResult{err: fmt.Errorf("oauth2: callback server: %w", err)}:
			default:
			}
		}
	}()
	return nil
}

// customSchemeReceivers holds schemes the host app has registered as
// receivable via deep links (the desktop app registers its scheme at
// startup). Without a registration, flows using a custom scheme fail fast.
var (
	customSchemeMu        sync.Mutex
	customSchemeReceivers = map[string]bool{}
)

// RegisterCustomSchemeReceiver marks scheme as receivable by this process
// and returns an unregister func. The CLI registers nothing, so a
// non-loopback redirect_uri fails fast with an actionable error; the desktop
// app registers its deep-link scheme and feeds callbacks via
// DeliverCustomSchemeCallback.
func RegisterCustomSchemeReceiver(scheme string) func() {
	customSchemeMu.Lock()
	customSchemeReceivers[scheme] = true
	customSchemeMu.Unlock()
	return func() {
		customSchemeMu.Lock()
		delete(customSchemeReceivers, scheme)
		customSchemeMu.Unlock()
	}
}

// customSchemeFlows maps scheme → the flow currently waiting for a deep-link
// callback on that scheme. Only one flow per scheme can wait at a time.
var customSchemeFlows sync.Map

// DeliverCustomSchemeCallback feeds a deep-link callback payload into the
// flow waiting on uri's scheme. It verifies state, extracts the code (or the
// provider error), delivers it to the waiting flow, and removes the flow
// (one-shot: a second delivery for the same scheme is rejected). The desktop
// app calls this when the OS hands it a registered URL.
func DeliverCustomSchemeCallback(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("oauth2: invalid custom-scheme callback %q: %w", uri, err)
	}
	raw, ok := customSchemeFlows.LoadAndDelete(u.Scheme)
	if !ok {
		return fmt.Errorf("oauth2: no authorization flow waiting for scheme %q", u.Scheme)
	}
	flow, ok := raw.(*AuthorizationFlow)
	if !ok {
		return fmt.Errorf("oauth2: waiting flow for scheme %q has the wrong type", u.Scheme)
	}
	code, err := flow.parseCallbackQuery(u.Query())
	if err != nil {
		flow.deliver(callbackResult{err: err})
		return err
	}
	flow.deliver(callbackResult{code: code})
	return nil
}

// WaitCode blocks until the provider redirect delivers an authorization code
// (or an error), or ctx is done. It returns the code for ExchangeCode. When
// ctx carries no deadline, a hard cap bounds the wait so an abandoned flow
// cannot block forever.
func (f *AuthorizationFlow) WaitCode(ctx context.Context) (string, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, maxCallbackWait)
		defer cancel()
	}
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("oauth2: waiting for authorization callback: %w", ctx.Err())
	case r := <-f.result:
		if r.err != nil {
			return "", r.err
		}
		return r.code, nil
	}
}

// Close shuts down the callback transport. Safe to call more than once.
func (f *AuthorizationFlow) Close() error {
	var err error
	f.closed.Do(func() {
		if f.server != nil {
			err = f.server.Close()
		}
		if f.customScheme != "" {
			customSchemeFlows.Delete(f.customScheme)
		}
	})
	return err
}

// serveHTTP is the one-shot callback handler: it verifies state, extracts the
// code (or surfaces the provider error), delivers the result, and renders a
// small page for the browser. Any request after the first is rejected.
func (f *AuthorizationFlow) serveHTTP(w http.ResponseWriter, r *http.Request) {
	first := false
	f.handled.Do(func() { first = true })
	if !first {
		http.Error(w, "authorization flow already completed", http.StatusBadRequest)
		return
	}

	code, err := f.parseCallbackQuery(r.URL.Query())
	if err != nil {
		f.deliver(callbackResult{err: err})
		http.Error(w, "authorization failed", http.StatusBadRequest)
		return
	}
	f.deliver(callbackResult{code: code})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<!doctype html><html><body><h3>Authorization complete</h3><p>You can close this tab and return to Reqly.</p></body></html>")
}

// parseCallbackQuery verifies the state and extracts the authorization code
// (or the provider error) from a callback query. It is shared by the
// loopback HTTP handler and the custom-scheme deep-link delivery.
func (f *AuthorizationFlow) parseCallbackQuery(q url.Values) (string, error) {
	if state := q.Get("state"); state != f.state {
		return "", fmt.Errorf("oauth2: authorization callback state mismatch")
	}
	if errStr := q.Get("error"); errStr != "" {
		err := fmt.Errorf("oauth2: authorization failed: %s", errStr)
		if desc := q.Get("error_description"); desc != "" {
			err = fmt.Errorf("%w: %s", err, desc)
		}
		return "", err
	}
	code := q.Get("code")
	if code == "" {
		return "", fmt.Errorf("oauth2: authorization callback missing code")
	}
	return code, nil
}

func (f *AuthorizationFlow) deliver(r callbackResult) {
	select {
	case f.result <- r:
	default:
	}
}

// ExchangeCode performs the token-endpoint exchange (RFC 6749 §4.1.3) with
// the PKCE code_verifier (RFC 7636 §4.5) and Basic client auth.
func ExchangeCode(ctx context.Context, cfg map[string]string, vars Interpolator, redirectURI, code, verifier string) (Token, error) {
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
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", clientID)
	form.Set("code_verifier", verifier)
	if scope, err := vars.Interpolate(cfg["scope"]); err != nil {
		return Token{}, fmt.Errorf("oauth2 scope: %w", err)
	} else if scope != "" {
		form.Set("scope", scope)
	}

	return postTokenForm(ctx, tokenURL, clientID, clientSecret, form, tokenName(cfg))
}

// AuthorizationCodeSource is a TokenSource implementing the Authorization
// Code + PKCE grant end to end: it starts the one-shot loopback flow, calls
// Open with the authorization URL (when set) so the caller can launch the
// system browser, waits for the callback, and exchanges the code. When Open
// is nil the flow runs but nothing drives the callback — tests and the CLI
// login command supply their own driver.
type AuthorizationCodeSource struct {
	// Open is called with the authorization URL so the caller can launch
	// the system browser. It runs before the flow waits for the callback.
	Open func(ctx context.Context, authorizationURL string) error
}

// Token runs the full authorization code flow and returns the exchanged
// token (access token, expiry, and refresh token).
func (s *AuthorizationCodeSource) Token(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error) {
	flow, err := StartAuthorizationFlow(cfg, vars)
	if err != nil {
		return Token{}, err
	}
	defer flow.Close()

	if s.Open != nil {
		if err := s.Open(ctx, flow.AuthorizationURL); err != nil {
			return Token{}, fmt.Errorf("oauth2: open authorization page: %w", err)
		}
	}

	code, err := flow.WaitCode(ctx)
	if err != nil {
		return Token{}, err
	}
	return ExchangeCode(ctx, cfg, vars, flow.RedirectURI, code, flow.verifier)
}

// maxCallbackWait bounds how long a callback may take so an abandoned flow
// cannot block forever even when the request context has no deadline.
const maxCallbackWait = 10 * time.Minute

// oauth2BrowserOpener launches the system browser when the engine acquires an
// authorization_code token automatically (first request with no cached
// token). The CLI installs a real launcher via SetOAuth2BrowserOpener; the
// default fails fast with a clear error so an unconfigured flow never hangs.
var oauth2BrowserOpener = func(_ context.Context, _ string) error {
	return errors.New("oauth2: authorization_code flow needs a browser opener; configure it (CLI) or run reqly auth login first")
}

// SetOAuth2BrowserOpener installs the browser launcher used for automatic
// authorization_code acquisition, returning the previous launcher so callers
// (and tests) can restore it.
func SetOAuth2BrowserOpener(open func(ctx context.Context, authorizationURL string) error) func(ctx context.Context, authorizationURL string) error {
	prev := oauth2BrowserOpener
	if open != nil {
		oauth2BrowserOpener = open
	}
	return prev
}
