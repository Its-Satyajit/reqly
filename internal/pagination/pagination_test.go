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

package pagination

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
)

func makeResp(status int, body string, headers map[string][]string) *response.Response {
	return &response.Response{StatusCode: status, Headers: headers, Body: []byte(body)}
}

func TestRun_PageStrategy(t *testing.T) {
	req := request.Request{
		Method: "GET",
		URL:    "https://api.example.com/items",
		Pagination: &request.Pagination{
			Strategy: "page",
			MaxPages: 3,
		},
	}
	var steps []Step
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		// page param should increment
		v := ""
		for _, q := range r.Query {
			if q.Key == "page" {
				v = q.Value
			}
		}
		// return empty on page 3 to trigger stop
		if v == "3" {
			return makeResp(200, "[]", nil), nil
		}
		return makeResp(200, `[{"id":1}]`, nil), nil
	}
	err := Run(context.Background(), req, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("steps: got %d want 3", len(steps))
	}
	// verify page increments
	for i, s := range steps {
		want := strconv.Itoa(i + 1)
		got := ""
		for _, q := range s.Request.Query {
			if q.Key == "page" {
				got = q.Value
			}
		}
		if got != want {
			t.Fatalf("step %d page: got %q want %q", i+1, got, want)
		}
	}
}

func TestRun_OffsetStrategy(t *testing.T) {
	req := request.Request{
		Method: "GET",
		URL:    "https://api.example.com/items",
		Pagination: &request.Pagination{
			Strategy: "offset",
			Limit:    10,
			MaxPages: 3,
		},
	}
	var steps []Step
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		v := ""
		for _, q := range r.Query {
			if q.Key == "offset" {
				v = q.Value
			}
		}
		if v == "20" {
			return makeResp(200, "[]", nil), nil
		}
		return makeResp(200, `[{"id":1}]`, nil), nil
	}
	err := Run(context.Background(), req, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("steps: got %d want 3", len(steps))
	}
	wantOffsets := []string{"0", "10", "20"}
	for i, s := range steps {
		got := ""
		for _, q := range s.Request.Query {
			if q.Key == "offset" {
				got = q.Value
			}
		}
		if got != wantOffsets[i] {
			t.Fatalf("step %d offset: got %q want %q", i+1, got, wantOffsets[i])
		}
	}
}

func TestRun_CursorStrategy(t *testing.T) {
	req := request.Request{
		Method: "GET",
		URL:    "https://api.example.com/items",
		Pagination: &request.Pagination{
			Strategy:    "cursor",
			NextPath:    "$.nextCursor",
			MaxPages:    3,
			CursorParam: "cursor",
		},
	}
	cursors := []string{"", "a", "b"}
	var steps []Step
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		idx := len(steps)
		cursor := cursors[idx]
		_ = cursor
		// verify cursor param for step >1
		if idx > 0 {
			got := ""
			for _, q := range r.Query {
				if q.Key == "cursor" {
					got = q.Value
				}
			}
			if got != cursors[idx] {
				return nil, fmt.Errorf("cursor: got %q want %q", got, cursors[idx])
			}
		}
		if idx == 2 {
			// last page, no nextCursor
			return makeResp(200, `{"items":[{"id":3}], "nextCursor": ""}`, nil), nil
		}
		body, _ := json.Marshal(map[string]any{"items": []any{map[string]any{"id": idx + 1}}, "nextCursor": cursors[idx+1]})
		return makeResp(200, string(body), nil), nil
	}
	err := Run(context.Background(), req, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("steps: got %d want 3", len(steps))
	}
}

func TestRun_LinkHeader(t *testing.T) {
	req := request.Request{
		Method: "GET",
		URL:    "https://api.example.com/items?page=1",
		Pagination: &request.Pagination{
			Strategy: "link-header",
			MaxPages: 2,
		},
	}
	var steps []Step
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		if len(steps) == 0 {
			h := map[string][]string{"Link": {`<https://api.example.com/items?page=2>; rel="next"`}}
			return makeResp(200, `[{"id":1}]`, h), nil
		}
		// second step should have URL from Link
		if r.URL != "https://api.example.com/items?page=2" {
			return nil, fmt.Errorf("link-header next URL: got %q", r.URL)
		}
		return makeResp(200, `[]`, nil), nil
	}
	err := Run(context.Background(), req, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("steps: got %d want 2", len(steps))
	}
}

