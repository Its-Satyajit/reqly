// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/history"
	"github.com/Its-Satyajit/reqly/internal/request"
)

func TestHistoryServiceRecordAndReplay(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "history.db")
	store, _ := history.NewStore(dbPath)
	defer store.Close()
	svc := NewHistoryService(store, request.NewClient())
	ctx := context.Background()
	e := &history.Entry{URL: "https://api.example.com/a", Method: "GET", Status: 200, ReqBody: []byte("req"), RespBody: []byte("resp")}
	if err := svc.Record(ctx, e); err != nil {
		t.Fatalf("Record: %v", err)
	}
	list, _ := svc.List(ctx, 10, 0, nil)
	if len(list) != 1 {
		t.Fatalf("List: %d", len(list))
	}
	got, _ := svc.Show(ctx, e.ID)
	if string(got.ReqBody) != "req" {
		t.Fatalf("Show: %q", got.ReqBody)
	}
}

func TestHistoryServiceMasking(t *testing.T) {
	dir := t.TempDir()
	store, _ := history.NewStore(filepath.Join(dir, "history.db"))
	defer store.Close()
	svc := NewHistoryService(store, request.NewClient())
	// inject secret via env? For now test that Authorization header is masked on display
	ctx := context.Background()
	e := &history.Entry{
		URL:        "https://api.example.com/a",
		ReqHeaders: map[string][]string{"Authorization": {"Bearer secret123"}},
		RespHeaders: map[string][]string{"Set-Cookie": {"sess=abc"}},
	}
	_ = svc.Record(ctx, e)
	got, _ := svc.Show(ctx, e.ID)
	if got.ReqHeaders["Authorization"][0] == "Bearer secret123" {
		t.Fatalf("expected masked Authorization, got %v", got.ReqHeaders)
	}
}

func TestHistoryServiceSearchAndClear(t *testing.T) {
	dir := t.TempDir()
	store, _ := history.NewStore(filepath.Join(dir, "history.db"))
	defer store.Close()
	svc := NewHistoryService(store, request.NewClient())
	ctx := context.Background()
	_ = svc.Record(ctx, &history.Entry{URL: "https://api.example.com/users", RequestPath: "users/list"})
	_ = svc.Record(ctx, &history.Entry{URL: "https://api.example.com/posts", RequestPath: "posts/list"})
	res, _ := svc.Search(ctx, "users", 10)
	if len(res) != 1 {
		t.Fatalf("Search: %d", len(res))
	}
	_ = svc.Clear(ctx, nil)
	list, _ := svc.List(ctx, 10, 0, nil)
	if len(list) != 0 {
		t.Fatalf("Clear: %d", len(list))
	}
}
