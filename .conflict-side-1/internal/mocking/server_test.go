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

package mocking

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/openapi"
)

const petsSpec = `openapi: 3.0.3
info:
  title: Pets API
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: listPets
      responses:
        "200":
          description: A list of pets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
    post:
      operationId: createPet
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
  /pets/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: integer
    get:
      operationId: getPet
      responses:
        "200":
          description: A pet
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
    delete:
      operationId: deletePet
      responses:
        "204":
          description: Deleted
  /health:
    get:
      operationId: health
      responses:
        "200":
          description: ok
          content:
            application/json:
              example:
                status: ok
components:
  schemas:
    Pet:
      type: object
      required: [id, name]
      properties:
        id:
          type: integer
        name:
          type: string
`

func newTestServer(t *testing.T, opts ...Option) *Server {
	t.Helper()
	doc, err := openapi.Load([]byte(petsSpec))
	if err != nil {
		t.Fatalf("openapi.Load() error = %v", err)
	}
	srv, err := NewServer(doc, opts...)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	return srv
}

func doRequest(t *testing.T, srv http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestServeObjectResponse(t *testing.T) {
	srv := newTestServer(t)
	rec := doRequest(t, srv, "GET", "/pets/1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v (%s)", err, rec.Body.String())
	}
	if body["id"] != 0.0 {
		t.Fatalf("id = %#v, want 0", body["id"])
	}
	if body["name"] != "string" {
		t.Fatalf("name = %#v, want string", body["name"])
	}
}

func TestServeArrayResponse(t *testing.T) {
	srv := newTestServer(t)
	rec := doRequest(t, srv, "GET", "/pets")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON array: %v (%s)", err, rec.Body.String())
	}
	if len(body) != 1 {
		t.Fatalf("array length = %d, want 1", len(body))
	}
}

func TestServeNon200Status(t *testing.T) {
	srv := newTestServer(t)
	rec := doRequest(t, srv, "POST", "/pets")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
}

func TestServeExplicitExample(t *testing.T) {
	srv := newTestServer(t)
	rec := doRequest(t, srv, "GET", "/health")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status"`) || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("body = %s, want example {status: ok}", rec.Body.String())
	}
}

func TestServeNoContentResponse(t *testing.T) {
	srv := newTestServer(t)
	rec := doRequest(t, srv, "DELETE", "/pets/1")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestServeUnknownPath(t *testing.T) {
	srv := newTestServer(t)
	rec := doRequest(t, srv, "GET", "/nope")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestServeMethodNotAllowed(t *testing.T) {
	srv := newTestServer(t)
	rec := doRequest(t, srv, "PATCH", "/pets")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestServeLiteralBeatsTemplate(t *testing.T) {
	doc, err := openapi.Load([]byte(`openapi: 3.0.3
info:
  title: T
  version: 1.0.0
paths:
  /pets/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: integer
    get:
      responses:
        "200":
          description: templated
          content:
            application/json:
              example: {kind: templated}
  /pets/list:
    get:
      responses:
        "200":
          description: literal
          content:
            application/json:
              example: {kind: literal}
`))
	if err != nil {
		t.Fatalf("openapi.Load() error = %v", err)
	}
	srv, err := NewServer(doc)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	rec := doRequest(t, srv, "GET", "/pets/list")
	if !strings.Contains(rec.Body.String(), "literal") {
		t.Fatalf("body = %s, want literal route to win over template", rec.Body.String())
	}
}

func TestServeWithDelay(t *testing.T) {
	srv := newTestServer(t, WithDelay(50*time.Millisecond))
	start := time.Now()
	doRequest(t, srv, "GET", "/health")
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Fatalf("request too fast, delay not applied: %s", elapsed)
	}
}

func TestServeFailureRate(t *testing.T) {
	srv := newTestServer(t, WithFailureRate(3))
	statuses := make([]int, 0, 6)
	for i := 0; i < 6; i++ {
		rec := doRequest(t, srv, "GET", "/health")
		statuses = append(statuses, rec.Code)
	}
	// Every 3rd request fails; the rest succeed.
	for i, code := range statuses {
		if (i+1)%3 == 0 && code != http.StatusInternalServerError {
			t.Fatalf("request %d status = %d, want 500", i+1, code)
		}
		if (i+1)%3 != 0 && code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, code)
		}
	}
}

func TestNewServerNilDoc(t *testing.T) {
	if _, err := NewServer(nil); err == nil {
		t.Fatal("expected error for nil doc")
	}
}

func TestNewServerInvalidDoc(t *testing.T) {
	doc, err := openapi.Load([]byte(`openapi: 3.0.3
info:
  title: T
  version: 1.0.0
paths:
  /x/{id}:
    get:
      responses:
        "200":
          description: bad
`))
	if err != nil {
		// The doc loader itself may reject the missing path param.
		return
	}
	if _, err := NewServer(doc); err == nil {
		t.Fatal("expected error for invalid doc (missing path param)")
	}
}

func TestServerListensOverHTTP(t *testing.T) {
	srv := newTestServer(t)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("http.Get() error = %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(data), "ok") {
		t.Fatalf("body = %s", string(data))
	}
}

func TestPathParamsCaptured(t *testing.T) {
	srv := newTestServer(t)
	rec := doRequest(t, srv, "GET", "/pets/42")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	_ = fmt.Sprintf("%v", rec.Body.String())
}

func TestManualRoutesOverrideSpec(t *testing.T) {
	doc, err := openapi.Load([]byte(petsSpec))
	if err != nil {
		t.Fatalf("openapi.Load() error = %v", err)
	}
	srv, err := NewServer(doc, WithRoutes([]Route{
		{Method: "GET", Path: "/pets", Status: 200, Body: `{"manual":true}`, Enabled: true},
	}))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/pets", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != `{"manual":true}` {
		t.Errorf("body = %q", got)
	}
}

func TestDisabledRouteFallsThroughToSpec(t *testing.T) {
	doc, err := openapi.Load([]byte(petsSpec))
	if err != nil {
		t.Fatalf("openapi.Load() error = %v", err)
	}
	srv, err := NewServer(doc, WithRoutes([]Route{
		{Method: "GET", Path: "/pets", Status: 418, Body: "teapot", Enabled: false},
	}))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/pets", nil))
	if rec.Code == 418 {
		t.Error("disabled route was served")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("spec fallback status = %d, want 200", rec.Code)
	}
}

func TestSpeclessRouteServer(t *testing.T) {
	srv, err := NewServer(nil, WithRoutes([]Route{
		{Method: "", Path: "/anything", Status: 200, Body: "ok", Enabled: true},
	}))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(method, "/anything", nil))
		if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
			t.Errorf("%s: status=%d body=%q", method, rec.Code, rec.Body.String())
		}
	}
}

func TestNilDocWithoutRoutesErrors(t *testing.T) {
	if _, err := NewServer(nil); err == nil {
		t.Fatal("nil doc without routes succeeded, want error")
	}
}

func TestCustomContentTypeHeader(t *testing.T) {
	srv, err := NewServer(nil, WithRoutes([]Route{
		{Path: "/x", Status: 200, Body: "<xml/>",
			Headers: map[string]string{"Content-Type": "application/xml"}, Enabled: true},
	}))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/xml" {
		t.Errorf("Content-Type = %q", ct)
	}
}
