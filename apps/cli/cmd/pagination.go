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
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/pagination"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
	"github.com/Its-Satyajit/reqly/internal/response"
)

var paginationMaxPages int

var paginationCmd = &cobra.Command{
	Use:   "pagination",
	Short: "Paginated request runners",
	Long:  `Iteratively execute a paginated request until a stop condition.`,
}

var paginationRunCmd = &cobra.Command{
	Use:   "run <request-file> [--max-pages <n>]",
	Short: "Run a paginated request",
	Long: `Iteratively execute a paginated request (page/offset/cursor/link-header) until a structural stop.

Pagination is declared on the request file via a "pagination" block:

  request:
    url: https://api.example.com/items
    pagination:
      strategy: page
      maxPages: 10

Strategies: page, offset, cursor (needs nextPath: $.nextCursor), link-header.

  reqly pagination run ./collections/items/list.yaml --max-pages 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		f, err := requestfile.LoadFile(path)
		if err != nil {
			return err
		}
		if f.Request.Pagination == nil {
			return fmt.Errorf("pagination: request %q has no pagination block", path)
		}
		baseDir := filepath.Dir(path)
		// env + vars like run
		vars := f.VariablesSet()
		masker, envSet, err := activeEnvironment(baseDir, f.Environment)
		if err != nil {
			return err
		}
		mergeEnvScope(vars, envSet)

		client := newRequestClient(baseDir)
		sendFn := func(ctx context.Context, r request.Request) (*response.Response, error) {
			// vars are shared but pagination mutates Query copy, not vars
			return client.Execute(ctx, &r, vars)
		}
		// masker not needed for pagination output? But we keep.

		opts := pagination.Options{}
		if paginationMaxPages > 0 {
			opts.MaxPages = paginationMaxPages
		}

		// OnStep prints progress
		onStep := func(s pagination.Step) {
			if s.Err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "step %d: error: %v\n", s.Index, masker.Mask(s.Err.Error()))
				return
			}
			if s.Response != nil {
				url := s.Request.URL
				// include query for visibility
				if len(s.Request.Query) > 0 {
					// reconstruct URL with query for display
					// keep simple: show request URL + query string
					q := ""
					for i, p := range s.Request.Query {
						if i > 0 {
							q += "&"
						}
						q += p.Key + "=" + p.Value
					}
					if q != "" {
						sep := "?"
						if len(url) > 0 && (url[len(url)-1] == '?' || contains(url, "?")) {
							sep = "&"
						}
						url = url + sep + q
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "step %d: %d %s (%s) %s\n", s.Index, s.Response.StatusCode, s.Response.StatusText, s.Response.Duration.Round(time.Millisecond), masker.Mask(url))
			}
		}

		ctx := context.Background()
		if err := pagination.Run(ctx, f.Request, opts, sendFn, onStep); err != nil {
			return err
		}
		return nil
	},
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (sub == "" || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func init() {
	paginationRunCmd.Flags().IntVar(&paginationMaxPages, "max-pages", 0, "maximum pages to fetch (overrides file maxPages, default 100)")
	paginationCmd.AddCommand(paginationRunCmd)
}
