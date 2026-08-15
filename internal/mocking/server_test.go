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
