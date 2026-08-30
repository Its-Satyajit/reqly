// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"encoding/json"
	"fmt"

	"go.yaml.in/yaml/v3"
)

type rawSwagger2 struct {
	Swagger  string         `json:"swagger" yaml:"swagger"`
	Info     map[string]any `json:"info" yaml:"info"`
	Host     string         `json:"host" yaml:"host"`
	BasePath string         `json:"basePath" yaml:"basePath"`
	Schemes  []string       `json:"schemes" yaml:"schemes"`
	Paths    map[string]any `json:"paths" yaml:"paths"`
}

type openapi3Spec struct {
	OpenAPI string              `json:"openapi" yaml:"openapi"`
	Info    map[string]any      `json:"info" yaml:"info"`
	Servers []map[string]string `json:"servers" yaml:"servers"`
	Paths   map[string]any      `json:"paths" yaml:"paths"`
}

// ConvertSwagger2ToOpenAPI3 converts a Swagger 2.0 JSON or YAML document to an OpenAPI 3.0.3 YAML spec.
func ConvertSwagger2ToOpenAPI3(swaggerData []byte) ([]byte, error) {
	var s2 rawSwagger2
	if err := json.Unmarshal(swaggerData, &s2); err != nil {
		if err := yaml.Unmarshal(swaggerData, &s2); err != nil {
			return nil, fmt.Errorf("parse Swagger 2.0 spec: %w", err)
		}
	}
	if s2.Swagger != "2.0" {
		return nil, fmt.Errorf("not a Swagger 2.0 document")
	}

	scheme := "https"
	if len(s2.Schemes) > 0 && s2.Schemes[0] != "" {
		scheme = s2.Schemes[0]
	}
	serverURL := fmt.Sprintf("%s://%s%s", scheme, s2.Host, s2.BasePath)

	o3 := openapi3Spec{
		OpenAPI: "3.0.3",
		Info:    s2.Info,
		Servers: []map[string]string{{"url": serverURL}},
		Paths:   s2.Paths,
	}

	return yaml.Marshal(o3)
}
