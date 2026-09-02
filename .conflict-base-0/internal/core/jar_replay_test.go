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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
)

func jarWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "reqly.yaml"), []byte("name: ws\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRun_AttachesJarCookies(t *testing.T) {
	dir := jarWorkspace(t)
	var sawCookie atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("sid"); err == nil {
			sawCookie.Store(c.Value)
		}
		http.SetCookie(w, &http.Cookie{Name: "sid", Value: "s3cret"})
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	svc := NewRunService(dir)
	defer svc.Close()

	// First send ingests Set-Cookie into the workspace jar.
	noRecord := false
	if _, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet, URL: srv.URL,
	}, RunRequestOptions{RecordHistory: &noRecord}); err != nil {
		t.Fatal(err)
	}
	// Second send must carry the jar's cookie.
	h := http.Header{}
	if _, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet, URL: srv.URL,
	}, RunRequestOptions{RecordHistory: &noRecord}); err != nil {
		t.Fatal(err)
	}
	if got := sawCookie.Load(); got == nil || got.(string) != "s3cret" {
		t.Fatalf("expected jar cookie attached on second send, got %v", got)
	}
	_ = h
}

func TestRun_AttachCookiesOptOut(t *testing.T) {
	dir := jarWorkspace(t)
	var sawCookie atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("sid"); err == nil {
			sawCookie.Store(c.Value)
		} else {
			sawCookie.Store("<none>")
		}
		http.SetCookie(w, &http.Cookie{Name: "sid", Value: "s3cret"})
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	svc := NewRunService(dir)
	defer svc.Close()
	noRecord := false
	for i := 0; i < 2; i++ {
		off := false
		if _, err := svc.Run(context.Background(), request.Request{
			Method: request.MethodGet, URL: srv.URL,
		}, RunRequestOptions{RecordHistory: &noRecord, AttachCookies: &off}); err != nil {
			t.Fatal(err)
		}
	}
	if got := sawCookie.Load(); got != nil && got.(string) == "s3cret" {
		t.Fatal("jar cookie should not attach when opted out")
	}
}

func TestHistoryShowRawVsMasked(t *testing.T) {
	dir := jarWorkspace(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	record := true
	svc := NewRunService(dir)
	defer svc.Close()
	if _, err := svc.Run(context.Background(), request.Request{
		Method: request.MethodGet,
		URL:    srv.URL,
		Headers: []request.Header{
			{Key: "Authorization", Value: "Bearer top-secret"},
			{Key: "X-Trace", Value: "abc"},
		},
	}, RunRequestOptions{RecordHistory: &record}); err != nil {
		t.Fatal(err)
	}

	h := svc.History()
	entries, err := h.List(context.Background(), 1, 0, nil)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one entry, got %v (%v)", entries, err)
	}
	id := entries[0].ID

	masked, err := h.Show(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(masked.ReqHeaders["Authorization"], "") != "[SECRET]" {
		t.Fatalf("Show must mask Authorization, got %q", strings.Join(masked.ReqHeaders["Authorization"], ""))
	}
	raw, err := h.ShowRaw(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(raw.ReqHeaders["Authorization"], "") != "Bearer top-secret" {
		t.Fatalf("ShowRaw must keep Authorization exact, got %q", strings.Join(raw.ReqHeaders["Authorization"], ""))
	}
}
