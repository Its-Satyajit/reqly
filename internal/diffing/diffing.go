// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package diffing

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/r3labs/diff/v3"

	"github.com/Its-Satyajit/reqly/internal/openapi"
)

// DiffResult holds structural diff information.
type DiffResult struct {
	HasChanges bool     `json:"hasChanges"`
	Changes    []Change `json:"changes,omitempty"`
}

type Change struct {
	Type string      `json:"type"` // "create", "update", "delete"
	Path []string    `json:"path"`
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
}

// JSON computes structural diffs between two JSON/YAML byte slices.
func JSON(a, b []byte) (*DiffResult, error) {
	var objA, objB interface{}

	if err := json.Unmarshal(a, &objA); err != nil {
		return nil, fmt.Errorf("unmarshal first JSON: %w", err)
	}
	if err := json.Unmarshal(b, &objB); err != nil {
		return nil, fmt.Errorf("unmarshal second JSON: %w", err)
	}

	changelog, err := diff.Diff(objA, objB)
	if err != nil {
		return nil, fmt.Errorf("computing diff: %w", err)
	}

	var changes []Change
	for _, c := range changelog {
		changes = append(changes, Change{
			Type: c.Type,
			Path: c.Path,
			From: c.From,
			To:   c.To,
		})
	}

	return &DiffResult{
		HasChanges: len(changes) > 0,
		Changes:    changes,
	}, nil
}

// OpenAPI computes structural diffs between two OpenAPI specification documents.
func OpenAPI(docA, docB *openapi3.T) (*DiffResult, error) {
	bytesA, err := docA.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal docA: %w", err)
	}
	bytesB, err := docB.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal docB: %w", err)
	}

	return JSON(bytesA, bytesB)
}

// OpenAPIFiles loads and diffs two OpenAPI specification files.
func OpenAPIFiles(pathA, pathB string) (*DiffResult, error) {
	docA, err := openapi.LoadFile(pathA)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", pathA, err)
	}
	docB, err := openapi.LoadFile(pathB)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", pathB, err)
	}

	return OpenAPI(docA, docB)
}
