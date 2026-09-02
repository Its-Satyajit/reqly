// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/jsonschema"
)

// SchemaValidateResult is the bridge result for schema validate.
type SchemaValidateResult struct {
	Valid      bool                    `json:"valid"`
	Violations []jsonschema.Violation `json:"violations"`
}

// SchemaValidate validates an instance against a schema.
func (s *AppService) SchemaValidate(schemaJSON, instanceJSON, draft string) (*SchemaValidateResult, error) {
	if strings.TrimSpace(schemaJSON) == "" {
		return nil, fmt.Errorf("schema is required")
	}
	sch, err := jsonschema.Compile([]byte(schemaJSON), draft)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	violations, err := jsonschema.Validate(sch, []byte(instanceJSON))
	if err != nil {
		return nil, err
	}
	return &SchemaValidateResult{Valid: len(violations) == 0, Violations: violations}, nil
}

// SchemaInspect renders a schema as a text tree.
func (s *AppService) SchemaInspect(schemaJSON string) (string, error) {
	if strings.TrimSpace(schemaJSON) == "" {
		return "", fmt.Errorf("schema is required")
	}
	return jsonschema.Inspect([]byte(schemaJSON))
}

// SchemaGenerate synthesizes a sample instance for a schema.
func (s *AppService) SchemaGenerate(schemaJSON string, seed int64) (string, error) {
	if strings.TrimSpace(schemaJSON) == "" {
		return "", fmt.Errorf("schema is required")
	}
	out, warnings, err := jsonschema.Generate([]byte(schemaJSON), jsonschema.GenerateOptions{Seed: seed, IncludeOptional: true})
	if err != nil {
		return "", err
	}
	if len(warnings) > 0 {
		return string(out) + "\n\n// warnings: " + strings.Join(warnings, "; "), nil
	}
	return string(out), nil
}
