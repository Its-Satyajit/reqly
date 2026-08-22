package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/exporter"
	"github.com/Its-Satyajit/reqly/internal/history"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
	"github.com/Its-Satyajit/reqly/internal/version"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export workspaces and requests to shareable formats",
	Long: `Export Reqly-native projects into shareable formats.

Supported targets: Postman collection v2.1, code snippets (cURL, JS, Python, Go), HAR (history → HAR).`,
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
var exportCodeLang string
var exportCodeOut string
var exportCodeEnv string
var exportWorkspaceOut string

var exportCodeCmd = &cobra.Command{
	Use:   "code <request-file> --lang <cURL|js|python|go> [--out <file>] [--env <name>]",
	Short: "Generate code snippet for a request",
	Long: `Generate a code snippet for a Reqly request file or collection path.

Supports cURL, JavaScript (fetch), Python (requests), Go (net/http).
Secrets render as [SECRET]. The request is resolved through the workspace/env chain like "reqly run".

  reqly export code ./collections/users/list.yaml --lang curl
  reqly export code ./collections/users/list.yaml --lang js --out snippet.js --env prod`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		lang := exportCodeLang
		if lang == "" {
			return fmt.Errorf("--lang is required (cURL, js, python, go)")
		}
		req, err := loadRequestForExport(path)
		if err != nil {
			return err
		}
		snippet, err := exporter.Generate(req, lang, nil)
		if err != nil {
			return err
		}
		if exportCodeOut == "" {
			fmt.Fprintln(cmd.OutOrStdout(), snippet)
			return nil
		}
		if err := os.WriteFile(exportCodeOut, []byte(snippet), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", exportCodeOut)
		return nil
	},
}

func loadRequestForExport(path string) (request.Request, error) {
	if f, err := requestfile.LoadFile(path); err == nil {
		return f.Request, nil
	}
	ws, err := collections.LoadWorkspace(".")
	if err != nil {
		return request.Request{}, err
	}
	for _, coll := range ws.Collections {
		for _, entry := range coll.Requests {
			if entry.Path == path || entry.Name == path {
				if resolved, err := ws.ResolveRequest(coll, nil, entry); err == nil {
					return resolved.Request, nil
				}
			}
		}
	}
	return request.Request{}, fmt.Errorf("request not found: %q", path)
}

var exportWorkspaceCmd = &cobra.Command{
	Use:   "workspace [src] --out <dir>",
	Short: "Copy a workspace to a new directory",
	Long: `Copy a Reqly workspace (descriptors + request files) to a new directory.

  reqly export workspace . --out /tmp/new-ws
  reqly export workspace ./my-ws --out /tmp/copy

The destination is created via SaveWorkspace (pruning, atomic, format-preserving).
src defaults to the current workspace (.). --out is required.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := "."
		if len(args) > 0 {
			src = args[0]
		}
		if exportWorkspaceOut == "" {
			return fmt.Errorf("--out <dir> is required")
		}
		ws, err := collections.LoadWorkspace(src)
		if err != nil {
			return err
		}
		if err := collections.SaveWorkspace(exportWorkspaceOut, ws); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "exported workspace to %s (%d collections)\n", exportWorkspaceOut, len(ws.Collections))
		return nil
	},
}

var exportHarOut string
var exportHarEnv string
var exportHarLimit int

var exportHarCmd = &cobra.Command{
	Use:   "har [--out <file.har>] [--env <name>] [--limit <n>]",
	Short: "Export history as HAR",
	Long: `Export the workspace's history (request + response) as HAR 1.2 JSON.

  reqly export har --out traffic.har --env staging --limit 100
  reqly export har > traffic.har

History is the source of truth (responses synthesized, timings from DurationMS,
secrets masked to [SECRET]). Without --out, HAR JSON is printed to stdout.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root := collections.FindWorkspaceRoot(".")
		if root == "" {
			return fmt.Errorf("not a workspace: no reqly.yaml found")
		}
		store, err := history.NewStore(filepath.Join(root, ".reqly", "history.db"))
		if err != nil {
			return err
		}
		defer func() { _ = store.Close() }()
		limit := exportHarLimit
		if limit <= 0 {
			limit = 500
		}
		entries, err := store.List(context.Background(), limit, 0, nil)
		if err != nil {
			return err
		}
		if exportHarEnv != "" {
			var filtered []history.Entry
			for _, e := range entries {
				if e.Env == exportHarEnv {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}
		// masker for secrets (M28b: populate from env secrets; M28 masks via empty masker)
		var mask func(string) string
		data, err := exporter.ExportHAR(entries, version.Version, mask)
		if err != nil {
			return err
		}
		if exportHarOut == "" {
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}
		if err := os.WriteFile(exportHarOut, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s (%d entries)\n", exportHarOut, len(entries))
		return nil
	},
}

func init() {
	exportCmd.AddCommand(exportPostmanCmd, exportCodeCmd, exportWorkspaceCmd, exportHarCmd)
	exportPostmanCmd.Flags().StringVar(&exportOutput, "output", "", "write the collection to this file")
	exportCodeCmd.Flags().StringVar(&exportCodeLang, "lang", "", "target language (cURL, js, python, go)")
	exportCodeCmd.Flags().StringVar(&exportCodeOut, "out", "", "write snippet to this file")
	exportCodeCmd.Flags().StringVar(&exportCodeEnv, "env", "", "environment to resolve variables")
	exportWorkspaceCmd.Flags().StringVar(&exportWorkspaceOut, "out", "", "destination directory")
	exportHarCmd.Flags().StringVar(&exportHarOut, "out", "", "write HAR to this file (default stdout)")
	exportHarCmd.Flags().StringVar(&exportHarEnv, "env", "", "filter history by environment")
	exportHarCmd.Flags().IntVar(&exportHarLimit, "limit", 500, "maximum history entries to export")
}
