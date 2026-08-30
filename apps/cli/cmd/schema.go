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

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"

	"github.com/Its-Satyajit/reqly/internal/jsonschema"
	"github.com/Its-Satyajit/reqly/internal/validation"
)

var (
	schemaValidateDraft   string
	schemaValidateType    string
	schemaValidateJSON    bool
	schemaInspectJSON     bool
	schemaGenerateSeed    int64
	schemaGenerateOptIncl bool
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "JSON Schema tooling",
	Long:  "Validate documents against JSON Schemas, inspect schemas, and generate sample instances.",
}

var schemaValidateCmd = &cobra.Command{
	Use:   "validate <schema> [instance|-]",
	Short: "Validate a JSON/YAML document against a JSON Schema",
	Long: `Check an instance document against a JSON Schema and report every
violation at its instance path. The instance may be a file path or - for
stdin; omitting it reads stdin.

  reqly schema validate user.schema.json payload.json
  curl -s api.test/users | reqly schema validate user.schema.json
  reqly schema validate s.yaml i.yaml --json`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaData, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read schema: %w", err)
		}
		isXSD := schemaValidateType == "xml" || strings.HasSuffix(args[0], ".xsd")
		if isXSD {
			instanceData, err := readInstance(args)
			if err != nil {
				return fmt.Errorf("read instance: %w", err)
			}
			res, err := validation.ValidateXMLAgainstXSD(instanceData, schemaData, validation.ValidationOptions{})
			if err != nil {
				return fmt.Errorf("validate xsd: %w", err)
			}
			if schemaValidateJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}
			for _, v := range res.Errors {
				fmt.Fprintln(cmd.OutOrStdout(), v.Message)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d violation(s)\n", len(res.Errors))
			if !res.Valid {
				return fmt.Errorf("schema validation failed with %d violation(s)", len(res.Errors))
			}
			return nil
		}

		sch, err := jsonschema.Compile(schemaData, schemaValidateDraft)
		if err != nil {
			return fmt.Errorf("compile schema: %w", err)
		}
		instanceData, err := readInstance(args)
		if err != nil {
			return fmt.Errorf("read instance: %w", err)
		}
		violations, err := jsonschema.Validate(sch, instanceData)
		if err != nil {
			return fmt.Errorf("validate instance: %w", err)
		}
		if schemaValidateJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			if err := enc.Encode(violations); err != nil {
				return fmt.Errorf("encode json: %w", err)
			}
		} else {
			for _, v := range violations {
				fmt.Fprintln(cmd.OutOrStdout(), v.String())
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d violation(s)\n", len(violations))
		}
		if len(violations) > 0 {
			return fmt.Errorf("schema validation failed with %d violation(s)", len(violations))
		}
		return nil
	},
}

var schemaInspectCmd = &cobra.Command{
	Use:   "inspect <schema> [--json]",
	Short: "Inspect the structure and keywords of a JSON Schema",
	Long: `Display a human-readable outline of a JSON Schema, or dump the parsed
keyword map as JSON with --json.

  reqly schema inspect user.schema.json
  reqly schema inspect user.schema.yaml --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read schema: %w", err)
		}
		if schemaInspectJSON {
			var doc any
			if err := yaml.Unmarshal(data, &doc); err != nil {
				return fmt.Errorf("parse schema: %w", err)
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(doc)
		}
		out, err := jsonschema.Inspect(data)
		if err != nil {
			return fmt.Errorf("inspect schema: %w", err)
		}
		fmt.Fprint(cmd.OutOrStdout(), out)
		return nil
	},
}

var schemaGenerateCmd = &cobra.Command{
	Use:   "generate <schema> [--seed n] [--optional]",
	Short: "Generate a sample instance document from a JSON Schema",
	Long: `Synthesize a sample JSON document honoring the schema's examples,
defaults, enums, formats, and numeric/string constraints. Deterministic by
default; --seed varies generated choices; --optional includes optional
properties. Unresolvable constraints produce warnings on stderr.

  reqly schema generate user.schema.yaml | reqly run -u https://api.test/users`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read schema: %w", err)
		}
		instance, warnings, err := jsonschema.Generate(data, jsonschema.GenerateOptions{
			Seed:            schemaGenerateSeed,
			IncludeOptional: schemaGenerateOptIncl,
		})
		if err != nil {
			return fmt.Errorf("generate sample: %w", err)
		}
		for _, w := range warnings {
			fmt.Fprintln(cmd.ErrOrStderr(), "warning:", w)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(instance))
		return nil
	},
}

func readInstance(args []string) ([]byte, error) {
	if len(args) < 2 || args[1] == "-" {
		return io.ReadAll(os.Stdin)
	}
	data, err := os.ReadFile(args[1])
	if err != nil {
		return nil, fmt.Errorf("read instance: %w", err)
	}
	return data, nil
}

func init() {
	schemaValidateCmd.Flags().StringVar(&schemaValidateDraft, "draft", "", "override $schema draft detection (2020, 2019, 7, 6, 4)")
	schemaValidateCmd.Flags().StringVar(&schemaValidateType, "type", "", "schema type (json|xml)")
	schemaValidateCmd.Flags().BoolVar(&schemaValidateJSON, "json", false, "print violations as JSON")
	schemaInspectCmd.Flags().BoolVar(&schemaInspectJSON, "json", false, "dump the schema keyword map as JSON")

	schemaGenerateCmd.Flags().Int64Var(&schemaGenerateSeed, "seed", 0, "seed for varied generation (default deterministic)")
	schemaGenerateCmd.Flags().BoolVar(&schemaGenerateOptIncl, "optional", false, "include optional properties")

	schemaCmd.AddCommand(schemaValidateCmd, schemaInspectCmd, schemaGenerateCmd)
}
