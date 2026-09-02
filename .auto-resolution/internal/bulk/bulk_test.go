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

package bulk

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
)

func makeRespBulk(status int, body string) *response.Response {
	return &response.Response{StatusCode: status, Body: []byte(body)}
}

func TestRun_Sequential(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://api.example.com/users/{{id}}"}
	rows := []map[string]string{{"id": "1"}, {"id": "2"}, {"id": "3"}}
	var steps []Step
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		return makeRespBulk(200, `ok`), nil
	}
	err := Run(context.Background(), req, rows, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("steps: got %d want 3", len(steps))
	}
	for i, s := range steps {
		want := fmt.Sprintf("https://api.example.com/users/%d", i+1)
		if s.Request.URL != want {
			t.Fatalf("step %d URL: got %q want %q", i+1, s.Request.URL, want)
		}
	}
}

func TestRun_ParallelOrdered(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://api.example.com/users/{{id}}"}
	rows := []map[string]string{{"id": "1"}, {"id": "2"}, {"id": "3"}, {"id": "4"}, {"id": "5"}}
	var steps []Step
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		return makeRespBulk(200, `ok`), nil
	}
	err := Run(context.Background(), req, rows, Options{Parallel: true, Concurrency: 2}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(steps) != 5 {
		t.Fatalf("steps: got %d want 5", len(steps))
	}
	// ordered by input index
	for i, s := range steps {
		want := rows[i]["id"]
		if !strings.Contains(s.Request.URL, want) {
			t.Fatalf("step %d URL: got %q want contain %q", i+1, s.Request.URL, want)
		}
		if s.Index != i+1 {
			t.Fatalf("index: got %d want %d", s.Index, i+1)
		}
	}
}

func TestRun_StopOnError(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://api.example.com/users/{{id}}"}
	rows := []map[string]string{{"id": "1"}, {"id": "2"}, {"id": "3"}}
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		if strings.Contains(r.URL, "/2") {
			return makeRespBulk(500, `err`), nil
		}
		return makeRespBulk(200, `ok`), nil
	}
	var steps []Step
	err := Run(context.Background(), req, rows, Options{}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// should stop after step 2 (500) without continue
	if len(steps) != 2 {
		t.Fatalf("steps: got %d want 2", len(steps))
	}
}

func TestRun_ContinueOnError(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://api.example.com/users/{{id}}"}
	rows := []map[string]string{{"id": "1"}, {"id": "2"}, {"id": "3"}}
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		if strings.Contains(r.URL, "/2") {
			return makeRespBulk(500, `err`), nil
		}
		return makeRespBulk(200, `ok`), nil
	}
	var steps []Step
	err := Run(context.Background(), req, rows, Options{ContinueOnError: true}, sendFn, func(s Step) { steps = append(steps, s) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("continue steps: got %d want 3", len(steps))
	}
}

func TestRun_InterpolateQueryAndBody(t *testing.T) {
	req := request.Request{
		Method: "POST",
		URL:    "https://api.example.com/users",
		Query:  []request.Parameter{{Key: "id", Value: "{{id}}"}},
		Body:   `{"name":"{{name}}"}`,
	}
	rows := []map[string]string{{"id": "1", "name": "alice"}}
	var gotReq request.Request
	sendFn := func(_ context.Context, r request.Request) (*response.Response, error) {
		gotReq = r
		return makeRespBulk(200, `ok`), nil
	}
	_ = Run(context.Background(), req, rows, Options{}, sendFn, nil)
	if len(gotReq.Query) != 1 || gotReq.Query[0].Value != "1" {
		t.Fatalf("query interpolation: got %v", gotReq.Query)
	}
	if gotReq.Body != `{"name":"alice"}` {
		t.Fatalf("body interpolation: got %q", gotReq.Body)
	}
}

func TestRun_EmptyRows(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://example.com"}
	err := Run(context.Background(), req, nil, Options{}, func(_ context.Context, _ request.Request) (*response.Response, error) {
		t.Fatal("should not be called")
		return nil, nil
	}, nil)
	if err != nil {
		t.Fatalf("empty rows: %v", err)
	}
}
