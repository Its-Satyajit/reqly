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
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/request"
)

var collectionWorkspace string

var collectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Work with collections",
	Long: `Inspect and run collections defined on disk.

A workspace is a directory of Git-native project files: a reqly.yaml descriptor,
a collections/ tree of collection and folder descriptors, and plain request
files. See the ROADMAP "Workspaces, collections & storage" milestone.`,
}

var collectionRunCmd = &cobra.Command{
	Use:   "run <collection/request>",
	Short: "Run a request inside a collection",
	Long: `Run a single request from a workspace, resolving its inherited
configuration (base URL, headers, auth, variables) down the
Workspace → Collection → Folder → Request chain.

The argument is the request's path within the workspace, collection
prefix optional when it is unambiguous:

  reqly collection run users/list-users.yaml
  reqly collection run users/auth/login.yaml

Use --workspace to point at a workspace directory other than the current one.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, coll, chain, entry, err := findCollectionRequest(args[0])
		if err != nil {
			return err
		}

		resolved, err := ws.ResolveRequest(coll, chain, entry)
		if err != nil {
			return err
		}

		client := request.NewClient()
		resp, err := client.Execute(context.Background(), &resolved.Request, resolved.Vars)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		status := resp.StatusCode
		color := ""
		if !resp.OK() {
			color = "\x1b[31m" // red
		} else {
			color = "\x1b[32m" // green
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s%d %s\x1b[0m (%s)\n",
			color, status, resp.StatusText, resp.Duration.Round(time.Millisecond))
		fmt.Fprintf(cmd.OutOrStdout(), "%s %d %s\n", resp.Proto, status, resp.StatusText)

		for key, values := range resp.Headers {
			for _, value := range values {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", key, value)
			}
		}

		fmt.Fprintln(cmd.OutOrStdout())
		if len(resp.Body) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), string(resp.Body))
		}
		return nil
	},
}

var collectionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List collections and requests in a workspace",
	Long: `Print the workspace tree: collections, folders, requests, and the
resolved base URL for each container.

Use --workspace to point at a workspace directory other than the current one.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := collections.LoadWorkspace(collectionWorkspace)
		if err != nil {
			return err
		}

		name := ws.Config.Name
		if name == "" {
			name = filepath.Base(ws.Root)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s/\n", name)
		if ws.Config.BaseURL != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  baseURL: %s\n", ws.Config.BaseURL)
		}

		var colls []string
		for _, c := range ws.Collections {
			colls = append(colls, c.Name)
		}
		sort.Strings(colls)

		for _, name := range colls {
			c := findCollection(ws, name)
			fmt.Fprintf(cmd.OutOrStdout(), "  %s/\n", c.Name)
			printContainerRequests(cmd, c.Requests, 4)
			printFolders(cmd, c.Folders, 4)
		}
		return nil
	},
}

// printFolders recursively prints folder trees at the given indent level.
func printFolders(cmd *cobra.Command, folders []*collections.Folder, indent int) {
	names := make([]string, 0, len(folders))
	for _, f := range folders {
		names = append(names, f.Name)
	}
	sort.Strings(names)

	for _, name := range names {
		var f *collections.Folder
		for _, cand := range folders {
			if cand.Name == name {
				f = cand
				break
			}
		}
		pad := fmt.Sprintf("%*s", indent, "")
		fmt.Fprintf(cmd.OutOrStdout(), "%s%s/\n", pad, f.Name)
		printContainerRequests(cmd, f.Requests, indent+2)
		printFolders(cmd, f.Folders, indent+2)
	}
}

// printContainerRequests prints request file names at the given indent level.
func printContainerRequests(cmd *cobra.Command, reqs []*collections.RequestEntry, indent int) {
	names := make([]string, 0, len(reqs))
	for _, r := range reqs {
		names = append(names, r.Name)
	}
	sort.Strings(names)

	pad := fmt.Sprintf("%*s", indent, "")
	for _, name := range names {
		fmt.Fprintf(cmd.OutOrStdout(), "%s%s\n", pad, name)
	}
}

// findCollection returns the named collection from a workspace.
func findCollection(ws *collections.Workspace, name string) *collections.Collection {
	for _, c := range ws.Collections {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// findCollectionRequest loads the workspace and resolves a request path such
// as "users/list-users.yaml".
func findCollectionRequest(path string) (*collections.Workspace, *collections.Collection, []*collections.Folder, *collections.RequestEntry, error) {
	ws, err := collections.LoadWorkspace(collectionWorkspace)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	coll, chain, entry, ok := ws.FindRequest(collections.RequestPath(path))
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("request %q not found in workspace %s", path, collectionWorkspace)
	}
	return ws, coll, chain, entry, nil
}

func init() {
	collectionCmd.AddCommand(collectionRunCmd, collectionListCmd)

	pwd, err := filepath.Abs(".")
	if err != nil {
		pwd = "."
	}
	collectionRunCmd.Flags().StringVar(&collectionWorkspace, "workspace", pwd, "workspace directory")
	collectionListCmd.Flags().StringVar(&collectionWorkspace, "workspace", pwd, "workspace directory")
}