func TestRun_StopOnEmpty(t *testing.T) {
	req := request.Request{
		Method:     "GET",
		URL:        "https://api.example.com/items",
		Pagination: &request.Pagination{Strategy: "page", MaxPages: 10},
	}
	sendFn := func(_ context.Context, _ request.Request) (*response.Response, error) {
		return makeResp(200, "[]", nil), nil
	}
	var steps []Step
	_ = Run(context.Background(), req, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if len(steps) != 1 {
		t.Fatalf("empty should stop after 1 step, got %d", len(steps))
	}
}

func TestRun_StopOnNon2xx(t *testing.T) {
	req := request.Request{
		Method:     "GET",
		URL:        "https://api.example.com/items",
		Pagination: &request.Pagination{Strategy: "page", MaxPages: 10},
	}
	sendFn := func(_ context.Context, _ request.Request) (*response.Response, error) {
		return makeResp(500, "error", nil), nil
	}
	var steps []Step
	_ = Run(context.Background(), req, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if len(steps) != 1 {
		t.Fatalf("non-2xx should stop after 1, got %d", len(steps))
	}
}

func TestRun_MaxPages(t *testing.T) {
	req := request.Request{
		Method:     "GET",
		URL:        "https://api.example.com/items",
		Pagination: &request.Pagination{Strategy: "page", MaxPages: 2},
	}
	sendFn := func(_ context.Context, _ request.Request) (*response.Response, error) {
		return makeResp(200, `[{"id":1}]`, nil), nil
	}
	var steps []Step
	_ = Run(context.Background(), req, Options{MaxPages: 2}, sendFn, func(s Step) { steps = append(steps, s) })
	if len(steps) != 2 {
		t.Fatalf("maxPages 2: got %d", len(steps))
	}
}

func TestRun_MaxPagesOverride(t *testing.T) {
	req := request.Request{
		Method:     "GET",
		URL:        "https://api.example.com/items",
		Pagination: &request.Pagination{Strategy: "page", MaxPages: 100},
	}
	sendFn := func(_ context.Context, _ request.Request) (*response.Response, error) {
		return makeResp(200, `[{"id":1}]`, nil), nil
	}
	var steps []Step
	_ = Run(context.Background(), req, Options{MaxPages: 1}, sendFn, func(s Step) { steps = append(steps, s) })
	if len(steps) != 1 {
		t.Fatalf("override maxPages 1: got %d", len(steps))
	}
}

func TestRun_CursorMissingNextPathError(t *testing.T) {
	req := request.Request{
		Method:     "GET",
		URL:        "https://api.example.com/items",
		Pagination: &request.Pagination{Strategy: "cursor"},
	}
	err := Run(context.Background(), req, Options{}, func(_ context.Context, _ request.Request) (*response.Response, error) {
		return makeResp(200, `{}`, nil), nil
	}, nil)
	if err == nil || !contains(err.Error(), "nextPath") {
		t.Fatalf("expected nextPath error, got %v", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (sub == "" || len(s) > 0 && (func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}

// Ensure link header parsing handles multiple links
func TestRun_LinkHeaderMultiple(t *testing.T) {
	h := http.Header{}
	h.Set("Link", `<https://api.example.com/items?page=2>; rel="next", <https://api.example.com/items?page=5>; rel="last"`)
	m := map[string][]string(h)
	next := extractLinkNext(m)
	if next != "https://api.example.com/items?page=2" {
		t.Fatalf("next: got %q", next)
	}
}

func TestRun_PageStrategy_URLTemplateVariable(t *testing.T) {
	req := request.Request{
		Method: "GET",
		URL:    "https://api.example.com/items?page={{page}}",
		Pagination: &request.Pagination{
			Strategy: "page",
			MaxPages: 2,
		},
	}
	var steps []Step
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		return makeResp(200, `[{"id":1}]`, nil), nil
	}
	err := Run(context.Background(), req, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0].Request.URL != "https://api.example.com/items?page=1" {
		t.Fatalf("step 1 URL: got %q, want 'https://api.example.com/items?page=1'", steps[0].Request.URL)
	}
	if steps[1].Request.URL != "https://api.example.com/items?page=2" {
		t.Fatalf("step 2 URL: got %q, want 'https://api.example.com/items?page=2'", steps[1].Request.URL)
	}
}
