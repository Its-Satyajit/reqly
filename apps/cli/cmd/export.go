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

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/exporter"
	"github.com/Its-Satyajit/reqly/internal/request"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export workspaces and requests to shareable formats",
	Long: `Export Reqly-native projects into shareable formats.

Supported targets: Postman collection v2.1.`,
}

var exportPostmanCmd = &cobra.Command{
	Use:   "postman <workspace-dir> [--output <file>]",
	Short: "Export a workspace as a Postman collection",
	Long: `Convert a Reqly workspace (see "reqly collection") into a Postman
Collection v2.1 JSON document. Inherited base URL and headers are applied to
each request.

  reqly export postman . --output collection.json

Without --output the JSON is printed to stdout.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := collections.LoadWorkspace(args[0])
		if err != nil {
			return err
		}

		name := ws.Config.Name
		if name == "" {
			name = "Reqly workspace"
		}

		requests, err := flattenWorkspace(ws)
		if err != nil {
			return err
		}

		data, err := exporter.ExportToPostmanJSON(name, requests)
		if err != nil {
			return err
		}

		if exportOutput == "" {
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}
		if err := os.WriteFile(exportOutput, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s (%d requests)\n", exportOutput, len(requests))
		return nil
	},
}

// flattenWorkspace resolves every request in the workspace and returns them in
// a flat list, with inherited base URL and headers applied.
func flattenWorkspace(ws *collections.Workspace) ([]request.Request, error) {
	var requests []request.Request

	var walkFolders func(coll *collections.Collection, chain []*collections.Folder, folders []*collections.Folder)
	walkFolders = func(coll *collections.Collection, chain []*collections.Folder, folders []*collections.Folder) {
		for _, f := range folders {
			childChain := append(chain, f)
			for _, entry := range f.Requests {
				if resolved, err := ws.ResolveRequest(coll, childChain, entry); err == nil {
					if resolved.Request.Name == "" {
						resolved.Request.Name = entry.Name
					}
					requests = append(requests, resolved.Request)
				}
			}
			walkFolders(coll, childChain, f.Folders)
		}
	}

	for _, coll := range ws.Collections {
		for _, entry := range coll.Requests {
			if resolved, err := ws.ResolveRequest(coll, nil, entry); err == nil {
				if resolved.Request.Name == "" {
					resolved.Request.Name = entry.Name
				}
				requests = append(requests, resolved.Request)
			}
		}
		walkFolders(coll, nil, coll.Folders)
	}
	return requests, nil
}

var exportOutput string

func init() {
	exportCmd.AddCommand(exportPostmanCmd)
	exportPostmanCmd.Flags().StringVar(&exportOutput, "output", "", "write the collection to this file")
}
