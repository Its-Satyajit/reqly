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

package openapi

import (
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// Endpoint is a single operation of a spec, flattened for exploration.
type Endpoint struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	OperationID string   `json:"operationId,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Summary     string   `json:"summary,omitempty"`
}

// methodRank orders methods within one path in the conventional
// get/post/put/patch/delete/head/options order used across the explorer.
var methodRank = map[string]int{
	"GET": 0, "POST": 1, "PUT": 2, "PATCH": 3, "DELETE": 4, "HEAD": 5, "OPTIONS": 6,
}

// Explore flattens a document into a deterministic endpoint list, ordered by
// path (lexicographic) then by conventional method order.
func Explore(doc *openapi3.T) []Endpoint {
	var eps []Endpoint
	for _, path := range doc.Paths.InMatchingOrder() {
		item := doc.Paths.Value(path)
		for method, op := range item.Operations() {
			eps = append(eps, Endpoint{
				Method:      method,
				Path:        path,
				OperationID: op.OperationID,
				Tags:        append([]string(nil), op.Tags...),
				Summary:     op.Summary,
			})
		}
	}
	sort.SliceStable(eps, func(i, j int) bool {
		if eps[i].Path != eps[j].Path {
			return eps[i].Path < eps[j].Path
		}
		return methodRank[eps[i].Method] < methodRank[eps[j].Method]
	})
	return eps
}

// FilterByTag keeps only endpoints whose tag list contains tag.
func FilterByTag(eps []Endpoint, tag string) []Endpoint {
	var out []Endpoint
	for _, ep := range eps {
		for _, t := range ep.Tags {
			if t == tag {
				out = append(out, ep)
				break
			}
		}
	}
	return out
}

// selectEndpoints resolves GenerateOptions selectors against doc and returns
// the matched operations with their path-item context.
func selectEndpoints(doc *openapi3.T, opts GenerateOptions) ([]Endpoint, error) {
	all := Explore(doc)
	if !opts.All && len(opts.Operations) == 0 && len(opts.Tags) == 0 && opts.Method == "" && opts.Path == "" {
		return nil, fmt.Errorf("no operation selected; use --operation <id>, --method/--path, --tag, or --all; available operations:\n%s", summarizeOperations(all))
	}

	if opts.All {
		return all, nil
	}

	byOpID := map[string]Endpoint{}
	for _, ep := range all {
		if ep.OperationID != "" {
			byOpID[ep.OperationID] = ep
		}
	}

	selected := map[string]bool{}
	key := func(ep Endpoint) string { return ep.Method + "|" + ep.Path }
	add := func(ep Endpoint) { selected[key(ep)] = true }

	for _, id := range opts.Operations {
		ep, ok := byOpID[id]
		if !ok {
			return nil, fmt.Errorf("unknown operation %q; available operations:\n%s", id, summarizeOperations(all))
		}
		add(ep)
	}
	for _, tag := range opts.Tags {
		matched := FilterByTag(all, tag)
		if len(matched) == 0 {
			return nil, fmt.Errorf("no operations carry tag %q", tag)
		}
		for _, ep := range matched {
			add(ep)
		}
	}
	if opts.Method != "" || opts.Path != "" {
		if opts.Method == "" || opts.Path == "" {
			return nil, fmt.Errorf("--method and --path must be given together")
		}
		method := upper(opts.Method)
		found := false
		for _, ep := range all {
			if ep.Method == method && ep.Path == opts.Path {
				add(ep)
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("no %s operation on path %q", method, opts.Path)
		}
	}

	out := make([]Endpoint, 0, len(selected))
	for _, ep := range all {
		if selected[key(ep)] {
			out = append(out, ep)
		}
	}
	return out, nil
}

// summarizeOperations renders the endpoint list as an indented help block.
func summarizeOperations(eps []Endpoint) string {
	s := ""
	for _, ep := range eps {
		id := ep.OperationID
		if id == "" {
			id = "(no operationId)"
		}
		s += fmt.Sprintf("  %-7s %-30s %s\n", ep.Method, ep.Path, id)
	}
	return s
}

func upper(s string) string {
	out := []byte(s)
	for i := range out {
		if out[i] >= 'a' && out[i] <= 'z' {
			out[i] -= 'a' - 'A'
		}
	}
	return string(out)
}
