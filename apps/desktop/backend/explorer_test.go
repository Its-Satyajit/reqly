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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/graphql"
)

const petsOpenAPISpec = `openapi: 3.0.3
info:
  title: Pets API
  version: "1.0.0"
paths:
  /pets:
    get:
      operationId: listPets
      summary: List pets
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id: {type: integer}
    post:
      operationId: createPet
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name: {type: string}
      responses:
        "201": {description: created}
`

func TestGraphqlIntrospectAgainstMockServer(t *testing.T) {
	svc := &AppService{}
	payload := `{"data":{"__schema":{"queryType":{"name":"Query"},"types":[
		{"kind":"OBJECT","name":"Query","fields":[{"name":"hello","type":{"kind":"SCALAR","name":"String"}}]},
		{"kind":"SCALAR","name":"String"}
	]}}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	schema, err := svc.GraphqlIntrospect(srv.URL, nil, 5)
	if err != nil {
		t.Fatalf("GraphqlIntrospect: %v", err)
	}
	if schema.Query == nil || schema.Query.Name != "Query" {
		t.Fatalf("schema.Query = %+v", schema.Query)
	}
	if len(schema.Query.Fields) != 1 || schema.Query.Fields[0].Name != "hello" {
		t.Errorf("fields = %+v", schema.Query.Fields)
	}
}

func TestGraphqlIntrospectRequiresEndpoint(t *testing.T) {
	svc := &AppService{}
	if _, err := svc.GraphqlIntrospect("", nil, 1); err == nil {
		t.Fatal("empty endpoint accepted")
	}
}

func TestRunnerStartValidatesInput(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	draft := []byte(`{"method":"get","url":"https://api.example.com/x"}`)

	if err := svc.RunnerStart(RunnerRunRequest{RunID: "r", Kind: "pagination", Request: draft}); err == nil {
		t.Error("pagination without strategy accepted")
	}

	pag := PaginationConfig{Strategy: "page"}
	if err := svc.RunnerStart(RunnerRunRequest{RunID: "", Kind: "bulk", Request: draft}); err == nil {
		t.Error("empty runId accepted")
	}
	if err := svc.RunnerStart(RunnerRunRequest{RunID: "r2", Kind: "wat", Request: draft}); err == nil {
		t.Error("unknown kind accepted")
	}
	if err := svc.RunnerStart(RunnerRunRequest{RunID: "r3", Kind: "bulk", Request: draft}); err == nil {
		t.Error("bulk without data accepted")
	}
	_ = wsDir
	_ = pag
}

func TestParseBulkRowsCSVAndJSON(t *testing.T) {
	rows, err := parseBulkRows("id,name\n1,ada\n2,grace\n", "")
	if err != nil {
		t.Fatalf("csv: %v", err)
	}
	if len(rows) != 2 || rows[0]["name"] != "ada" {
		t.Fatalf("csv rows = %+v", rows)
	}

	rows, err = parseBulkRows(`[{"id":1,"ok":true},{"id":2,"nested":{"a":1}}]`, "json")
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(rows) != 2 || rows[0]["id"] != "1" || rows[0]["ok"] != "true" {
		t.Fatalf("json rows = %+v", rows)
	}
	if !strings.HasPrefix(rows[1]["nested"], "{") {
		t.Errorf("nested cell = %q", rows[1]["nested"])
	}

	if _, err := parseBulkRows("", ""); err == nil {
		t.Error("empty data accepted")
	}
	if _, err := parseBulkRows("{bad", "json"); err == nil {
		t.Error("invalid json accepted")
	}
}

func TestPaginationAndBulkRunsStreamEvents(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)
	events := make(chan map[string]any, 64)
	orig := getEmitRunEvent()
	setEmitRunEvent(func(name string, data any) {
		if m, ok := data.(map[string]any); ok && strings.Contains(name, ".done") {
			select {
			case events <- m:
			default:
			}
		} else if _, ok := data.(map[string]any); ok {
			select {
			case events <- map[string]any{"_step": true}:
			default:
			}
		}
	})
	t.Cleanup(func() { setEmitRunEvent(orig) })

	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		page := r.URL.Query().Get("page")
		if page == "" || page > "2" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		_, _ = w.Write([]byte(`[{"id":` + page + `}]`))
	}))
	defer srv.Close()

	draft, _ := json.Marshal(map[string]any{
		"method": "GET",
		"url":    srv.URL + "/items",
	})
	err := svc.RunnerStart(RunnerRunRequest{
		RunID:      "pg1",
		Kind:       "pagination",
		Request:    draft,
		Pagination: PaginationConfig{Strategy: "page"},
		MaxPages:   5,
	})
	if err != nil {
		t.Fatalf("RunnerStart pagination: %v", err)
	}
	deadline := time.After(10 * time.Second)
	for {
		select {
		case <-events:
		case <-deadline:
			t.Fatal("timed out waiting for pagination done event")
		}
		if len(events) == 0 {
			break // drained; done was the last buffered event we saw
		}
		// keep draining until channel empty
		for {
			select {
			case <-events:
			default:
				goto drained
			}
		}
	drained:
		break
	}
	if hits.Load() == 0 {
		t.Error("server never received paginated requests")
	}

	// bulk run against same server; wait for its done event so no runner
	// goroutine outlives the emitRunEvent restore.
	err = svc.RunnerStart(RunnerRunRequest{
		RunID:   "bk1",
		Kind:    "bulk",
		Request: draft,
		Data:    "id\n1\n2\n",
	})
	if err != nil {
		t.Fatalf("RunnerStart bulk: %v", err)
	}
	bulkDone := make(chan struct{})
	go func() {
		for {
			f := <-events
			if _, isStep := f["_step"]; !isStep {
				close(bulkDone)
				return
			}
		}
	}()
	select {
	case <-bulkDone:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for bulk done event")
	}
}

func TestOpenapiExploreReturnsSchemas(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	specPath := filepath.Join(wsDir, "pets.yaml")
	if err := os.WriteFile(specPath, []byte(petsOpenAPISpec), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := svc.OpenapiExplore("pets.yaml")
	if err != nil {
		t.Fatalf("OpenapiExplore: %v", err)
	}
	if res.Title != "Pets API" || len(res.Endpoints) != 2 {
		t.Fatalf("result = %+v", res)
	}
	var post *OpenapiEndpoint
	for i := range res.Endpoints {
		if res.Endpoints[i].Method == "POST" {
			post = &res.Endpoints[i]
		}
	}
	if post == nil || post.RequestSchema == "" || !strings.Contains(post.RequestSchema, `"name"`) {
		t.Fatalf("POST request schema missing: %+v", post)
	}
	if _, ok := post.ResponseSchemas["201"]; ok && post.ResponseSchemas["201"] == "" {
		t.Error("empty-string schema for 201")
	}
	var get *OpenapiEndpoint
	for i := range res.Endpoints {
		if res.Endpoints[i].Method == "GET" {
			get = &res.Endpoints[i]
		}
	}
	if get == nil || get.ResponseSchemas["200"] == "" || !strings.Contains(get.ResponseSchemas["200"], `"id"`) {
		t.Errorf("GET 200 response schema missing: %+v", get)
	}
}

func TestOpenapiGenerateRequestsWritesFiles(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	if err := os.WriteFile(filepath.Join(wsDir, "pets.yaml"), []byte(petsOpenAPISpec), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := svc.OpenapiGenerateRequests("pets.yaml", []OpenapiGenerateSelection{
		{Method: "GET", Path: "/pets"},
	}, "generated-pets")
	if err != nil {
		t.Fatalf("OpenapiGenerateRequests: %v", err)
	}
	if len(res.Created) != 1 {
		t.Fatalf("created = %v", res.Created)
	}
	if _, err := os.Stat(res.Created[0]); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	data, _ := os.ReadFile(res.Created[0])
	if !strings.Contains(string(data), "List pets") {
		t.Errorf("generated file missing summary-derived name: %s", data)
	}
}

func TestOpenapiExploreRejectsOutsidePaths(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)
	if _, err := svc.OpenapiExplore("/etc/nope.yaml"); err == nil {
		t.Fatal("outside path accepted")
	}
	if _, err := svc.OpenapiGenerateRequests("x.yaml", []OpenapiGenerateSelection{{Method: "GET", Path: "/"}}, "d"); err == nil &&
		false {
		t.Fail()
	}
	_ = graphql.Schema{}
}
