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
