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

package collections

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// Inherited is the configuration a request inherits from its container chain
// (Workspace → Collection → Folder), resolved bottom-up with overrides applied.
type Inherited struct {
	// BaseURL is the resolved base URL, or "" when no container set one.
	BaseURL string
	// Headers merged from all containers; the closest container wins per key.
	Headers []request.Header
	// Auth inherited from the closest container that defines one.
	Auth request.Auth
}

// VariablesSet builds a variables.Set with the workspace (global) and, when
// present, collection/folder variable scopes.
func (w *Workspace) VariablesSet() *variables.Set {
	set := variables.NewSet()
	for key, value := range w.Config.Variables {
		set.Set(variables.ScopeGlobal, key, value)
	}
	return set
}

// Resolve computes the inherited configuration for a request inside a
// collection, applying the Workspace → Collection → Folder override chain.
func (w *Workspace) Resolve(c *Collection, chain []*Folder) Inherited {
	resolved := Inherited{}

	// Workspace level.
	resolved.apply(&w.Config)

	// Collection level.
	resolved.apply(&c.Config)

	// Folder levels, outermost first.
	for _, f := range chain {
		resolved.apply(&f.Config)
	}

	return resolved
}

// apply merges one container's config onto the inherited state. Headers are
// merged per key (child wins); auth replaces only when non-empty; baseURL is
// joined or replaced as described by Config.BaseURL.
func (i *Inherited) apply(cfg *Config) {
	i.Headers = mergeHeaders(i.Headers, cfg.Headers)

	if cfg.Auth.Type != "" {
		i.Auth = cfg.Auth
	}

	if cfg.BaseURL != "" {
		i.BaseURL = resolveBaseURL(i.BaseURL, cfg.BaseURL)
	}
}

// mergeHeaders appends child headers onto parent, overriding existing keys.
func mergeHeaders(parent, child []request.Header) []request.Header {
	if len(child) == 0 {
		return parent
	}
	merged := append([]request.Header{}, parent...)
	seen := make(map[string]int, len(merged))
	for idx, h := range merged {
		seen[strings.ToLower(h.Key)] = idx
	}
	for _, h := range child {
		key := strings.ToLower(h.Key)
		if idx, ok := seen[key]; ok {
			merged[idx] = h
		} else {
			merged = append(merged, h)
			seen[key] = len(merged) - 1
		}
	}
	return merged
}

// resolveBaseURL combines a parent base URL with a child value. An absolute
// child value replaces the parent; a relative value is joined.
func resolveBaseURL(parent, child string) string {
	if parent == "" {
		return child
	}
	if strings.Contains(child, "://") {
		return child
	}
	base := strings.TrimSuffix(parent, "/")
	ref := strings.TrimPrefix(child, "/")
	return base + "/" + ref
}

// EffectiveURL joins the request's URL onto the inherited base URL. An
// absolute request URL is used as-is. The join is performed as raw strings so
// {{variable}} placeholders survive until interpolation.
func (i *Inherited) EffectiveURL(requestURL string) (string, error) {
	if requestURL == "" {
		return "", fmt.Errorf("request has no URL")
	}
	if strings.Contains(requestURL, "://") {
		return requestURL, nil
	}
	if i.BaseURL == "" {
		return requestURL, nil
	}
	if _, err := url.Parse(i.BaseURL); err != nil {
		return "", fmt.Errorf("invalid base URL %q: %w", i.BaseURL, err)
	}
	base := strings.TrimSuffix(i.BaseURL, "/")
	return base + "/" + strings.TrimPrefix(requestURL, "/"), nil
}

// applyTo copies the inherited base URL and auth onto a request, and merges
// the inherited headers with the request's own headers (request wins).
func (i *Inherited) applyTo(r *request.Request) {
	if r.Method == "" {
		r.Method = request.MethodGet
	}
	if r.Auth.Type == "" && i.Auth.Type != "" {
		r.Auth = i.Auth
	}
	r.Headers = mergeHeaders(i.Headers, r.Headers)
}

// ResolvedRequest is a request file combined with its inherited configuration
// and full variable scope chain, ready for execution.
type ResolvedRequest struct {
	// Request is the request with inherited base URL, headers, and auth
	// applied, and the URL made absolute against the base URL.
	Request request.Request
	// Vars is the merged variable scope chain (workspace → collection →
	// folder → request).
	Vars *variables.Set
}

// ResolveRequest resolves entry against the workspace, applying the full
// inheritance chain and building the variable set. chain must list the folder
// ancestors outermost-first (the same order returned by RequestEntry.Path).
func (w *Workspace) ResolveRequest(c *Collection, chain []*Folder, entry *RequestEntry) (*ResolvedRequest, error) {
	inherited := w.Resolve(c, chain)

	req := entry.File.Request
	inherited.applyTo(&req)

	effectiveURL, err := inherited.EffectiveURL(req.URL)
	if err != nil {
		return nil, err
	}
	req.URL = effectiveURL

	vars := w.VariablesSet()
	for key, value := range c.Config.Variables {
		vars.Set(variables.ScopeCollection, key, value)
	}
	for _, f := range chain {
		for key, value := range f.Config.Variables {
			vars.Set(variables.ScopeFolder, key, value)
		}
	}
	for key, value := range entry.File.Variables {
		vars.Set(variables.ScopeRequest, key, value)
	}

	return &ResolvedRequest{Request: req, Vars: vars}, nil
}
