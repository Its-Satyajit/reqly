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
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

var (
	runMethod  string
	runHeaders []string
	runBody    string
	runTimeout time.Duration
)

var runCmd = &cobra.Command{
	Use:   "run <url|file>",
	Short: "Execute a single HTTP request",
	Long: `Execute a single HTTP request and print the status line, headers, and body.

The argument is either a URL or a plain-text request file (JSON or YAML):

  reqly run https://api.example.com/users
  reqly run request.yaml

No public API? Point Reqly at the companion mock API (reqly-test-api, hosted
on Vercel) or run a local test server:

  reqly run https://reqly-test-api.vercel.app/api/users
  reqly run http://localhost:3123/api/status/404

A request file couples a request definition with its variables:

  name: users
  variables:
    token: abc123
  request:
    method: GET
    url: https://api.example.com/users
    headers:
      - key: Authorization
        value: Bearer {{token}}

When a file is used, flags override the file's fields. Use --method, --header,
and --data to build requests directly on the CLI:

  reqly run -m POST -H 'Content-Type: application/json' -d '{"name":"reqly"}' https://api.example.com/users`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		var req *request.Request
		var vars *variables.Set
		baseDir := "."
		var fileEnv string

		if requestfile.LooksLikeFile(target) {
			f, err := requestfile.LoadFile(target)
			if err != nil {
				return err
			}
			req = &f.Request
			baseDir = filepath.Dir(target)
			vars = f.VariablesSet()
			fileEnv = f.Environment
			if err := applyRunOverrides(cmd, req); err != nil {
				return err
			}
		} else {
			headers, err := parseHeaders(runHeaders)
			if err != nil {
				return err
			}
			req = &request.Request{
				Method:  request.Method(strings.ToUpper(runMethod)),
				URL:     target,
				Headers: headers,
				Body:    runBody,
				Timeout: runTimeout.Milliseconds(),
			}
			vars = variables.NewSet()
		}

		masker, envSet, err := activeEnvironment(baseDir, fileEnv)
		if err != nil {
			return err
		}
		mergeEnvScope(vars, envSet)
		masker.Add(auth.MaskValues(req.Auth.Type, req.Auth.Config, vars)...)

		client := request.NewClient()
		resp, err := client.Execute(context.Background(), req, vars)
		if err != nil {
			return fmt.Errorf("request failed: %s", masker.Mask(err.Error()))
		}
		// Mask the acquired OAuth token (and any other runtime credential)
		// so headers/body echoing it never leak it.
		if resp.AuthToken != "" {
			masker.Add(resp.AuthToken)
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

func init() {
	runCmd.Flags().StringVarP(&runMethod, "method", "m", "GET", "HTTP method to use")
	runCmd.Flags().StringArrayVarP(&runHeaders, "header", "H", nil, "request header in 'Key: Value' form (repeatable)")
	runCmd.Flags().StringVarP(&runBody, "data", "d", "", "request body")
	runCmd.Flags().DurationVarP(&runTimeout, "timeout", "t", 30*time.Second, "request timeout")
	runCmd.Flags().StringVar(&envFlag, "env", "", "environment to use (falls back to the file's environment field; REQLY_ENV wins)")
}

// applyRunOverrides copies explicitly-set CLI flags onto a request loaded from
// a file. Flags that were not changed keep the file's values.
func applyRunOverrides(cmd *cobra.Command, req *request.Request) error {
	flags := cmd.Flags()
	if flags.Changed("method") {
		req.Method = request.Method(strings.ToUpper(runMethod))
	}
	if flags.Changed("header") {
		headers, err := parseHeaders(runHeaders)
		if err != nil {
			return err
		}
		req.Headers = headers
	}
	if flags.Changed("data") {
		req.Body = runBody
	}
	if flags.Changed("timeout") {
		req.Timeout = runTimeout.Milliseconds()
	}
	return nil
}

// parseHeaders converts CLI "Key: Value" strings into request.Header values.
func parseHeaders(raw []string) ([]request.Header, error) {
	var headers []request.Header
	for _, item := range raw {
		key, value, ok := strings.Cut(item, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header %q: expected 'Key: Value'", item)
		}
		headers = append(headers, request.Header{
			Key:   strings.TrimSpace(key),
			Value: strings.TrimSpace(value),
		})
	}
	return headers, nil
}
