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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// tokenExpirySkew is how long before the recorded expiry a cached token is
// treated as stale. It absorbs clock drift and network latency so a token
// never reaches the server expired.
const tokenExpirySkew = 30 * time.Second

// TokenCacheKey derives a stable cache key for a workspace root and auth
// config. The workspace root scopes tokens to the workspace, and the
// canonicalized config hash scopes them to a specific auth setup so changing
// client_id, token_url, or scope invalidates the cached token.
func TokenCacheKey(root string, cfg map[string]string) string {
	keys := make([]string, 0, len(cfg))
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	h.Write([]byte(root + "\x00"))
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{'='})
		h.Write([]byte(cfg[k]))
		h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// CachedTokenSource decorates a TokenSource with a store-backed cache. On
// Token it returns the persisted token while it is still fresh (now plus the
// expiry skew is before the recorded expiry); otherwise it acquires a new
// token from the underlying source and persists it. The key scopes the cache
// entry: callers derive it from the workspace root and the auth config.
type CachedTokenSource struct {
	src   TokenSource
	store secrets.Store
	key   string
}

// NewCachedTokenSource returns a CachedTokenSource that caches tokens for key
// in store, delegating misses to src.
func NewCachedTokenSource(src TokenSource, store secrets.Store, key string) *CachedTokenSource {
	return &CachedTokenSource{src: src, store: store, key: key}
}

// CachedToken is the decoded on-disk representation of a cached OAuth token,
// exported so the CLI can report status without acquiring.
type CachedToken struct {
	AccessToken  string
	TokenType    string
	Endpoint     string
	Expiry       time.Time
	RefreshToken string
	GrantType    string
}

// ParseCachedToken decodes the raw store value written by CachedTokenSource.
func ParseCachedToken(raw string) (CachedToken, error) {
	var c cachedToken
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return CachedToken{}, err
	}
	return CachedToken{
		AccessToken:  c.AccessToken,
		TokenType:    c.TokenType,
		Endpoint:     c.Endpoint,
		Expiry:       c.Expiry,
		RefreshToken: c.RefreshToken,
		GrantType:    c.GrantType,
	}, nil
}

// cachedToken is the on-disk representation of a Token.
type cachedToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	// Endpoint is the token URL the token was acquired from, stored so the
	// CLI can report it without re-deriving the config.
	Endpoint string    `json:"endpoint,omitempty"`
	Expiry   time.Time `json:"expiry,omitempty"`
	// RefreshToken is persisted so an expired access token can be renewed
	// via the refresh-token grant without re-running the browser flow.
	RefreshToken string `json:"refresh_token,omitempty"`
	// GrantType is the grant used to acquire the token, stored so the CLI
	// can report it (client_credentials vs authorization_code).
	GrantType string `json:"grant_type,omitempty"`
}

