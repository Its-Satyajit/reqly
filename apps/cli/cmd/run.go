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
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

var (
	runMethod  string
	runHeaders []string
	runBody    string
	runTimeout time.Duration
)

var runCmd = &cobra.Command{
	Use:   "run <url>",
	Short: "Execute a single HTTP request",
	Long: `Execute a single HTTP request and print the status line, headers, and body.

By default a GET request is sent. Use --method, --header, and --data to build
richer requests:

  reqly run https://api.example.com/users
  reqly run -m POST -H 'Content-Type: application/json' -d '{"name":"reqly"}' https://api.example.com/users`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		vars := variables.NewSet()
		headers, err := parseHeaders(runHeaders)
		if err != nil {
			return err
		}

		req := &request.Request{
			Method:  request.Method(strings.ToUpper(runMethod)),
			URL:     url,
			Headers: headers,
			Body:    runBody,
			Timeout: runTimeout.Milliseconds(),
		}

		client := request.NewClient()
		resp, err := client.Execute(context.Background(), req, vars)
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

func init() {
	runCmd.Flags().StringVarP(&runMethod, "method", "m", "GET", "HTTP method to use")
	runCmd.Flags().StringArrayVarP(&runHeaders, "header", "H", nil, "request header in 'Key: Value' form (repeatable)")
	runCmd.Flags().StringVarP(&runBody, "data", "d", "", "request body")
	runCmd.Flags().DurationVarP(&runTimeout, "timeout", "t", 30*time.Second, "request timeout")
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
