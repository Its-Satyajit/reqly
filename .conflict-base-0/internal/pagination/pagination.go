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
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
)

// Options configures a pagination run (mirrors Request.Pagination but allows override).
type Options struct {
	MaxPages int // 0 means 100 default
}

// Step is one iteration of the pagination loop.
type Step struct {
	Index    int
	Request  request.Request
	Response *response.Response
	Err      error
	Next     string
}

// SendFunc sends a request and returns a response.
type SendFunc func(context.Context, request.Request) (*response.Response, error)

// Run iteratively fetches paginated pages until a structural stop.
func Run(ctx context.Context, req request.Request, opts Options, sendFn SendFunc, onStep func(Step)) error {
	if req.Pagination == nil {
		return fmt.Errorf("pagination: missing pagination config")
	}
	pg := req.Pagination
	strategy := strings.ToLower(pg.Strategy)
	if strategy == "" {
		return fmt.Errorf("pagination: strategy required")
	}
	maxPages := pg.MaxPages
	if opts.MaxPages > 0 {
		maxPages = opts.MaxPages
	}
	if maxPages <= 0 {
		maxPages = 100
	}
	// defaults
	pageParam := pg.PageParam
	if pageParam == "" {
		pageParam = "page"
	}
	offsetParam := pg.OffsetParam
	if offsetParam == "" {
		offsetParam = "offset"
	}
	limitParam := pg.LimitParam
	if limitParam == "" {
		limitParam = "limit"
	}
	cursorParam := pg.CursorParam
	if cursorParam == "" {
		cursorParam = "cursor"
	}
	pageSize := pg.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	limit := pg.Limit
	if limit <= 0 {
		limit = 20
	}
	// cursor strategy requires nextPath
	if strategy == "cursor" && pg.NextPath == "" {
		return fmt.Errorf("pagination: cursor strategy requires nextPath")
	}

	curReq := req
	// For page/offset we seed initial values if not present.
	initialPage := 1
	initialOffset := 0
	// Try to preserve existing query values? For page, if query already has page param, respect it as start.
	if strategy == "page" {
		if v := queryValue(curReq.Query, pageParam); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				initialPage = n
			}
		} else {
			curReq = withQuery(curReq, pageParam, strconv.Itoa(initialPage))
			// also set pageSize if not present?
			if queryValue(curReq.Query, pg.PageSizeParam) == "" && pg.PageSizeParam != "" {
				// use pageSizeParam if set
			}
			if pg.PageSizeParam != "" {
				if queryValue(curReq.Query, pg.PageSizeParam) == "" {
					curReq = withQuery(curReq, pg.PageSizeParam, strconv.Itoa(pageSize))
				}
			} else {
				// default pageSize not required; but set if not present? Keep optional.
			}
		}
	}
	if strategy == "offset" {
		if v := queryValue(curReq.Query, offsetParam); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				initialOffset = n
			}
		} else {
			curReq = withQuery(curReq, offsetParam, strconv.Itoa(initialOffset))
			if queryValue(curReq.Query, limitParam) == "" {
				curReq = withQuery(curReq, limitParam, strconv.Itoa(limit))
			}
		}
	}

	page := initialPage
	offset := initialOffset

	for i := 1; i <= maxPages; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp, err := sendFn(ctx, curReq)
		step := Step{Index: i, Request: curReq, Response: resp, Err: err}
		if err != nil {
			if onStep != nil {
				onStep(step)
			}
			return err
		}
		if resp != nil && resp.StatusCode != 0 && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
			step.Next = ""
			if onStep != nil {
				onStep(step)
			}
			break
		}
		// Extract next for cursor/link-header
		next := ""
		if strategy == "cursor" && resp != nil {
			next = extractJSONPath(resp.Body, pg.NextPath)
		} else if strategy == "link-header" && resp != nil {
			next = extractLinkNext(resp.Headers)
		}
		step.Next = next

		// Empty body check (structural stop) — only when next missing and body is empty array
		isEmpty := isEmptyBody(resp)
		if isEmpty && next == "" {
			if onStep != nil {
				onStep(step)
			}
			break
		}
		if onStep != nil {
			onStep(step)
		}
		// Decide if should continue
		shouldContinue := true
		if strategy == "cursor" || strategy == "link-header" {
			if next == "" {
				shouldContinue = false
			}
		} else {
			// page/offset: continue until empty body
			if isEmpty {
				shouldContinue = false
			}
		}
		if !shouldContinue {
			break
		}
		if i == maxPages {
			break
		}
		// Prepare next request
		nextReq := curReq
		switch strategy {
		case "page":
			page++
			nextReq = withQuery(curReq, pageParam, strconv.Itoa(page))
		case "offset":
			offset += limit
			nextReq = withQuery(curReq, offsetParam, strconv.Itoa(offset))
		case "cursor":
			nextReq = withQuery(curReq, cursorParam, next)
		case "link-header":
			// next is full URL
			nextReq.URL = next
			// keep headers etc, but URL overrides query
		default:
			return fmt.Errorf("pagination: unknown strategy %q", strategy)
		}
		curReq = nextReq
	}
	return nil
}

