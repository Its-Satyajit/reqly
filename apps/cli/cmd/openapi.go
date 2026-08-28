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
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/openapi"
)

var (
	openapiExploreTags []string
	openapiExploreJSON bool

	openapiGenerateOps    []string
	openapiGenerateTags   []string
	openapiGenerateMethod string
	openapiGeneratePath   string
	openapiGenerateAll    bool
	openapiGenerateOut    string
)

var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "OpenAPI spec tooling",
	Long:  "Explore OpenAPI specs and generate runnable request files from them.",
}

var openapiExploreCmd = &cobra.Command{
	Use:   "explore <spec> [--tag <t>]... [--json]",
	Short: "List the operations of an OpenAPI spec",
	Long: `Print a table of every operation in an OpenAPI 3.x document:
method, path, operationId, first tag, and summary.

  reqly openapi explore api.yaml
  reqly openapi explore api.yaml --tag users
  reqly openapi explore api.yaml --json | jq '.[].operationId'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := openapi.LoadFile(args[0])
		if err != nil {
			return err
		}
		endpoints := openapi.Explore(doc)
		for _, tag := range openapiExploreTags {
			endpoints = openapi.FilterByTag(endpoints, tag)
		}
		if openapiExploreJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(endpoints)
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "METHOD\tPATH\tOPERATION ID\tTAG\tSUMMARY")
		for _, ep := range endpoints {
			firstTag := ""
			if len(ep.Tags) > 0 {
				firstTag = ep.Tags[0]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ep.Method, ep.Path, ep.OperationID, firstTag, ep.Summary)
		}
		return w.Flush()
	},
}

var openapiGenerateCmd = &cobra.Command{
	Use:   "generate <spec> [--operation <id>]... | [--method <m> --path <p>] | [--tag <t>]... | --all [--output dir]",
	Short: "Generate runnable request files for selected operations",
	Long: `Render selected operations of an OpenAPI 3.x document as native request
files (YAML). Bodies and parameters are filled inline from spec examples and
defaults; unresolved required parameters stay as literal placeholders.
Bearer/api-key security becomes a native auth block with placeholder
variables; unmappable schemes are skipped with a warning.

  reqly openapi generate api.yaml --operation createUser --operation getUser
  reqly openapi generate api.yaml --tag users --output requests/users
  reqly openapi generate api.yaml --all`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := openapi.LoadFile(args[0])
		if err != nil {
			return err
		}
		opts := openapi.GenerateOptions{
			All:        openapiGenerateAll,
			Operations: openapiGenerateOps,
			Tags:       openapiGenerateTags,
			Method:     openapiGenerateMethod,
			Path:       openapiGeneratePath,
		}
		files, warnings, err := openapi.Generate(doc, opts)
		if err != nil {
			return err
		}
		out := openapiGenerateOut
		if out == "" {
			out = "./requests"
		}
		paths, err := openapi.Write(out, files)
		if err != nil {
			return err
		}
		for _, warning := range warnings {
			fmt.Fprintln(cmd.ErrOrStderr(), "warning:", warning)
		}
		for _, path := range paths {
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", path)
		}
		if len(paths) == 0 {
			return fmt.Errorf("no operations matched; nothing written")
		}
		return nil
	},
}

var openapiValidateCmd = &cobra.Command{
	Use:   "validate <spec>",
	Short: "Validate an OpenAPI 3.x specification",
	Long: `Validate an OpenAPI 3.0 or 3.1 specification (JSON or YAML).

  reqly openapi validate api.yaml
  reqly openapi validate spec.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := openapi.LoadFile(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Validation passed cleanly.")
		return nil
	},
}

var openapiConvertV2Cmd = &cobra.Command{
	Use:   "convert-v2 <swagger2.json|yaml>",
	Short: "Convert a Swagger 2.0 / OpenAPI 2.0 spec to OpenAPI 3.0.3",
	Long: `Convert a legacy Swagger 2.0 specification file (JSON or YAML) into
a clean OpenAPI 3.0.3 YAML spec.

  reqly openapi convert-v2 swagger.json > openapi3.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read spec file: %w", err)
		}
		converted, err := openapi.ConvertSwagger2ToOpenAPI3(data)
		if err != nil {
			return fmt.Errorf("convert spec: %w", err)
		}
		fmt.Fprint(cmd.OutOrStdout(), string(converted))
		return nil
	},
}

func init() {
	openapiExploreCmd.Flags().StringArrayVar(&openapiExploreTags, "tag", nil, "filter by tag, repeatable")
	openapiExploreCmd.Flags().BoolVar(&openapiExploreJSON, "json", false, "print endpoints as JSON")

	openapiGenerateCmd.Flags().StringArrayVar(&openapiGenerateOps, "operation", nil, "operationId to generate, repeatable")
	openapiGenerateCmd.Flags().StringArrayVar(&openapiGenerateTags, "tag", nil, "generate every operation carrying this tag, repeatable")
	openapiGenerateCmd.Flags().StringVar(&openapiGenerateMethod, "method", "", "HTTP method for the method+path selector (use with --path)")
	openapiGenerateCmd.Flags().StringVar(&openapiGeneratePath, "path", "", "path for the method+path selector (use with --method)")
	openapiGenerateCmd.Flags().BoolVar(&openapiGenerateAll, "all", false, "generate every operation in the spec")
	openapiGenerateCmd.Flags().StringVar(&openapiGenerateOut, "output", "", "output directory (default ./requests)")

	openapiCmd.AddCommand(openapiExploreCmd, openapiGenerateCmd, openapiValidateCmd, openapiConvertV2Cmd)
	rootCmd.AddCommand(openapiCmd)
}
