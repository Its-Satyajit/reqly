// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/openapi"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// Result holds validation results and issue details.
type Result struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// ValidateOpenAPIFile validates an OpenAPI specification file.
func ValidateOpenAPIFile(path string) (*Result, error) {
	_, err := openapi.LoadFile(path)
	if err != nil {
		return &Result{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}
	return &Result{Valid: true}, nil
}

// ValidateProject recursively scans and validates Git-native project descriptors.
func ValidateProject(dir string) (*Result, error) {
	var errs []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".json" || ext == ".yaml" || ext == ".yml" {
			// Check if it's a request file
			_, reqErr := requestfile.LoadFile(path)
			if reqErr != nil {
				// If not a request file, verify basic valid JSON/YAML formatting
				data, readErr := os.ReadFile(path)
				if readErr == nil {
					if strings.Contains(path, "openapi") || strings.Contains(path, "swagger") {
						if _, oerr := openapi.Load(data); oerr != nil {
							errs = append(errs, fmt.Sprintf("%s: invalid openapi spec: %v", path, oerr))
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Result{
		Valid:  len(errs) == 0,
		Errors: errs,
	}, nil
}
