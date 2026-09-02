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
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/history"
)

func seedSpec(t *testing.T, wsDir, name, spec string) string {
	t.Helper()
	path := filepath.Join(wsDir, name)
	if err := os.WriteFile(path, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	return name
}

const diffSpecA = `openapi: 3.0.3
info: {title: T, version: "1"}
paths:
  /pets:
    get:
      responses:
        "200": {description: ok}
`

const diffSpecB = diffSpecA + `  /dogs:
    get:
      responses:
        "200": {description: ok}
`

func TestDiffSpecsClassifiesSeverity(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	a := seedSpec(t, wsDir, "a.yaml", diffSpecA)
	b := seedSpec(t, wsDir, "b.yaml", diffSpecB)

	res, err := svc.DiffSpecs(a, b)
	if err != nil {
		t.Fatalf("DiffSpecs: %v", err)
	}
	if !res.Result.HasChanges {
		t.Fatal("expected changes")
	}
	if res.Addition == 0 {
		t.Errorf("Addition = %d, want > 0", res.Addition)
	}
	if res.Breaking != 0 {
		t.Errorf("Breaking = %d for pure addition, want 0", res.Breaking)
	}
}

func TestDiffSpecsOutsideWorkspaceRejected(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	a := seedSpec(t, wsDir, "a.yaml", diffSpecA)
	if _, err := svc.DiffSpecs(a, "/etc/passwd.yaml"); err == nil {
		t.Fatal("diff with path outside workspace succeeded, want error")
	}
}

func TestDiffResponsesComparesBodies(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)
	h := svc.hist()
	if h == nil {
		t.Fatal("no history store")
	}
	entryA := history.Entry{
		RequestPath: "pets/list", Method: http.MethodGet,
		URL: "https://api.example.com/pets", Status: 200, Env: "test",
		RespBody: []byte(`{"n":1}`),
	}
	if err := h.Record(context.Background(), &entryA); err != nil {
		t.Fatalf("record A: %v", err)
	}
	entryB := history.Entry{
		RequestPath: "pets/list", Method: http.MethodGet,
		URL: "https://api.example.com/pets", Status: 200, Env: "test",
		RespBody: []byte(`{"n":2,"extra":true}`),
	}
	if err := h.Record(context.Background(), &entryB); err != nil {
		t.Fatalf("record B: %v", err)
	}
	idA, idB := entryA.ID, entryB.ID

	res, err := svc.DiffResponses(idA, idB)
	if err != nil {
		t.Fatalf("DiffResponses: %v", err)
	}
	if !res.Result.HasChanges {
		t.Fatal("expected response changes")
	}
	if res.MetaA.Status != 200 || res.MetaB.Status != 200 {
		t.Errorf("meta statuses = %d/%d", res.MetaA.Status, res.MetaB.Status)
	}
}

func TestDiffJSONTextYAMLAndJSON(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)
	res, err := svc.DiffJSONText("a: 1\nb: 2\n", `{"a": 1, "b": 3}`)
	if err != nil {
		t.Fatalf("DiffJSONText: %v", err)
	}
	if !res.HasChanges || len(res.Changes) != 1 {
		t.Fatalf("changes = %+v, want exactly one", res)
	}
	if _, err := svc.DiffJSONText("{invalid", `{"a":1}`); err == nil {
		t.Fatal("invalid document accepted")
	}
}
