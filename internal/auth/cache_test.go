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

package auth_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/secrets"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// acquireCounter is a TokenSource that counts acquisitions.
type acquireCounter struct {
	calls int
}

func (a *acquireCounter) Token(ctx context.Context, _ map[string]string, _ auth.Interpolator) (auth.Token, error) {
	a.calls++
	return auth.Token{
		AccessToken: "fresh-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Expiry:      time.Now().Add(1 * time.Hour),
	}, nil
}

func TestCachedTokenSourceReusesToken(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	src := &acquireCounter{}
	ts := auth.NewCachedTokenSource(src, store, auth.TokenCacheKey(dir, map[string]string{"a": "1"}))

	first, err := ts.Token(context.Background(), nil, variables.NewSet())
	if err != nil {
		t.Fatalf("first Token: %v", err)
	}
	second, err := ts.Token(context.Background(), nil, variables.NewSet())
	if err != nil {
		t.Fatalf("second Token: %v", err)
	}

	if first.AccessToken != second.AccessToken {
		t.Fatalf("token changed across calls: %q != %q", first.AccessToken, second.AccessToken)
	}
	if src.calls != 1 {
		t.Fatalf("underlying source called %d times, want 1", src.calls)
	}
}

func TestCachedTokenSourceKeyScoped(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	src := &acquireCounter{}

	tsA := auth.NewCachedTokenSource(src, store, auth.TokenCacheKey(dir, map[string]string{"client_id": "a"}))
	tsB := auth.NewCachedTokenSource(src, store, auth.TokenCacheKey(dir, map[string]string{"client_id": "b"}))

	if _, err := tsA.Token(context.Background(), nil, variables.NewSet()); err != nil {
		t.Fatalf("Token A: %v", err)
	}
	if _, err := tsB.Token(context.Background(), nil, variables.NewSet()); err != nil {
		t.Fatalf("Token B: %v", err)
	}
	if src.calls != 2 {
		t.Fatalf("underlying source called %d times, want 2 (distinct keys)", src.calls)
	}
}

func TestCachedTokenSourceExpiredRefetches(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	src := &expiringSource{ttl: -1 * time.Second} // already expired
	ts := auth.NewCachedTokenSource(src, store, auth.TokenCacheKey(dir, map[string]string{"a": "1"}))

	if _, err := ts.Token(context.Background(), nil, variables.NewSet()); err != nil {
		t.Fatalf("first Token: %v", err)
	}
	if _, err := ts.Token(context.Background(), nil, variables.NewSet()); err != nil {
		t.Fatalf("second Token: %v", err)
	}
	if src.calls != 2 {
		t.Fatalf("underlying source called %d times, want 2 (expired token refetched)", src.calls)
	}
}

// expiringSource returns a token with a caller-controlled lifetime.
type expiringSource struct {
	calls int
	ttl   time.Duration
}

func (a *expiringSource) Token(ctx context.Context, _ map[string]string, _ auth.Interpolator) (auth.Token, error) {
	a.calls++
	return auth.Token{
		AccessToken: "fresh-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(a.ttl),
	}, nil
}

func TestCachedTokenSourcePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "tokens.json")
	store1, err := secrets.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("NewFileStore 1: %v", err)
	}
	src := &acquireCounter{}
	ts1 := auth.NewCachedTokenSource(src, store1, auth.TokenCacheKey(dir, map[string]string{"a": "1"}))
	if _, err := ts1.Token(context.Background(), nil, variables.NewSet()); err != nil {
		t.Fatalf("Token 1: %v", err)
	}

	// A fresh process-equivalent source (new store, new underlying) reuses
	// the persisted token.
	store2, err := secrets.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("NewFileStore 2: %v", err)
	}
	src2 := &acquireCounter{}
	ts2 := auth.NewCachedTokenSource(src2, store2, auth.TokenCacheKey(dir, map[string]string{"a": "1"}))
	tok, err := ts2.Token(context.Background(), nil, variables.NewSet())
	if err != nil {
		t.Fatalf("Token 2: %v", err)
	}
	if tok.AccessToken != "fresh-token" {
		t.Fatalf("AccessToken = %q, want %q", tok.AccessToken, "fresh-token")
	}
	if src2.calls != 0 {
		t.Fatalf("fresh source called %d times, want 0 (persisted token reused)", src2.calls)
	}
}

