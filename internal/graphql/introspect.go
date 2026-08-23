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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IntrospectOptions carries transport settings for Introspect.
type IntrospectOptions struct {
	// Headers are sent verbatim on the introspection POST (auth, etc.).
	Headers [][2]string
	// Timeout bounds the request; 0 means no explicit timeout.
	Timeout time.Duration
}

// TypeRef is a wrapped GraphQL type reference.
type TypeRef struct {
	Name string   // set on named types
	Kind string   // LIST / NON_NULL / named kind
	Of   *TypeRef // inner type for wrappers
}

// String renders the reference GraphQL-style: [String!]!, ID, [Episode!].
func (t *TypeRef) String() string {
	if t == nil {
		return ""
	}
	switch t.Kind {
	case "LIST":
		return "[" + t.Of.String() + "]"
	case "NON_NULL":
		inner := t.Of.String()
		if strings.HasSuffix(inner, "!") {
			return inner
		}
		return inner + "!"
	default:
		return t.Name
	}
}

// Arg is one field argument.
type Arg struct {
	Name string
	Type *TypeRef
	Def  string // default value literal, "" when none
}

// Field is one object/interface field.
type Field struct {
	Name string
	Type *TypeRef
	Args []Arg
}

// Type is a schema type (object, scalar, enum, interface, union, input).
type Type struct {
	Kind        string
	Name        string
	Description string
	Fields      []Field // objects/interfaces only
	EnumValues  []string
}

// Schema is an introspected GraphQL schema.
type Schema struct {
	Query        *Type
	Mutation     *Type
	Subscription *Type
	Types        []*Type // non-builtin types, sorted by name
}

const introspectionQuery = `query IntrospectionQuery { __schema { queryType { name } mutationType { name } subscriptionType { name } types { kind name description fields(includeDeprecated: true) { name args { name defaultValue type { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } type { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } enumValues(includeDeprecated: true) { name } } } }`

type rawRef struct {
	Kind   string  `json:"kind"`
	Name   string  `json:"name"`
	OfType *rawRef `json:"ofType"`
}

func (r rawRef) ref() *TypeRef {
	out := &TypeRef{Kind: r.Kind, Name: r.Name}
	if r.OfType != nil {
		out.Of = r.OfType.ref()
	}
	return out
}

type introspectionPayload struct {
	Data struct {
		Schema struct {
			Query        map[string]string `json:"queryType"`
			Mutation     map[string]string `json:"mutationType"`
			Subscription map[string]string `json:"subscriptionType"`
			Types        []struct {
				Kind        string `json:"kind"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Fields      []struct {
					Name string `json:"name"`
					Args []struct {
						Name    string `json:"name"`
						Default any    `json:"defaultValue"`
						Type    rawRef `json:"type"`
					} `json:"args"`
					Type rawRef `json:"type"`
				} `json:"fields"`
				EnumValues []struct {
					Name string `json:"name"`
				} `json:"enumValues"`
			} `json:"types"`
		} `json:"__schema"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// rootName returns the type name from a queryType/mutationType payload entry.
func rootName(m map[string]string) string {
	if m == nil {
		return ""
	}
	return m["name"]
}

// Introspect runs the standard introspection query against endpoint and
// returns the parsed schema plus the raw response body.
func Introspect(ctx context.Context, endpoint string, opts IntrospectOptions) (*Schema, json.RawMessage, error) {
	body, err := json.Marshal(map[string]string{"query": introspectionQuery})
	if err != nil {
		return nil, nil, err
	}
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, h := range opts.Headers {
		req.Header.Set(h[0], h[1])
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("introspection request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(raw))
		if len(snippet) > 300 {
			snippet = snippet[:300] + "…"
		}
		if snippet != "" {
			return nil, raw, fmt.Errorf("endpoint returned HTTP %d: %s", resp.StatusCode, snippet)
		}
		return nil, raw, fmt.Errorf("endpoint returned HTTP %d", resp.StatusCode)
	}
	var p introspectionPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, raw, fmt.Errorf("malformed introspection response: %w", err)
	}
	if len(p.Errors) > 0 {
		msgs := make([]string, 0, len(p.Errors))
		for _, e := range p.Errors {
			msgs = append(msgs, e.Message)
		}
		return nil, raw, fmt.Errorf("GraphQL errors: %s", strings.Join(msgs, "; "))
	}
	return buildSchema(&p), json.RawMessage(raw), nil
}

func buildSchema(p *introspectionPayload) *Schema {
	s := &Schema{}
	builtin := map[string]bool{
		"String": true, "Int": true, "Float": true, "Boolean": true, "ID": true,
	}
	rootByName := map[string]**Type{
		rootName(p.Data.Schema.Query):        &s.Query,
		rootName(p.Data.Schema.Mutation):     &s.Mutation,
		rootName(p.Data.Schema.Subscription): &s.Subscription,
	}
	for _, t := range p.Data.Schema.Types {
		if strings.HasPrefix(t.Name, "__") || builtin[t.Name] {
			continue
		}
		typ := &Type{Kind: t.Kind, Name: t.Name, Description: t.Description}
		for _, f := range t.Fields {
			field := Field{Name: f.Name, Type: f.Type.ref()}
			for _, a := range f.Args {
				arg := Arg{Name: a.Name, Type: a.Type.ref()}
				if a.Default != nil {
					arg.Def = fmt.Sprintf("%v", a.Default)
				}
				field.Args = append(field.Args, arg)
			}
			typ.Fields = append(typ.Fields, field)
		}
		for _, ev := range t.EnumValues {
			typ.EnumValues = append(typ.EnumValues, ev.Name)
		}
		if dst, ok := rootByName[t.Name]; ok && t.Name != "" {
			*dst = typ
		} else {
			s.Types = append(s.Types, typ)
		}
	}
	sortTypes(s.Types)
	return s
}
