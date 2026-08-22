// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package history

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "history.db")
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStoreInsertAndList(t *testing.T) {
	s := newTempStore(t)
	ctx := context.Background()
	e := &Entry{
		RequestPath: "users/list",
		Method:      "GET",
		URL:         "https://api.example.com/users",
		Env:         "dev",
		Status:      200,
		DurationMS:  120,
		Size:        42,
		ReqHeaders:  map[string][]string{"Accept": {"application/json"}},
		ReqBody:     []byte(`{"q":1}`),
		RespHeaders: map[string][]string{"Content-Type": {"application/json"}},
		RespBody:    []byte(`{"ok":true}`),
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.Insert(ctx, e); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if e.ID == "" {
		t.Fatal("expected ID assigned")
	}
	list, err := s.List(ctx, 10, 0, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].URL != e.URL {
		t.Fatalf("List: got %v", list)
	}
}

func TestStoreShow(t *testing.T) {
	s := newTempStore(t)
	ctx := context.Background()
	e := &Entry{URL: "https://api.example.com/a", Method: "POST", Status: 201, CreatedAt: time.Now().UTC(), ReqBody: []byte("hello"), RespBody: []byte("world")}
	_ = s.Insert(ctx, e)
	got, err := s.Show(ctx, e.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if string(got.ReqBody) != "hello" || string(got.RespBody) != "world" {
		t.Fatalf("Show body mismatch: %q %q", got.ReqBody, got.RespBody)
	}
}

func TestStoreSearchFTS(t *testing.T) {
	s := newTempStore(t)
	ctx := context.Background()
	_ = s.Insert(ctx, &Entry{URL: "https://api.example.com/users/list", RequestPath: "users/list", CreatedAt: time.Now().UTC()})
	_ = s.Insert(ctx, &Entry{URL: "https://api.example.com/posts", RequestPath: "posts/list", CreatedAt: time.Now().UTC()})
	res, err := s.Search(ctx, "users", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(res) != 1 || res[0].RequestPath != "users/list" {
		t.Fatalf("Search: got %v", res)
	}
}

func TestStoreRetention(t *testing.T) {
	s := newTempStore(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_ = s.Insert(ctx, &Entry{URL: "https://api.example.com/a", CreatedAt: time.Now().UTC()})
		time.Sleep(2 * time.Millisecond)
	}
	if err := s.EnforceRetention(ctx, 2); err != nil {
		t.Fatalf("EnforceRetention: %v", err)
	}
	list, _ := s.List(ctx, 10, 0, nil)
	if len(list) != 2 {
		t.Fatalf("Retention: expected 2, got %d", len(list))
	}
}

func TestStoreSpillLargeBody(t *testing.T) {
	s := newTempStore(t)
	ctx := context.Background()
	large := make([]byte, 1<<20+100)
	for i := range large {
		large[i] = 'x'
	}
	e := &Entry{URL: "https://api.example.com/large", ReqBody: large, RespBody: large, CreatedAt: time.Now().UTC()}
	if err := s.Insert(ctx, e); err != nil {
		t.Fatalf("Insert large: %v", err)
	}
	blobsDir := filepath.Join(filepath.Dir(s.dbPath), "history", "blobs")
	if _, err := os.Stat(blobsDir); err != nil {
		t.Fatalf("blobs dir not created: %v", err)
	}
	got, _ := s.Show(ctx, e.ID)
	if len(got.ReqBody) != len(large) {
		t.Fatalf("spill body len: got %d want %d", len(got.ReqBody), len(large))
	}
}

func TestStoreClear(t *testing.T) {
	s := newTempStore(t)
	ctx := context.Background()
	_ = s.Insert(ctx, &Entry{URL: "https://a.com", Env: "dev", CreatedAt: time.Now().UTC()})
	_ = s.Insert(ctx, &Entry{URL: "https://a.com", Env: "prod", CreatedAt: time.Now().UTC()})
	if err := s.Clear(ctx, strPtr("dev")); err != nil {
		t.Fatalf("Clear env: %v", err)
	}
	list, _ := s.List(ctx, 10, 0, nil)
	if len(list) != 1 || list[0].Env != "prod" {
		t.Fatalf("Clear env: got %v", list)
	}
	if err := s.Clear(ctx, nil); err != nil {
		t.Fatalf("Clear all: %v", err)
	}
	list, _ = s.List(ctx, 10, 0, nil)
	if len(list) != 0 {
		t.Fatalf("Clear all: got %d", len(list))
	}
}

func TestCookieJarCRUD(t *testing.T) {
	s := newTempStore(t)
	ctx := context.Background()
	c := Cookie{Name: "sess", Value: "abc", Domain: "example.com", Path: "/", Env: "dev", ExpiresAt: time.Now().Add(time.Hour)}
	if err := s.InsertCookie(ctx, c); err != nil {
		t.Fatalf("InsertCookie: %v", err)
	}
	cookies, _ := s.ListCookies(ctx, "dev")
	if len(cookies) != 1 || cookies[0].Value != "abc" {
		t.Fatalf("ListCookies: %v", cookies)
	}
	if err := s.DeleteCookie(ctx, "sess", "example.com", "/", "dev"); err != nil {
		t.Fatalf("DeleteCookie: %v", err)
	}
	cookies, _ = s.ListCookies(ctx, "dev")
	if len(cookies) != 0 {
		t.Fatalf("DeleteCookie: got %d", len(cookies))
	}
	_ = s.InsertCookie(ctx, c)
	_ = s.InsertCookie(ctx, Cookie{Name: "x", Value: "1", Domain: "example.com", Path: "/", Env: "prod", ExpiresAt: time.Now().Add(time.Hour)})
	if err := s.ClearCookies(ctx, strPtr("dev")); err != nil {
		t.Fatalf("ClearCookies dev: %v", err)
	}
	cookies, _ = s.ListCookies(ctx, "dev")
	if len(cookies) != 0 {
		t.Fatalf("ClearCookies dev: %v", cookies)
	}
	cookies, _ = s.ListCookies(ctx, "prod")
	if len(cookies) != 1 {
		t.Fatalf("ClearCookies prod: %v", cookies)
	}
}

func strPtr(s string) *string { return &s }
