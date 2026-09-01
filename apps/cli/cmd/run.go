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
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

var (
	runMethod      string
	runHeaders     []string
	runBody        string
	runTimeout     time.Duration
	runRetries     int
	runRetryDelay  time.Duration
	runProxy       string
	runInsecure    bool
	runCAFile      string
	runHTTP2       bool
	runNoKeepAlive bool
	runTimeline    bool
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

  reqly run -m POST -H 'Content-Type: application/json' -d '{"name":"reqly"}' https://api.example.com/users

Variables use {{name}} with 6 scopes (global → process-env → environment → collection → folder → request).
Strict mode: an undefined {{var}} aborts the send with "undefined variable \"name\"" — it is not sent as a literal.
This is fail-closed (unlike Postman/Insomnia leniency) so missing secrets are not silently sent as "{{...}}".`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		var req *request.Request
		var vars *variables.Set // still built for file mode; Run re-resolves scopes internally
		_ = vars
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
				Retry:   retryFromFlags(cmd),
				Proxy:   runProxy,
				TLS:     tlsFromFlags(cmd),
			}
			vars = variables.NewSet()
		}

		// One deep pipeline call: env precedence, variable layering, token
		// caching, masking, retries, and history recording live in core (ADR 0025).
		root := findWorkspaceRoot(baseDir)
		svc := core.NewRunService(root)
		defer svc.Close()

		var onRetry func(request.RetryEvent)
		if req.Retry != nil && req.Retry.Count > 0 {
			total := req.Retry.Count + 1
			out := cmd.OutOrStdout()
			onRetry = func(e request.RetryEvent) {
				reason := strconv.Itoa(e.StatusCode)
				if e.Err != nil {
					reason = e.Err.Error()
				}
				fmt.Fprintf(out, "retrying in %s (%s, attempt %d/%d)\n",
					e.Delay.Round(time.Millisecond), reason, e.Attempt, total)
			}
		}

		requestPath := ""
		if requestfile.LooksLikeFile(target) {
			requestPath = target
		}
		res, err := svc.Run(context.Background(), *req, core.RunRequestOptions{
			EnvFlag:     envFlag,
			FileEnv:     fileEnv,
			FileVars:    vars,
			RequestPath: requestPath,
			OnRetry:     onRetry,
		})
		if err != nil {
			return fmt.Errorf("request failed: %s", err)
		}
		resp := res.Response

		status := resp.StatusCode
		color := ""
		if !resp.OK() {
			color = "\x1b[31m" // red
		} else {
			color = "\x1b[32m" // green
		}
		if resp.Attempts > 1 {
			fmt.Fprintf(cmd.OutOrStdout(), "%s%d %s\x1b[0m (%s, %d attempts)\n",
				color, status, resp.StatusText, resp.Duration.Round(time.Millisecond), resp.Attempts)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s%d %s\x1b[0m (%s)\n",
				color, status, resp.StatusText, resp.Duration.Round(time.Millisecond))
		}
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
		if runTimeline && resp.Timings != nil {
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "TIMELINE WATERFALL:")
			fmt.Fprintf(cmd.OutOrStdout(), "  DNS Lookup:        %4dms\n", resp.Timings.DNS)
			fmt.Fprintf(cmd.OutOrStdout(), "  TCP Connect:       %4dms\n", resp.Timings.Connect)
			fmt.Fprintf(cmd.OutOrStdout(), "  TLS Handshake:     %4dms\n", resp.Timings.TLS)
			fmt.Fprintf(cmd.OutOrStdout(), "  Server Processing: %4dms (TTFB)\n", resp.Timings.Server)
			fmt.Fprintf(cmd.OutOrStdout(), "  Content Transfer:  %4dms\n", resp.Timings.Transfer)
		}
		return nil
	},
}

func init() {
	runCmd.Flags().StringVarP(&runMethod, "method", "m", "GET", "HTTP method to use")
	runCmd.Flags().StringArrayVarP(&runHeaders, "header", "H", nil, "request header in 'Key: Value' form (repeatable)")
	runCmd.Flags().StringVarP(&runBody, "data", "d", "", "request body")
	runCmd.Flags().DurationVarP(&runTimeout, "timeout", "t", 30*time.Second, "request timeout")
	runCmd.Flags().IntVar(&runRetries, "retries", 0, "automatic retries after transient failures (network errors, 429/502/503/504)")
	runCmd.Flags().DurationVar(&runRetryDelay, "retry-delay", time.Second, "base delay between retries (exponential backoff by default)")
	runCmd.Flags().StringVar(&envFlag, "env", "", "environment to use (falls back to the file's environment field; REQLY_ENV wins)")
	runCmd.Flags().StringVar(&runProxy, "proxy", "", "proxy URL for this request (overrides environment proxy)")
	runCmd.Flags().BoolVar(&runInsecure, "insecure", false, "skip TLS verification for this request")
	runCmd.Flags().StringVar(&runCAFile, "ca-file", "", "path to PEM CA bundle for this request")
	runCmd.Flags().BoolVar(&runHTTP2, "http2", false, "force HTTP/2 ALPN protocol for this request")
	runCmd.Flags().BoolVar(&runNoKeepAlive, "no-keepalive", false, "disable HTTP connection pooling for this request")
	runCmd.Flags().BoolVar(&runTimeline, "timeline", false, "display ASCII network timing waterfall")
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
	if flags.Changed("retries") || flags.Changed("retry-delay") {
		policy := req.Retry
		if policy == nil {
			policy = &request.Retry{}
		} else {
			copied := *policy
			policy = &copied
		}
		if flags.Changed("retries") {
			policy.Count = runRetries
		}
		if flags.Changed("retry-delay") {
			policy.DelayMs = runRetryDelay.Milliseconds()
		}
		req.Retry = policy
	}
	if flags.Changed("proxy") {
		req.Proxy = runProxy
	}
	if flags.Changed("http2") && runHTTP2 {
		req.HTTPVersion = "http2"
	}
	if flags.Changed("no-keepalive") && runNoKeepAlive {
		req.DisableKeepAlives = true
	}
	if flags.Changed("insecure") || flags.Changed("ca-file") {
		tlsCfg := req.TLS
		if tlsCfg == nil {
			tlsCfg = &request.TLSConfig{}
		} else {
			copied := *tlsCfg
			tlsCfg = &copied
		}
		if flags.Changed("insecure") {
			tlsCfg.InsecureSkipVerify = runInsecure
		}
		if flags.Changed("ca-file") {
			tlsCfg.CAFile = runCAFile
		}
		req.TLS = tlsCfg
	}
	return nil
}

// tlsFromFlags builds a TLSConfig purely from CLI flags for URL-mode requests.
func tlsFromFlags(cmd *cobra.Command) *request.TLSConfig {
	flags := cmd.Flags()
	if !flags.Changed("insecure") && !flags.Changed("ca-file") {
		return nil
	}
	return &request.TLSConfig{
		InsecureSkipVerify: runInsecure,
		CAFile:             runCAFile,
	}
}

// retryFromFlags builds a Retry policy purely from CLI flags for URL-mode
// requests, returning nil when neither flag was set.
func retryFromFlags(cmd *cobra.Command) *request.Retry {
	flags := cmd.Flags()
	if !flags.Changed("retries") && !flags.Changed("retry-delay") {
		return nil
	}
	return &request.Retry{
		Count:   runRetries,
		DelayMs: runRetryDelay.Milliseconds(),
	}
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
