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

package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func cannedSchema() string {
	return `{
	  "data": { "__schema": {
	    "queryType": { "name": "Query" },
	    "mutationType": { "name": "Mutation" },
	    "subscriptionType": null,
	    "types": [
	      { "kind": "SCALAR", "name": "String" },
	      { "kind": "OBJECT", "name": "Query", "fields": [
	        { "name": "user", "args": [ { "name": "id", "type": { "kind": "NON_NULL", "name": null, "ofType": { "kind": "SCALAR", "name": "ID" } } } ], "type": { "kind": "OBJECT", "name": "User" } },
	        { "name": "users", "args": [], "type": { "kind": "LIST", "name": null, "ofType": { "kind": "NON_NULL", "name": null, "ofType": { "kind": "OBJECT", "name": "User" } } } }
	      ] },
	      { "kind": "OBJECT", "name": "Mutation", "fields": [
	        { "name": "rename", "args": [ { "name": "name", "type": { "kind": "SCALAR", "name": "String" } } ], "type": { "kind": "NON_NULL", "name": null, "ofType": { "kind": "OBJECT", "name": "User" } } }
	      ] },
	      { "kind": "OBJECT", "name": "User", "fields": [
	        { "name": "id", "args": [], "type": { "kind": "NON_NULL", "name": null, "ofType": { "kind": "SCALAR", "name": "ID" } } },
	        { "name": "tags", "args": [], "type": { "kind": "LIST", "name": null, "ofType": { "kind": "SCALAR", "name": "String" } } }
	      ] },
	      { "kind": "ENUM", "name": "Role", "enumValues": [ { "name": "ADMIN" }, { "name": "USER" } ] }
	    ]
	  } }
	}`
}

func newIntrospectionServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestIntrospectParsesAndRenders(t *testing.T) {
	var gotAuth, gotCT string
	srv := newIntrospectionServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		w.Write([]byte(cannedSchema()))
	}))
	schema, _, err := Introspect(context.Background(), srv.URL, IntrospectOptions{
		Headers: [][2]string{{"Authorization", "Bearer t"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer t" || !strings.HasPrefix(gotCT, "application/json") {
		t.Fatalf("headers auth=%q ct=%q", gotAuth, gotCT)
	}
	if schema.Query == nil || len(schema.Query.Fields) != 2 {
		t.Fatalf("query root = %+v", schema.Query)
	}
	user := schema.Query.Fields[0]
	if user.Signature() != "user(id: ID!): User" {
		t.Fatalf("signature = %q", user.Signature())
	}
	users := schema.Query.Fields[1]
	if users.Type.String() != "[User!]" {
		t.Fatalf("[User!] wrapper = %q", users.Type.String())
	}
	rename := schema.Mutation.Fields[0]
	if rename.Type.String() != "User!" {
		t.Fatalf("mutation return = %q", rename.Type.String())
	}
}

func TestSummaryRendering(t *testing.T) {
	schema, _, err := Introspect(context.Background(), newIntrospectionServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(cannedSchema()))
	})).URL, IntrospectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	full := schema.Summary("")
	for _, want := range []string{"query\n", "  user(id: ID!): User", "mutation\n", "enum Role { ADMIN | USER }", "type User\n"} {
		if !strings.Contains(full, want) {
			t.Errorf("summary missing %q:\n%s", want, full)
		}
	}
	// Roots must appear before the type listing.
	qi := strings.Index(full, "query\n")
	ui := strings.Index(full, "type User\n")
	if qi < 0 || ui < 0 || qi > ui {
		t.Fatalf("roots should render before types:\n%s", full)
	}

	filtered := schema.Summary("User")
	if !strings.Contains(filtered, "id: ID!") || strings.Contains(filtered, "query") {
		t.Errorf("--type filter failed:\n%s", filtered)
	}
}

func TestIntrospectErrorPaths(t *testing.T) {
	// GraphQL-level errors in a 200 response.
	errSrv := newIntrospectionServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"errors":[{"message":"unauthorized"}]}`))
	}))
	_, _, err := Introspect(context.Background(), errSrv.URL, IntrospectOptions{})
	if err == nil || !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("err = %v, want GraphQL errors message", err)
	}

	// Malformed JSON.
	badSrv := newIntrospectionServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	_, _, err = Introspect(context.Background(), badSrv.URL, IntrospectOptions{})
	if err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("err = %v, want malformed error", err)
	}

	// Non-2xx.
	codeSrv := newIntrospectionServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	_, _, err = Introspect(context.Background(), codeSrv.URL, IntrospectOptions{})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("err = %v, want HTTP status error", err)
	}
}

func TestRawJSONRoundTrip(t *testing.T) {
	srv := newIntrospectionServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(cannedSchema()))
	}))
	_, raw, err := Introspect(context.Background(), srv.URL, IntrospectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var pretty map[string]any
	if err := json.Unmarshal(raw, &pretty); err != nil {
		t.Fatal(err)
	}
}
