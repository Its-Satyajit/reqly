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
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/runner"
)

var collectionWorkspace string
var collectionFailFast bool

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

		masker, envSet, err := activeEnvironment(ws.Root, entry.File.Environment)
		if err != nil {
			return err
		}
		mergeEnvScope(resolved.Vars, envSet)
		masker.Add(auth.MaskValues(resolved.Request.Auth.Type, resolved.Request.Auth.Config, resolved.Vars)...)

		client := newRequestClient(ws.Root)
		resp, err := client.Execute(context.Background(), &resolved.Request, resolved.Vars)
		if err != nil {
			return fmt.Errorf("request failed: %s", masker.Mask(err.Error()))
		}
		maskAcquiredToken(masker, resp.AuthToken)

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
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", key, masker.Mask(value))
			}
		}

		fmt.Fprintln(cmd.OutOrStdout())
		if len(resp.Body) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), masker.Mask(string(resp.Body)))
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

// collectionTestCmd runs every request in a collection, evaluating pre/post
// scripts and reqly.test() assertions, with a summary report.
var collectionTestCmd = &cobra.Command{
	Use:   "test <collection>",
	Short: "Run every request in a collection with scripts and tests",
	Long: `Run every request in a collection in order, executing pre-request and
post-request scripts and evaluating reqly.test() assertions.

Variables set by a post-request script (reqly.setVariable) are available to
later requests, so a login step can feed a token into the next request.

  reqly collection test users
  reqly collection test users --fail-fast

Use --workspace to point at a workspace directory other than the current one.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := collections.LoadWorkspace(collectionWorkspace)
		if err != nil {
			return err
		}

		coll := findCollection(ws, args[0])
		if coll == nil {
			return fmt.Errorf("collection %q not found in workspace %s", args[0], collectionWorkspace)
		}

		// Collection runs use a single environment for the whole run and ignore
		// per-file environment: fields.
		masker, envSet, err := activeEnvironment(ws.Root, "")
		if err != nil {
			return err
		}
		report, err := runner.RunCollection(context.Background(), ws, coll, envSet, runner.Options{
			FailFast: collectionFailFast,
			Client:   newRequestClient(ws.Root),
		})
		if err != nil {
			return err
		}
		for _, step := range report.Steps {
			masker.Add(step.AuthValues()...)
		}

		for _, step := range report.Steps {
			status := "\x1b[32mPASS\x1b[0m"
			if !step.Passed {
				status = "\x1b[31mFAIL\x1b[0m"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", status, step.Name)
			if step.RequestError != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", masker.Mask(step.RequestError.Error()))
			}
			if step.Response != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d %s (%s)\n",
					step.Response.StatusCode, step.Response.StatusText,
					step.Response.Duration.Round(time.Millisecond))
			}
			for _, tr := range step.Tests {
				mark := "ok"
				if !tr.Passed {
					mark = "FAIL"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", mark, tr.Name)
			}
			if len(step.Logs) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  script output: %s\n", masker.Mask(strings.Join(step.Logs, "; ")))
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\n%d passed, %d failed (%d total, %s)\n",
			report.Passed, report.Failed, report.Total, report.Duration.Round(time.Millisecond))
		if !report.OK() {
			return fmt.Errorf("collection %q: %d step(s) failed", args[0], report.Failed)
		}
		return nil
	},
}

func init() {
	collectionCmd.AddCommand(collectionRunCmd, collectionListCmd, collectionTestCmd)

	pwd, err := filepath.Abs(".")
	if err != nil {
		pwd = "."
	}
	collectionRunCmd.Flags().StringVar(&collectionWorkspace, "workspace", pwd, "workspace directory")
	collectionRunCmd.Flags().StringVar(&envFlag, "env", "", "environment to use (falls back to the collection's environment field; REQLY_ENV wins)")
	collectionListCmd.Flags().StringVar(&collectionWorkspace, "workspace", pwd, "workspace directory")
	collectionTestCmd.Flags().StringVar(&collectionWorkspace, "workspace", pwd, "workspace directory")
	collectionTestCmd.Flags().BoolVar(&collectionFailFast, "fail-fast", false, "stop after the first failing step")
	collectionTestCmd.Flags().StringVar(&envFlag, "env", "", "environment to use for the whole collection run (REQLY_ENV wins)")
}
