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

package openapi

import (
	"context"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

// Load parses an OpenAPI 3.x document (JSON or YAML) from raw bytes, resolves
// Load parses, resolves internal references in, and validates an OpenAPI document.
// It returns the document or an error describing the parsing or validation failure.
func Load(data []byte) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("parsing OpenAPI: %w", err)
	}
	if err := doc.Validate(context.Background(), openapi3.DisableExamplesValidation()); err != nil {
		return nil, fmt.Errorf("validating OpenAPI: %w", err)
	}
	return doc, nil
}

// LoadFile reads an OpenAPI 3.x document from a file (JSON or YAML), resolves
// LoadFile reads and parses an OpenAPI document from the specified file.
// It returns an error if the file cannot be read or the document cannot be parsed or validated.
func LoadFile(path string) (*openapi3.T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return Load(data)
}
