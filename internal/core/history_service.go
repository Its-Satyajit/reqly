// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package core

import (
	"context"

	"github.com/Its-Satyajit/reqly/internal/history"
	"github.com/Its-Satyajit/reqly/internal/request"
)

// HistoryService exposes history recording and querying, shared by CLI and Desktop.
// It masks secrets on display (Authorization/Cookie) while storing exact bytes.
type HistoryService struct {
	store  *history.Store
	client *request.Client
}

// NewHistoryService returns a service backed by store. client may be nil (Replay disabled).
func NewHistoryService(store *history.Store, client *request.Client) *HistoryService {
	if client == nil {
		client = request.NewClient()
	}
	return &HistoryService{store: store, client: client}
}

// Record persists an entry.
func (s *HistoryService) Record(ctx context.Context, e *history.Entry) error {
	return s.store.Insert(ctx, e)
}

// List returns masked entries.
func (s *HistoryService) List(ctx context.Context, limit, offset int, statusFilter *int) ([]history.Entry, error) {
	entries, err := s.store.List(ctx, limit, offset, statusFilter)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		maskEntry(&entries[i])
	}
	return entries, nil
}

// Show returns one masked entry.
func (s *HistoryService) Show(ctx context.Context, id string) (history.Entry, error) {
	e, err := s.store.Show(ctx, id)
	if err != nil {
		return history.Entry{}, err
	}
	maskEntry(&e)
	return e, nil
}

// Search returns masked entries matching FTS.
func (s *HistoryService) Search(ctx context.Context, q string, limit int) ([]history.Entry, error) {
	entries, err := s.store.Search(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		maskEntry(&entries[i])
	}
	return entries, nil
}

// Clear deletes history.
func (s *HistoryService) Clear(ctx context.Context, env *string) error {
	return s.store.Clear(ctx, env)
}

func maskEntry(e *history.Entry) {
	for k, vals := range e.ReqHeaders {
		if k == "Authorization" || k == "Cookie" {
			for i := range vals {
				vals[i] = "[SECRET]"
			}
			e.ReqHeaders[k] = vals
		}
	}
	for k, vals := range e.RespHeaders {
		if k == "Set-Cookie" || k == "Cookie" {
			for i := range vals {
				vals[i] = "[SECRET]"
			}
			e.RespHeaders[k] = vals
		}
	}
}
