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
	"go.yaml.in/yaml/v3"

	"github.com/Its-Satyajit/reqly/internal/importer"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import external API artifacts",
	Long: `Import external API artifacts into Reqly-native project structures.

Supported sources: cURL, OpenAPI 3.x, HAR.`,
}

// printImportReport renders a structured import report on stderr: entries
// grouped by category with item paths, then a severity tally line.
func printImportReport(cmd *cobra.Command, rep *importer.ImportReport) {
	if s := rep.String(); s != "" {
		fmt.Fprint(cmd.ErrOrStderr(), s)
	}
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
var importHarCollection string

var importHarCmd = &cobra.Command{
	Use:   "har <file> [--output <dir>] [--collection <name>]",
	Short: "Import a HAR file into a workspace",
	Long: `Parse a HAR 1.2 file (Chrome DevTools Network → Export HAR) and write it as a
Git-native workspace: reqly.yaml plus collections/<name>/<request>.yaml files.

  reqly import har capture.har
  reqly import har capture.har --output ./ws --collection chrome

Supports method/url/headers/cookies/queryString/postData (base64 decoded, mimeType→Content-Type);
drops pageref/timings/cache with warnings. Bodies >1MB spill to .reqly/blobs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read HAR: %w", err)
		}
		result, report, err := importer.ParseHAR(data)
		if err != nil {
			return err
		}
		printImportReport(cmd, report)
		out := importOutput
		if out == "" {
			out = "."
		}
		coll := importHarCollection
		if coll == "" {
			coll = "har-import"
		}
		if err := result.Write(out, coll); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "imported %s into %s (%d requests, collection %q)\n",
			filepath.Base(args[0]), out, len(result.Requests), coll)
		return nil
	},
}

var importPostmanCmd = &cobra.Command{
	Use:   "postman <file> [--output <dir>]",
	Short: "Import a Postman collection (v2.1) into a workspace",
	Long: `Parse a Postman v2.1 collection JSON and write it as a Git-native workspace:
reqly.yaml plus collections/<name>/ with nested folders preserved.

  reqly import postman my-api.postman_collection.json
  reqly import postman my-api.json --output ./ws

Imports requests, nested folders, variables, bodies (raw/urlencoded/form-data/graphql),
and basic/bearer/apikey auth. Scripts and unsupported features are reported as warnings.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read Postman collection: %w", err)
		}
		result, report, err := importer.ParsePostman(data)
		if err != nil {
			return err
		}
		printImportReport(cmd, report)
		out := importOutput
		if out == "" {
			out = importer.SanitizeDirName(result.Title)
		}
		if err := result.Write(out); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "imported %s into %s (%d requests)\n",
			filepath.Base(args[0]), out, result.RequestCount())
		return nil
	},
}

var importInsomniaCmd = &cobra.Command{
	Use:   "insomnia <file> [--output <dir>]",
	Short: "Import an Insomnia export (v4/v5) into a workspace",
	Long: `Parse an Insomnia export — v4 JSON (__export_format: 4) or v5 YAML
(collection.insomnia.rest/5.0) — and write it as a Git-native workspace.

  reqly import insomnia insomnia_export.json
  reqly import insomnia collection.yaml --output ./ws

Imports requests, nested folders, environments (as native environments/*.yaml),
and basic/bearer/apikey/digest auth. Cookie jars and unsupported features are
reported as warnings.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read Insomnia export: %w", err)
		}
		result, report, err := importer.ParseInsomnia(data)
		if err != nil {
			return err
		}
		printImportReport(cmd, report)
		out := importOutput
		if out == "" {
			out = importer.SanitizeDirName(result.Title)
		}
		if err := result.Write(out); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "imported %s into %s (%d requests, %d environments)\n",
			filepath.Base(args[0]), out, result.RequestCount(), len(result.Environments))
		return nil
	},
}

var importBrunoCmd = &cobra.Command{
	Use:   "bruno <file> [--output <dir>]",
	Short: "Import a Bruno collection export (JSON) into a workspace",
	Long: `Parse a Bruno collection export JSON and write it as a Git-native workspace.

  reqly import bruno collection.json
  reqly import bruno collection.json --output ./ws

Imports requests, nested folders, collection-level auth/headers, environments
(with secret variables routed to the secrets map), and basic/bearer/apikey/digest
auth. Scripts, assertions, and unsupported features are reported as warnings.
Directory imports are not supported — export the collection as a single JSON file.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := os.Stat(args[0])
		if err == nil && info.IsDir() {
			return fmt.Errorf("%s is a directory; export the Bruno collection as a single JSON file first", args[0])
		}
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read Bruno collection: %w", err)
		}
		result, report, err := importer.ParseBruno(data)
		if err != nil {
			return err
		}
		printImportReport(cmd, report)
		out := importOutput
		if out == "" {
			out = importer.SanitizeDirName(result.Title)
		}
		if err := result.Write(out); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "imported %s into %s (%d requests, %d environments)\n",
			filepath.Base(args[0]), out, result.RequestCount(), len(result.Environments))
		return nil
	},
}

