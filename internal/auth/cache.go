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
	AccessToken string
	TokenType   string
	Endpoint    string
	Expiry      time.Time
}

// ParseCachedToken decodes the raw store value written by CachedTokenSource.
func ParseCachedToken(raw string) (CachedToken, error) {
	var c cachedToken
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return CachedToken{}, err
	}
	return CachedToken{
		AccessToken: c.AccessToken,
		TokenType:   c.TokenType,
		Endpoint:    c.Endpoint,
		Expiry:      c.Expiry,
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
}

// Token returns a fresh token for the cache key, reusing the persisted token
// while it remains valid (plus skew), and otherwise acquiring and persisting a
// new one.
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
				AccessToken: cached.AccessToken,
				TokenType:   cached.TokenType,
				Expiry:      cached.Expiry,
			}, nil
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return Token{}, fmt.Errorf("read cached token: %w", err)
	}

	tok, err := c.src.Token(ctx, cfg, vars)
	if err != nil {
		return Token{}, err
	}
	blob, err := json.Marshal(cachedToken{
		AccessToken: tok.AccessToken,
		TokenType:   tok.TokenType,
		Endpoint:    cfg["token_url"],
		Expiry:      tok.Expiry,
	})
	if err != nil {
		return Token{}, fmt.Errorf("encode cached token: %w", err)
	}
	if err := c.store.Set(c.key, string(blob)); err != nil {
		return Token{}, fmt.Errorf("persist cached token: %w", err)
	}
	return tok, nil
}
