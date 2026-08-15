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
