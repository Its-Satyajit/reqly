// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
)

// HARReplayOptions configures sequential HAR execution and filtering.
type HARReplayOptions struct {
	Diff          bool   `json:"diff,omitempty"`
	IncludeStatic bool   `json:"includeStatic,omitempty"`
	FilterURL     string `json:"filterUrl,omitempty"`
	FilterMethod  string `json:"filterMethod,omitempty"`
}

// HARReplayEntryResult holds evaluation flags for a single replayed HAR entry.
type HARReplayEntryResult struct {
	URL            string   `json:"url"`
	Method         string   `json:"method"`
	OriginalStatus int      `json:"originalStatus"`
	ReplayedStatus int      `json:"replayedStatus"`
	StatusMatch    bool     `json:"statusMatch"`
	Diffs          []string `json:"diffs,omitempty"`
}

// HARReplayResult summarizes an entire HAR archive replay session.
type HARReplayResult struct {
	Total   int                    `json:"total"`
	Passed  int                    `json:"passed"`
	Failed  int                    `json:"failed"`
	Entries []HARReplayEntryResult `json:"entries"`
}

// ReplayHAR reads a HAR file and sequentially re-executes its request entries.
func ReplayHAR(ctx context.Context, harPath string, opts HARReplayOptions) (*HARReplayResult, error) {
	data, err := os.ReadFile(harPath)
	if err != nil {
		return nil, fmt.Errorf("read HAR file: %w", err)
	}

	var h harFile
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("parse HAR file: %w", err)
	}

	client := request.NewClient()
	res := &HARReplayResult{}

	for _, entry := range h.Log.Entries {
		reqURL := entry.Request.URL
		method := strings.ToUpper(entry.Request.Method)

		if !opts.IncludeStatic && isStaticAsset(reqURL) {
			continue
		}
		if opts.FilterMethod != "" && !strings.EqualFold(opts.FilterMethod, method) {
			continue
		}
		if opts.FilterURL != "" && !strings.Contains(reqURL, opts.FilterURL) {
			continue
		}

		req := &request.Request{
			Method: request.Method(method),
			URL:    reqURL,
		}

		for _, hdr := range entry.Request.Headers {
			if strings.HasPrefix(hdr.Name, ":") {
				continue
			}
			req.Headers = append(req.Headers, request.Header{Key: hdr.Name, Value: hdr.Value})
		}

		if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
			req.Body = entry.Request.PostData.Text
		}

		resp, err := client.Execute(ctx, req, nil)
		entryRes := HARReplayEntryResult{
			URL:            reqURL,
			Method:         method,
			OriginalStatus: entry.Response.Status,
		}

		if err == nil && resp != nil {
			entryRes.ReplayedStatus = resp.StatusCode
			entryRes.StatusMatch = (resp.StatusCode == entry.Response.Status)
		} else {
			entryRes.StatusMatch = false
		}

		if entryRes.StatusMatch {
			res.Passed++
		} else {
			res.Failed++
		}

		res.Total++
		res.Entries = append(res.Entries, entryRes)
	}

	return res, nil
}

func isStaticAsset(urlStr string) bool {
	lower := strings.ToLower(urlStr)
	staticExts := []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".css", ".js", ".woff", ".woff2", ".ico"}
	for _, ext := range staticExts {
		if strings.HasSuffix(lower, ext) || strings.Contains(lower, ext+"?") {
			return true
		}
	}
	return false
}