// Token returns a fresh token for the cache key, reusing the persisted token
// while it remains valid (plus skew). An expired entry with a refresh token
// is renewed via the refresh-token grant when the underlying source supports
// it (no re-run of the original grant, e.g. no second browser flow);
// otherwise a new token is acquired from the underlying source.
func (c *CachedTokenSource) Token(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error) {
	if c.src == nil {
		return Token{}, fmt.Errorf("cached token source has no underlying source")
	}

	if raw, err := c.store.Get(c.key); err == nil {
		var cached cachedToken
		if err := json.Unmarshal([]byte(raw), &cached); err != nil {
			// Corrupt or foreign entry: drop it and re-acquire rather than
			// failing the request. A delete error is irrelevant here — the
			// entry is unusable either way.
			_ = c.store.Delete(c.key)
		} else if cached.AccessToken != "" && (cached.Expiry.IsZero() || time.Now().Add(tokenExpirySkew).Before(cached.Expiry)) {
			return Token{
				AccessToken:  cached.AccessToken,
				TokenType:    cached.TokenType,
				Expiry:       cached.Expiry,
				RefreshToken: cached.RefreshToken,
			}, nil
		} else if cached.RefreshToken != "" {
			// Expired but renewable: use the refresh-token grant instead of
			// re-running the original grant (which would reopen the browser
			// for authorization_code flows).
			if tok, renewed, err := c.renewFromRefreshToken(ctx, cfg, vars, cached.RefreshToken); err != nil {
				return Token{}, err
			} else if renewed {
				return tok, nil
			}
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return Token{}, fmt.Errorf("read cached token: %w", err)
	}

	tok, err := c.src.Token(ctx, cfg, vars)
	if err != nil {
		return Token{}, err
	}
	if _, err := c.persist(tok, cfg); err != nil {
		return Token{}, err
	}
	return tok, nil
}

// ForceRefresh invalidates the cached access token and renews it: via the
// refresh-token grant when the entry carries a refresh token (no re-run of
// the original grant), otherwise by re-acquiring from the underlying source.
// The request engine uses it on a reactive 401 so stale tokens recover
// without a retry loop and without reopening the browser.
func (c *CachedTokenSource) ForceRefresh(ctx context.Context, cfg map[string]string, vars Interpolator) (Token, error) {
	if raw, err := c.store.Get(c.key); err == nil {
		if cached, err := ParseCachedToken(raw); err == nil && cached.RefreshToken != "" {
			if tok, renewed, err := c.renewFromRefreshToken(ctx, cfg, vars, cached.RefreshToken); err != nil {
				return Token{}, err
			} else if renewed {
				return tok, nil
			}
		}
	}
	// No refresh token available: drop the entry and re-acquire from the
	// underlying source (the original grant).
	_ = c.store.Delete(c.key)
	tok, err := c.src.Token(ctx, cfg, vars)
	if err != nil {
		return Token{}, err
	}
	if _, err := c.persist(tok, cfg); err != nil {
		return Token{}, err
	}
	return tok, nil
}

// renewFromRefreshToken renews an expired token via the refresh-token grant
// (RFC 6749 §6) and persists the result, rotating the refresh token when the
// response carries a new one. It reports false when the underlying source
// cannot refresh, so callers fall back to re-running the original grant.
func (c *CachedTokenSource) renewFromRefreshToken(ctx context.Context, cfg map[string]string, vars Interpolator, refreshToken string) (Token, bool, error) {
	rs, ok := c.src.(RefreshingTokenSource)
	if !ok {
		return Token{}, false, nil
	}
	tok, err := rs.RefreshToken(ctx, cfg, vars, refreshToken)
	if err != nil {
		return Token{}, false, err
	}
	refresh, err := c.persist(tok, cfg)
	if err != nil {
		return Token{}, false, err
	}
	tok.RefreshToken = refresh
	return tok, true, nil
}

// persist encodes tok and stores it under the cache key, rotating the
// refresh token when the response carries a new one and otherwise keeping the
// previously cached refresh token (RFC 6749 §6 allows the refresh response to
// omit it). It returns the persisted refresh token so callers can surface it.
func (c *CachedTokenSource) persist(tok Token, cfg map[string]string) (string, error) {
	refresh := tok.RefreshToken
	if refresh == "" {
		if raw, err := c.store.Get(c.key); err == nil {
			var prev cachedToken
			if err := json.Unmarshal([]byte(raw), &prev); err == nil {
				refresh = prev.RefreshToken
			}
		}
	}
	blob, err := json.Marshal(cachedToken{
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		Endpoint:     cfg["token_url"],
		Expiry:       tok.Expiry,
		RefreshToken: refresh,
		GrantType:    cfg["grant_type"],
	})
	if err != nil {
		return "", fmt.Errorf("encode cached token: %w", err)
	}
	if err := c.store.Set(c.key, string(blob)); err != nil {
		return "", fmt.Errorf("persist cached token: %w", err)
	}
	return refresh, nil
}