var importWSDLCmd = &cobra.Command{
	Use:   "wsdl <file> [--output <dir>]",
	Short: "Import a WSDL 1.1 document into a SOAP workspace",
	Long: `Parse a WSDL 1.1 document and write it as a Git-native workspace with one
POST request per operation: a complete SOAP envelope skeleton (1.1 or 1.2
matched to the binding) sent to the port address with the binding's
SOAPAction header, and body children from the operation's inline XSD.

  reqly import wsdl service.wsdl
  reqly import wsdl service.wsdl --output ./soap-ws

External schemas (xsd:import/xsd:include) are resolved when the
schemaLocation is a local file relative to the WSDL (e.g. other.xsd);
remote URLs and rpc/encoded styles remain best-effort with warnings.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := os.Stat(args[0])
		if err == nil && info.IsDir() {
			return fmt.Errorf("%s is a directory; provide a single .wsdl or .xml file", args[0])
		}
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read WSDL document: %w", err)
		}
		baseDir := filepath.Dir(args[0])
		result, report, err := importer.ParseWSDLWithBase(data, baseDir)
		if err != nil {
			return err
		}
		printImportReport(cmd, report)
		out := importOutput
		if out == "" {
			out = importer.SanitizeDirName(result.Title)
		}
		if err := result.Write(out); err != nil {
			return err
		}
		count := 0
		for _, coll := range result.Collections {
			count += len(coll.Request)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "imported %s into %s (%d operations across %d services)\n",
			filepath.Base(args[0]), out, count, len(result.Collections))
		return nil
	},
}

var importFetchCmd = &cobra.Command{
	Use:   "fetch <snippet-or-file> [--output <file.yaml>]",
	Short: "Import a browser DevTools 'Copy as fetch' snippet",
	Long: `Convert a JavaScript fetch(...) snippet copied from Chrome, Firefox, or Safari
DevTools Network tab into a Reqly request.

  reqly import fetch 'fetch("https://api.example.com/data", { method: "POST", body: "{}" })'
  reqly import fetch snippet.js --output request.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		code := args[0]
		if data, err := os.ReadFile(args[0]); err == nil {
			code = string(data)
		}

		req, err := importer.ParseFetch(code)
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

func init() {
	importCmd.AddCommand(importCurlCmd, importFetchCmd, importOpenAPICmd, importHarCmd, importPostmanCmd, importInsomniaCmd, importBrunoCmd, importWSDLCmd)
	importCurlCmd.Flags().StringVar(&importOutput, "output", "", "write a request file to this path")
	importCurlCmd.Flags().StringVar(&importOutput, "out", "", "write a request file to this path")
	importFetchCmd.Flags().StringVar(&importOutput, "output", "", "write a request file to this path")
	importFetchCmd.Flags().StringVar(&importOutput, "out", "", "write a request file to this path")
	importOpenAPICmd.Flags().StringVar(&importOutput, "output", "", "directory to write the workspace into")
	importOpenAPICmd.Flags().StringVar(&importOutput, "out", "", "directory to write the workspace into")
	importHarCmd.Flags().StringVar(&importOutput, "output", "", "directory to write the workspace into")
	importHarCmd.Flags().StringVar(&importOutput, "out", "", "directory to write the workspace into")
	importHarCmd.Flags().StringVar(&importHarCollection, "collection", "har-import", "collection name for HAR entries")
	importPostmanCmd.Flags().StringVar(&importOutput, "output", "", "directory to write the workspace into")
	importPostmanCmd.Flags().StringVar(&importOutput, "out", "", "directory to write the workspace into")
	importInsomniaCmd.Flags().StringVar(&importOutput, "output", "", "directory to write the workspace into")
	importInsomniaCmd.Flags().StringVar(&importOutput, "out", "", "directory to write the workspace into")
	importBrunoCmd.Flags().StringVar(&importOutput, "output", "", "directory to write the workspace into")
	importBrunoCmd.Flags().StringVar(&importOutput, "out", "", "directory to write the workspace into")
	importWSDLCmd.Flags().StringVar(&importOutput, "output", "", "directory to write the workspace into")
	importWSDLCmd.Flags().StringVar(&importOutput, "out", "", "directory to write the workspace into")
}
