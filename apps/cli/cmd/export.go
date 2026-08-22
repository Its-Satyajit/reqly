package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/exporter"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export workspaces and requests to shareable formats",
	Long: `Export Reqly-native projects into shareable formats.

Supported targets: Postman collection v2.1, code snippets (cURL, JS, Python, Go).`,
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

func init() {
	exportCmd.AddCommand(exportPostmanCmd, exportCodeCmd)
	exportPostmanCmd.Flags().StringVar(&exportOutput, "output", "", "write the collection to this file")
	exportCodeCmd.Flags().StringVar(&exportCodeLang, "lang", "", "target language (cURL, js, python, go)")
	exportCodeCmd.Flags().StringVar(&exportCodeOut, "out", "", "write snippet to this file")
	exportCodeCmd.Flags().StringVar(&exportCodeEnv, "env", "", "environment to resolve variables")
}