func withQuery(req request.Request, key, value string) request.Request {
	out := req
	// copy slice first to avoid mutating shared backing array
	out.Query = append([]request.Parameter(nil), req.Query...)
	found := false
	for i, q := range out.Query {
		if q.Key == key {
			out.Query[i].Value = value
			found = true
			break
		}
	}
	if !found {
		out.Query = append(out.Query, request.Parameter{Key: key, Value: value})
	}
	return out
}

func queryValue(q []request.Parameter, key string) string {
	for _, p := range q {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

func isEmptyBody(resp *response.Response) bool {
	if resp == nil {
		return true
	}
	b := strings.TrimSpace(string(resp.Body))
	if b == "" {
		return true
	}
	if b == "[]" {
		return true
	}
	// check {"items":[]} etc? For M30, only [] is empty; object with empty items still has next check.
	// Try JSON array detection
	var arr []any
	if err := json.Unmarshal(resp.Body, &arr); err == nil {
		return len(arr) == 0
	}
	return false
}

func extractJSONPath(body []byte, path string) string {
	if path == "" {
		return ""
	}
	// support $.field or $.a.b
	p := strings.TrimSpace(path)
	if p == "$" {
		return strings.TrimSpace(string(body))
	}
	if strings.HasPrefix(p, "$.") {
		p = p[2:]
	} else if strings.HasPrefix(p, "$") {
		p = p[1:]
		if strings.HasPrefix(p, ".") {
			p = p[1:]
		}
	}
	fields := strings.Split(p, ".")
	var cur any
	if err := json.Unmarshal(body, &cur); err != nil {
		return ""
	}
	for _, f := range fields {
		if f == "" {
			continue
		}
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[f]
		if cur == nil {
			return ""
		}
	}
	switch v := cur.(type) {
	case string:
		return v
	case float64:
		// numeric cursor? stringify without trailing .0
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case nil:
		return ""
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func extractLinkNext(headers map[string][]string) string {
	// headers map may have canonical keys; find Link case-insensitive
	var vals []string
	for k, v := range headers {
		if strings.EqualFold(k, "Link") {
			vals = append(vals, v...)
		}
		// also check http.Header canonical "Link"
		if k == "Link" {
			// already handled
		}
	}
	// Also try direct lookup via http.Header semantics (case-insensitive already via EqualFold)
	if len(vals) == 0 {
		// try "link" lower
		for k, v := range headers {
			if strings.EqualFold(k, http.CanonicalHeaderKey("Link")) {
				vals = v
				break
			}
		}
	}
	for _, h := range vals {
		// Link header may contain multiple comma-separated links
		parts := strings.Split(h, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			// <url>; rel="next"
			if strings.Contains(part, `rel="next"`) || strings.Contains(part, `rel='next'`) || strings.Contains(part, `rel=next`) {
				// extract <...>
				start := strings.Index(part, "<")
				end := strings.Index(part, ">")
				if start >= 0 && end > start {
					return strings.TrimSpace(part[start+1 : end])
				}
			}
		}
	}
	return ""
}
