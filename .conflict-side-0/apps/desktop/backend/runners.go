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

package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/bulk"
	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/pagination"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
)

// PaginationConfig mirrors request.Pagination for the config panel.
type PaginationConfig struct {
	Strategy      string `json:"strategy"`
	PageParam     string `json:"pageParam,omitempty"`
	PageSizeParam string `json:"pageSizeParam,omitempty"`
	OffsetParam   string `json:"offsetParam,omitempty"`
	LimitParam    string `json:"limitParam,omitempty"`
	CursorParam   string `json:"cursorParam,omitempty"`
	NextPath      string `json:"nextPath,omitempty"`
	MaxPages      int    `json:"maxPages,omitempty"`
	PageSize      int    `json:"pageSize,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

// RunnerRunRequest drives both runner kinds from one draft payload; Kind
// selects pagination vs bulk.
type RunnerRunRequest struct {
	RunID       string           `json:"runId"`
	Kind        string           `json:"kind"` // "pagination" | "bulk"
	Request     json.RawMessage  `json:"request"`
	Pagination  PaginationConfig `json:"pagination,omitempty"`
	MaxPages    int              `json:"maxPagesOverride,omitempty"`
	Data        string           `json:"data,omitempty"`
	DataFormat  string           `json:"dataFormat,omitempty"`
	Parallel    bool             `json:"parallel,omitempty"`
	Concurrency int              `json:"concurrency,omitempty"`
}

// runnerStep is one streamed step event payload.
type runnerStep struct {
	RunIndex    int    `json:"index"`
	Status      int    `json:"status,omitempty"`
	Error       string `json:"error,omitempty"`
	URL         string `json:"url,omitempty"`
	BodyPreview string `json:"bodyPreview,omitempty"`
}

// draftRequest decodes the editor's TabDraft JSON into a core request.Request.
func draftRequest(raw json.RawMessage) (request.Request, error) {
	var draft struct {
		Method  string `json:"method"`
		URL     string `json:"url"`
		Headers []struct {
			Key     string `json:"key"`
			Value   string `json:"value"`
			Enabled bool   `json:"enabled"`
		} `json:"headers"`
		Params []struct {
			Key     string `json:"key"`
			Value   string `json:"value"`
			Enabled bool   `json:"enabled"`
		} `json:"params"`
		BodyType string `json:"bodyType"`
		Body     string `json:"body"`
	}
	if err := json.Unmarshal(raw, &draft); err != nil {
		return request.Request{}, fmt.Errorf("decode draft: %w", err)
	}
	req := request.Request{
		Method: request.Method(strings.ToUpper(strings.TrimSpace(draft.Method))),
		URL:    strings.TrimSpace(draft.URL),
	}
	for _, h := range draft.Headers {
		if h.Enabled && h.Key != "" {
			req.Headers = append(req.Headers, request.Header{Key: h.Key, Value: h.Value})
		}
	}
	for _, p := range draft.Params {
		if p.Enabled && p.Key != "" {
			req.Query = append(req.Query, request.Parameter{Key: p.Key, Value: p.Value})
		}
	}
	switch strings.ToLower(draft.BodyType) {
	case "form-data", "urlencoded", "":
		// key-value bodies are not interpolated by runners yet
	default:
		req.Body = draft.Body
	}
	return req, nil
}

func runnerSendFn(s *AppService, env string) func(context.Context, request.Request) (*response.Response, error) {
	noRecord := false
	return func(ctx context.Context, r request.Request) (*response.Response, error) {
		res, err := s.requests.Run(ctx, r, core.RunRequestOptions{
			EnvFlag:       env,
			RecordHistory: &noRecord,
		})
		if err != nil {
			return nil, err
		}
		return res.Response, nil
	}
}

func preview(body []byte) string {
	const cap = 2000
	if len(body) > cap {
		return string(body[:cap])
	}
	return string(body)
}

// RunnerStart launches an async run. Steps stream on
// `reqly.runner.<runId>.step` and the summary on `.done` — same shape as the
// collection Run View's event protocol.
func (s *AppService) RunnerStart(req RunnerRunRequest) error {
	if s == nil || s.requests == nil {
		return fmt.Errorf("no workspace found: open a reqly workspace to run")
	}
	runID := strings.TrimSpace(req.RunID)
	if runID == "" {
		return fmt.Errorf("runId is required")
	}
	r, err := draftRequest(req.Request)
	if err != nil {
		return err
	}
	sendFn := runnerSendFn(s, "")
	emit := func(suffix string, payload any) {
		emitEvent("reqly.runner."+runID+suffix, payload)
	}

	switch req.Kind {
	case "pagination":
		if strings.TrimSpace(req.Pagination.Strategy) == "" {
			return fmt.Errorf("pagination strategy is required")
		}
		r.Pagination = &request.Pagination{
			Strategy:      req.Pagination.Strategy,
			PageParam:     req.Pagination.PageParam,
			PageSizeParam: req.Pagination.PageSizeParam,
			OffsetParam:   req.Pagination.OffsetParam,
			LimitParam:    req.Pagination.LimitParam,
			CursorParam:   req.Pagination.CursorParam,
			NextPath:      req.Pagination.NextPath,
			MaxPages:      req.Pagination.MaxPages,
			PageSize:      req.Pagination.PageSize,
			Limit:         req.Pagination.Limit,
		}
		go func() {
			var steps, failures int
			var lastBody []byte
			err := pagination.Run(context.Background(), r, pagination.Options{MaxPages: req.MaxPages},
				sendFn,
				func(st pagination.Step) {
					steps++
					frame := runnerStep{RunIndex: st.Index}
					if st.Err != nil {
						failures++
						frame.Error = st.Err.Error()
					} else if st.Response != nil {
						frame.Status = st.Response.StatusCode
						frame.URL = st.Request.URL
						frame.BodyPreview = preview(st.Response.Body)
						lastBody = st.Response.Body
					}
					emit(".step", frame)
				})
			done := map[string]any{"steps": steps, "failures": failures}
			if err != nil {
				done["error"] = err.Error()
			} else if lastBody != nil {
				done["lastBody"] = string(lastBody)
			}
			emit(".done", done)
		}()
		return nil

	case "bulk":
		rows, err := parseBulkRows(req.Data, req.DataFormat)
		if err != nil {
			return err
		}
		go func() {
			opts := bulk.Options{
				Parallel:        req.Parallel,
				Concurrency:     req.Concurrency,
				ContinueOnError: true,
			}
			var okCount, failCount int
			err := bulk.Run(context.Background(), r, rows, opts, sendFn,
				func(st bulk.Step) {
					frame := runnerStep{RunIndex: st.Index}
					if st.Err != nil {
						failCount++
						frame.Error = st.Err.Error()
					} else if st.Response != nil {
						okCount++
						frame.Status = st.Response.StatusCode
						frame.URL = st.Request.URL
						frame.BodyPreview = preview(st.Response.Body)
					}
					emit(".step", frame)
				})
			done := map[string]any{"rows": len(rows), "passed": okCount, "failed": failCount}
			if err != nil {
				done["error"] = err.Error()
			}
			emit(".done", done)
		}()
		return nil

	default:
		return fmt.Errorf("unknown kind %q: pick pagination or bulk", req.Kind)
	}
}

// parseBulkRows accepts a JSON array of objects or CSV with a header row.
func parseBulkRows(data string, format string) ([]map[string]string, error) {
	text := strings.TrimSpace(data)
	if text == "" {
		return nil, fmt.Errorf("no data rows given")
	}
	if format == "" {
		format = "csv"
		if strings.HasPrefix(text, "[") {
			format = "json"
		}
	}
	switch format {
	case "json":
		var rows []map[string]string
		// tolerate numeric/bool cells by decoding into interface{} first
		var generic []map[string]any
		if err := json.Unmarshal([]byte(text), &generic); err != nil {
			return nil, fmt.Errorf("parse bulk JSON: %w", err)
		}
		for _, m := range generic {
			row := make(map[string]string, len(m))
			for k, v := range m {
				row[k] = cellString(v)
			}
			rows = append(rows, row)
		}
		return rows, nil
	case "csv":
		records, err := csv.NewReader(strings.NewReader(text)).ReadAll()
		if err != nil {
			return nil, fmt.Errorf("parse bulk CSV: %w", err)
		}
		if len(records) < 2 {
			return nil, nil
		}
		header := records[0]
		var rows []map[string]string
		for _, rec := range records[1:] {
			row := make(map[string]string, len(header))
			for i, h := range header {
				if i < len(rec) {
					row[strings.TrimSpace(h)] = rec[i]
				}
			}
			rows = append(rows, row)
		}
		return rows, nil
	default:
		return nil, fmt.Errorf("unknown data format %q: pick csv or json", format)
	}
}

func cellString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

// RunnerCancel cancels nothing yet — runner cancellation is deferred; the
// method exists so the frontend can call it unconditionally.
func (s *AppService) RunnerCancel(runID string) {}