// TestCachedTokenSourceNoExpiryBoundedDefault checks that a token whose
// provider omits expires_in is not trusted forever: the persisted entry gets a
// conservative lifetime, after which the underlying source re-acquires.
func TestCachedTokenSourceNoExpiryBoundedDefault(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	src := &noExpirySource{}
	key := auth.TokenCacheKey(dir, map[string]string{"a": "1"})
	ts := auth.NewCachedTokenSource(src, store, key)

	// Acquire: the returned token carries no expiry, but the persisted entry
	// must record a bounded lifetime.
	if _, err := ts.Token(context.Background(), nil, variables.NewSet()); err != nil {
		t.Fatalf("Token 1: %v", err)
	}
	raw, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get persisted: %v", err)
	}
	persisted, err := auth.ParseCachedToken(raw)
	if err != nil {
		t.Fatalf("ParseCachedToken: %v", err)
	}
	if persisted.Expiry.IsZero() {
		t.Fatal("persisted expiry is zero; a no-expiry token must get a bounded default lifetime")
	}
	if !persisted.IsFresh(time.Now()) {
		t.Fatal("freshly persisted token should be fresh")
	}

	// Reuse while within the default lifetime.
	if _, err := ts.Token(context.Background(), nil, variables.NewSet()); err != nil {
		t.Fatalf("Token 2: %v", err)
	}
	if src.calls != 1 {
		t.Fatalf("underlying source called %d times, want 1 (reused within TTL)", src.calls)
	}

	// Beyond the default lifetime the token is stale and re-acquired.
	if !persisted.Expiry.IsZero() && persisted.Expiry.After(time.Now().Add(11*time.Minute)) {
		t.Fatalf("default expiry %v too far out; want within the default TTL", persisted.Expiry)
	}
}

// noExpirySource always returns a token without an expiry.
type noExpirySource struct {
	calls int
}

func (a *noExpirySource) Token(ctx context.Context, _ map[string]string, _ auth.Interpolator) (auth.Token, error) {
	a.calls++
	return auth.Token{AccessToken: "no-expiry-token", TokenType: "Bearer"}, nil
}

// TestCachedTokenSourceStoredTokenRoundTrip exercises the persisted shape
// end to end: what Token() writes must be readable back by ParseCachedToken
// with the same fields, so a refactor of the store encoding can't silently
// break reuse or the auth status CLI.
func TestCachedTokenSourceStoredTokenRoundTrip(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "tokens.json")
	store, err := secrets.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	src := &expiringSource{ttl: 1 * time.Hour}
	key := auth.TokenCacheKey(dir, map[string]string{"a": "1"})
	if _, err := auth.NewCachedTokenSource(src, store, key).Token(context.Background(), nil, variables.NewSet()); err != nil {
		t.Fatalf("Token: %v", err)
	}

	raw, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get stored token: %v", err)
	}
	parsed, err := auth.ParseCachedToken(raw)
	if err != nil {
		t.Fatalf("ParseCachedToken: %v", err)
	}
	if parsed.AccessToken != "fresh-token" {
		t.Fatalf("AccessToken = %q, want %q", parsed.AccessToken, "fresh-token")
	}
	if parsed.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want %q", parsed.TokenType, "Bearer")
	}
	if parsed.Endpoint != "" {
		t.Fatalf("Endpoint = %q, want empty (no token_url in cfg)", parsed.Endpoint)
	}
	if parsed.Expiry.IsZero() {
		t.Fatal("Expiry is zero, want the derived expiry")
	}
}

// TestCachedTokenSourceCorruptEntryReacquires verifies a corrupt or foreign
// stored value does not brick the request: the entry is dropped and a fresh
// token acquired.
func TestCachedTokenSourceCorruptEntryReacquires(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	key := auth.TokenCacheKey(dir, map[string]string{"a": "1"})
	if err := store.Set(key, "{not-json"); err != nil {
		t.Fatalf("seed corrupt entry: %v", err)
	}

	src := &acquireCounter{}
	tok, err := auth.NewCachedTokenSource(src, store, key).Token(context.Background(), nil, variables.NewSet())
	if err != nil {
		t.Fatalf("Token with corrupt entry: %v", err)
	}
	if tok.AccessToken != "fresh-token" {
		t.Fatalf("AccessToken = %q, want %q", tok.AccessToken, "fresh-token")
	}
	if src.calls != 1 {
		t.Fatalf("underlying source called %d times, want 1 (re-acquired after corrupt entry)", src.calls)
	}

	// The corrupt entry must have been replaced by a valid one.
	raw, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get after re-acquire: %v", err)
	}
	if _, err := auth.ParseCachedToken(raw); err != nil {
		t.Fatalf("stored entry still unreadable: %v", err)
	}
}

// refreshingCounter is a TokenSource that also implements
// RefreshingTokenSource, so the cache can renew without re-running the
// original grant.
type refreshingCounter struct {
	tokenCalls   int
	refreshCalls int
	// newRefresh is the refresh token returned by RefreshToken; empty means
	// the refresh response omits one (RFC 6749 §6 allows this).
	newRefresh string
}

func (a *refreshingCounter) Token(context.Context, map[string]string, auth.Interpolator) (auth.Token, error) {
	a.tokenCalls++
	return auth.Token{AccessToken: "fresh-token", TokenType: "Bearer", ExpiresIn: 3600, Expiry: time.Now().Add(1 * time.Hour)}, nil
}

