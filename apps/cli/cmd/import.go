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
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/importer"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import external API artifacts",
	Long: `Import external API artifacts into Reqly-native project structures.

Supported sources: cURL, OpenAPI 3.x.`,
}

var importCurlCmd = &cobra.Command{
	Use:   "curl <command>",
	Short: "Import a cURL command",
	Long: `Convert a cURL command line into a request file.

  reqly import curl "curl -X POST https://api.example.com/users -H 'Content-Type: application/json' -d '{}'"

Use --output to write a request file instead of printing it.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, err := importer.ParseCurl(args[0])
		if err != nil {
			return err
		}
		if importOutput == "" {
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", req.Method, req.URL)
			for _, h := range req.Headers {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", h.Key, h.Value)
			}
			if req.Body != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", req.Body)
			}
			return nil
		}

		f := &requestfile.File{Name: "imported", Request: *req}
		data, err := yaml.Marshal(f)
		if err != nil {
			return err
		}
		if err := os.WriteFile(importOutput, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", importOutput)
		return nil
	},
}

var importOpenAPICmd = &cobra.Command{
	Use:   "openapi <file> [--output <dir>]",
	Short: "Import an OpenAPI 3.x document into a workspace",
	Long: `Parse an OpenAPI 3.x document (JSON or YAML) and write it as a
Git-native workspace: a reqly.yaml descriptor plus collections/ of request
files, grouped by tag.

  reqly import openapi petstore.yaml --output ./petstore

Without --output, the workspace is written into ./imported-openapi.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read spec: %w", err)
		}
		doc, err := importer.ParseOpenAPI(data)
		if err != nil {
			return err
		}
		result := doc.ToOpenAPIResult()

		out := importOutput
		if out == "" {
			out = "imported-openapi"
		}
		if err := result.Write(out); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "imported %s into %s (%d collections)\n",
			filepath.Base(args[0]), out, len(result.Collections))
		return nil
	},
}

var importOutput string

func init() {
	importCmd.AddCommand(importCurlCmd, importOpenAPICmd)
	importCurlCmd.Flags().StringVar(&importOutput, "output", "", "write a request file to this path")
	importOpenAPICmd.Flags().StringVar(&importOutput, "output", "", "directory to write the workspace into")
}
