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

package core

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// AuthService exposes OAuth 2.0 token management (interactive login, status,
// logout) to front-ends, mirroring the `reqly auth` CLI commands on the same
// seams: the auth.TokenSource grants, the per-workspace secrets.Store cache,
// and masked output. The service is UI-agnostic: opening the system browser
// for authorization_code flows is the caller's job via LoginRequest.Open.
type AuthService struct {
	store secrets.Store
	root  string
}

// NewAuthService returns an AuthService caching tokens in store, scoped to
// the workspace root. store must be non-nil.
func NewAuthService(store secrets.Store, root string) *AuthService {
	return &AuthService{store: store, root: root}
}

// LoginRequest describes an interactive OAuth 2.0 login.
type LoginRequest struct {
	// Config is the flat auth config (ADR 0005): grant_type,
	// authorization_url/device_authorization_url, token_url, client_id,
	// client_secret, redirect_uri, scope, audience.
	Config map[string]string
	// Flow selects the grant: "auto" (infer from the config), or an explicit
	// "authorization_code" / "device_code".
	Flow string
	// Open is called with the authorization URL so the caller can launch the
	// system browser (authorization_code flows only). May be nil, in which
	// case nothing drives the callback (the desktop can feed it via
	// auth.DeliverCustomSchemeCallback).
	Open func(ctx context.Context, authorizationURL string) error
}

// Login performs the interactive grant and caches the token under the same
// key the request engine uses, so subsequent requests reuse it.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (auth.Token, error) {
	if s == nil || s.store == nil {
		return auth.Token{}, fmt.Errorf("no workspace found: open a reqly workspace to log in")
	}
	cfg := make(map[string]string, len(req.Config))
	for k, v := range req.Config {
		cfg[k] = v
	}
	flow := req.Flow
	if flow == "" || flow == "auto" {
		if cfg["device_authorization_url"] != "" && cfg["authorization_url"] == "" {
			flow = "device_code"
		} else {
			flow = "authorization_code"
		}
	}

	var src auth.TokenSource
	switch flow {
	case "authorization_code":
		cfg["grant_type"] = "authorization_code"
		src = &auth.AuthorizationCodeSource{Open: req.Open}
	case "device_code":
		cfg["grant_type"] = "device_code"
		src = &auth.DeviceCodeSource{}
	default:
		return auth.Token{}, fmt.Errorf("unknown login flow %q (want authorization_code or device_code)", flow)
	}

	cached := auth.NewCachedTokenSource(src, s.store, auth.TokenCacheKey(s.root, cfg))
	tok, err := cached.Token(ctx, cfg, variables.NewSet())
	if err != nil {
		return auth.Token{}, err
	}
	return tok, nil
}

// AuthTokenStatus is the masked, UI-friendly view of one cached token.
type AuthTokenStatus struct {
	Endpoint    string `json:"endpoint"`
	GrantType   string `json:"grantType"`
	Expiry      string `json:"expiry"` // RFC3339, or "" when unknown
	AccessToken string `json:"accessToken"`
	HasRefresh  bool   `json:"hasRefresh"`
	State       string `json:"state"` // "cached" | "expired"
}

// Status lists the cached tokens with masked values and the store backend.
func (s *AuthService) Status() ([]AuthTokenStatus, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to see auth status")
	}
	keys, err := s.store.Keys()
	if err != nil {
		return nil, fmt.Errorf("read token store: %w", err)
	}
	sort.Strings(keys)
	now := time.Now()
	out := make([]AuthTokenStatus, 0, len(keys))
	for _, key := range keys {
		raw, err := s.store.Get(key)
		if err != nil {
			return nil, fmt.Errorf("read token %s: %w", key, err)
		}
		tok, err := auth.ParseCachedToken(raw)
		if err != nil {
			return nil, fmt.Errorf("decode token %s: %w", key, err)
		}
		state := "cached"
		if !tok.IsFresh(now) {
			state = "expired"
		}
		endpoint := tok.Endpoint
		if endpoint == "" {
			endpoint = "(unknown)"
		}
		grant := tok.GrantType
		if grant == "" {
			grant = "-"
		}
		out = append(out, AuthTokenStatus{
			Endpoint:    endpoint,
			GrantType:   grant,
			Expiry:      formatAuthExpiry(tok.Expiry),
			AccessToken: maskAuthToken(tok.AccessToken),
			HasRefresh:  tok.RefreshToken != "",
			State:       state,
		})
	}
	return out, nil
}

// Logout clears every cached token for the workspace and returns how many
// were removed.
func (s *AuthService) Logout() (int, error) {
	if s == nil || s.store == nil {
		return 0, fmt.Errorf("no workspace found: nothing to clear")
	}
	keys, err := s.store.Keys()
	if err != nil {
		return 0, fmt.Errorf("read token store: %w", err)
	}
	for _, key := range keys {
		if err := s.store.Delete(key); err != nil {
			return 0, fmt.Errorf("clear token %s: %w", key, err)
		}
	}
	return len(keys), nil
}

// formatAuthExpiry renders a token expiry, or "" when unknown.
func formatAuthExpiry(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// maskAuthToken renders a token with only its first and last four characters
// visible, so status output never prints a full credential.
func maskAuthToken(token string) string {
	if token == "" {
		return "(empty)"
	}
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + "********" + token[len(token)-4:]
}
