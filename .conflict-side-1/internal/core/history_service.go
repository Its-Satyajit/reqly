// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

// Record persists an entry and ingests Set-Cookie into the jar.
func (s *HistoryService) Record(ctx context.Context, e *history.Entry) error {
	if err := s.insert(ctx, e); err != nil {
		return err
	}
	s.IngestSetCookies(ctx, e.RespHeaders, e.Env)
	return nil
}

// insert persists an entry without touching the jar.
func (s *HistoryService) insert(ctx context.Context, e *history.Entry) error {
	return s.store.Insert(ctx, e)
}

// IngestSetCookies parses Set-Cookie response headers into the workspace jar.
// The jar works independently of history recording — a send that opts out of
// recording still participates in the jar.
func (s *HistoryService) IngestSetCookies(ctx context.Context, respHeaders map[string][]string, env string) {
	if s == nil || s.store == nil || len(respHeaders) == 0 {
		return
	}
	// Use net/http to parse Set-Cookie (handles Expires/Max-Age, Domain, Path, Secure)
	h := http.Header{}
	for k, vals := range respHeaders {
		for _, v := range vals {
			h.Add(k, v)
		}
	}
	for _, c := range (&http.Response{Header: h}).Cookies() {
		exp := c.Expires
		if exp.IsZero() && c.MaxAge > 0 {
			exp = time.Now().Add(time.Duration(c.MaxAge) * time.Second)
		}
		sameSite := ""
		switch c.SameSite {
		case http.SameSiteLaxMode:
			sameSite = "Lax"
		case http.SameSiteStrictMode:
			sameSite = "Strict"
		case http.SameSiteNoneMode:
			sameSite = "None"
		}
		_ = s.store.InsertCookie(ctx, history.Cookie{
			Name:      c.Name,
			Value:     c.Value,
			Domain:    c.Domain,
			Path:      c.Path,
			ExpiresAt: exp,
			Secure:    c.Secure,
			HttpOnly:  c.HttpOnly,
			SameSite:  sameSite,
			Env:       env,
		})
	}
}

// List returns masked entries.
func (s *HistoryService) List(ctx context.Context, limit, offset int, statusFilter *int) ([]history.Entry, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to view history")
	}
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
	if s == nil || s.store == nil {
		return history.Entry{}, fmt.Errorf("no workspace found: open a reqly workspace to view history")
	}
	e, err := s.store.Show(ctx, id)
	if err != nil {
		return history.Entry{}, err
	}
	maskEntry(&e)
	return e, nil
}

// ShowRaw returns one entry without masking headers. Replay uses it so a
// stored request is re-sent exactly as captured (Authorization included);
// display surfaces must keep using Show.
func (s *HistoryService) ShowRaw(ctx context.Context, id string) (history.Entry, error) {
	if s == nil || s.store == nil {
		return history.Entry{}, fmt.Errorf("no workspace found: open a reqly workspace to view history")
	}
	return s.store.Show(ctx, id)
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

// Cookies returns cookies for env.
func (s *HistoryService) Cookies(ctx context.Context, env string) ([]history.Cookie, error) {
	return s.store.ListCookies(ctx, env)
}

// DeleteCookie deletes one.
func (s *HistoryService) DeleteCookie(ctx context.Context, name, domain, path, env string) error {
	return s.store.DeleteCookie(ctx, name, domain, path, env)
}

// ClearCookies clears cookies.
func (s *HistoryService) ClearCookies(ctx context.Context, env *string) error {
	return s.store.ClearCookies(ctx, env)
}

// Close closes the underlying store.
func (s *HistoryService) Close() error { return s.store.Close() }

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