func (a *refreshingCounter) RefreshToken(_ context.Context, _ map[string]string, _ auth.Interpolator, _ string) (auth.Token, error) {
	a.refreshCalls++
	return auth.Token{AccessToken: "refreshed-token", TokenType: "Bearer", ExpiresIn: 3600, Expiry: time.Now().Add(1 * time.Hour), RefreshToken: a.newRefresh}, nil
}

func authCodeCfg() map[string]string {
	return map[string]string{"grant_type": "authorization_code", "token_url": "https://token.example.com"}
}

func seedCachedToken(t *testing.T, store *secrets.FileStore, key, access, refresh string, expiry time.Time) {
	t.Helper()
	blob, err := json.Marshal(map[string]any{
		"access_token":  access,
		"token_type":    "Bearer",
		"expiry":        expiry,
		"refresh_token": refresh,
		"grant_type":    "authorization_code",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set(key, string(blob)); err != nil {
		t.Fatal(err)
	}
}

func TestCachedTokenSourceRenewsExpiredViaRefreshToken(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	cfg := authCodeCfg()
	key := auth.TokenCacheKey(dir, cfg)
	seedCachedToken(t, store, key, "old", "rt-1", time.Now().Add(-1*time.Hour))

	src := &refreshingCounter{newRefresh: "rt-2"}
	tok, err := auth.NewCachedTokenSource(src, store, key).Token(context.Background(), cfg, variables.NewSet())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if tok.AccessToken != "refreshed-token" {
		t.Fatalf("AccessToken = %q, want refreshed-token", tok.AccessToken)
	}
	if src.tokenCalls != 0 {
		t.Fatalf("original grant re-run %d times, want 0 (refresh-token grant used)", src.tokenCalls)
	}
	if src.refreshCalls != 1 {
		t.Fatalf("RefreshToken called %d times, want 1", src.refreshCalls)
	}
	if tok.RefreshToken != "rt-2" {
		t.Fatalf("RefreshToken = %q, want rotated rt-2", tok.RefreshToken)
	}

	raw, err := store.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := auth.ParseCachedToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.AccessToken != "refreshed-token" || parsed.RefreshToken != "rt-2" {
		t.Fatalf("stored = %q/%q, want refreshed-token/rt-2", parsed.AccessToken, parsed.RefreshToken)
	}
}

func TestCachedTokenSourceRefreshKeepsOldRefreshToken(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	cfg := authCodeCfg()
	key := auth.TokenCacheKey(dir, cfg)
	seedCachedToken(t, store, key, "old", "rt-1", time.Now().Add(-1*time.Hour))

	src := &refreshingCounter{} // refresh response omits a new refresh token
	tok, err := auth.NewCachedTokenSource(src, store, key).Token(context.Background(), cfg, variables.NewSet())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if tok.RefreshToken != "rt-1" {
		t.Fatalf("RefreshToken = %q, want kept rt-1", tok.RefreshToken)
	}
}

func TestCachedTokenSourceForceRefreshViaRefreshToken(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	cfg := authCodeCfg()
	key := auth.TokenCacheKey(dir, cfg)
	seedCachedToken(t, store, key, "old", "rt-1", time.Now().Add(1*time.Hour))

	src := &refreshingCounter{newRefresh: "rt-3"}
	tok, err := auth.NewCachedTokenSource(src, store, key).ForceRefresh(context.Background(), cfg, variables.NewSet())
	if err != nil {
		t.Fatalf("ForceRefresh: %v", err)
	}
	if tok.AccessToken != "refreshed-token" {
		t.Fatalf("AccessToken = %q, want refreshed-token", tok.AccessToken)
	}
	if src.tokenCalls != 0 {
		t.Fatalf("original grant re-run %d times, want 0", src.tokenCalls)
	}
	if src.refreshCalls != 1 {
		t.Fatalf("RefreshToken called %d times, want 1", src.refreshCalls)
	}
	if tok.RefreshToken != "rt-3" {
		t.Fatalf("RefreshToken = %q, want rotated rt-3", tok.RefreshToken)
	}
}

func TestCachedTokenSourceForceRefreshReacquiresWithoutRefreshToken(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	cfg := map[string]string{"grant_type": "client_credentials", "token_url": "https://token.example.com"}
	key := auth.TokenCacheKey(dir, cfg)
	seedCachedToken(t, store, key, "old", "", time.Now().Add(1*time.Hour))

	src := &refreshingCounter{}
	tok, err := auth.NewCachedTokenSource(src, store, key).ForceRefresh(context.Background(), cfg, variables.NewSet())
	if err != nil {
		t.Fatalf("ForceRefresh: %v", err)
	}
	if tok.AccessToken != "fresh-token" {
		t.Fatalf("AccessToken = %q, want fresh-token (re-acquired)", tok.AccessToken)
	}
	if src.tokenCalls != 1 {
		t.Fatalf("original grant re-run %d times, want 1", src.tokenCalls)
	}
	if src.refreshCalls != 0 {
		t.Fatalf("RefreshToken called %d times, want 0 (no refresh token cached)", src.refreshCalls)
	}
}
