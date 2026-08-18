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

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/testing"
)

var testCmd = &cobra.Command{
	Use:   "test <file.json|file.yaml>",
	Short: "Run assertions against a request",
	Long: `Execute the request defined in a test file and run its assertions.

A test file is JSON or YAML that couples a request definition (with optional
variables) to the assertions that run against its response:

  {
    "name": "users",
    "variables": { "token": "abc123" },
    "request": { "method": "GET", "url": "https://api.example.com/users" },
    "tests": [
      { "name": "ok", "assertions": [
        { "kind": "status", "expected": 200 },
        { "kind": "json", "path": "$.count", "exact": true, "value": "2" }
      ]}
    ]
  }

The equivalent YAML form is also accepted. No public API? Point tests at the
companion mock API (reqly-test-api) or a local test server, e.g.:

  {"request": {"method": "GET", "url": "https://reqly-test-api.vercel.app/api/users"}, ...}

Supported assertion kinds: status, header, body_contains, body_equals, json,
response_time. The command exits non-zero when any assertion fails.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tf, err := testing.LoadTestFile(args[0])
		if err != nil {
			return err
		}

		vars := tf.VariablesSet()
		masker, envSet, err := activeEnvironment(filepath.Dir(args[0]), tf.Environment)
		if err != nil {
			return err
		}
		mergeEnvScope(vars, envSet)
		masker.Add(auth.MaskValues(tf.Request.Auth.Type, tf.Request.Auth.Config, vars)...)

		client := newRequestClient(filepath.Dir(args[0]))
		resp, err := client.Execute(context.Background(), &tf.Request, vars)
		if err != nil {
			return fmt.Errorf("request failed: %s", masker.Mask(err.Error()))
		}
		maskAcquiredToken(masker, resp.AuthToken)

		results := tf.Suite().Run(resp)

		allPassed := true
		for _, tr := range results {
			mark := "\x1b[32m✓\x1b[0m"
			if !tr.Passed {
				mark = "\x1b[31m✗\x1b[0m"
				allPassed = false
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", mark, tr.Name)
			for _, r := range tr.Results {
				prefix := "  "
				if r.Passed {
					prefix += "\x1b[32m✓\x1b[0m "
				} else {
					prefix += "\x1b[31m✗\x1b[0m "
				}
				message := masker.Mask(r.Message)
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s\n", prefix, message)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\n%d/%d tests passed\n",
			countPassed(results), len(results))

		if !allPassed {
			return fmt.Errorf("test suite %q failed", tf.Name)
		}
		return nil
	},
}

func countPassed(results []testing.TestResult) int {
	n := 0
	for _, tr := range results {
		if tr.Passed {
			n++
		}
	}
	return n
}

func init() {
	testCmd.Flags().StringVar(&envFlag, "env", "", "environment to use (falls back to the file's environment field; REQLY_ENV wins)")
}
