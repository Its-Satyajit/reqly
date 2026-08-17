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
