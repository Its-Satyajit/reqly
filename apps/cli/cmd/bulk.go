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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/bulk"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
	"github.com/Its-Satyajit/reqly/internal/response"
)

var (
	bulkDataPath       string
	bulkParallel       bool
	bulkConcurrency    int
	bulkContinueOnError bool
)

var bulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Bulk request execution",
	Long:  `Execute one request repeatedly against many input rows (CSV/JSON).`,
}

var bulkRunCmd = &cobra.Command{
	Use:   "run <request-file> --data <csv|json> [--parallel] [--concurrency <n>] [--continue-on-error]",
	Short: "Run a request against many rows",
	Long: `Run a request file repeatedly for each row in a CSV or JSON data file.

CSV: header row maps to {{var}} (e.g. "id,name" → {{id}}, {{name}})
JSON: array of objects (values stringified)

  reqly bulk run ./collections/users/create.yaml --data users.csv
  reqly bulk run ./collections/users/create.yaml --data users.json --parallel --concurrency 10
  reqly bulk run ./req.yaml --data data.csv --continue-on-error`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if bulkDataPath == "" {
			return fmt.Errorf("bulk: --data <csv|json> is required")
		}
		path := args[0]
		f, err := requestfile.LoadFile(path)
		if err != nil {
			return err
		}
		baseDir := filepath.Dir(path)
		vars := f.VariablesSet()
		masker, envSet, err := activeEnvironment(baseDir, f.Environment)
		if err != nil {
			return err
		}
		mergeEnvScope(vars, envSet)

		rows, err := parseBulkData(bulkDataPath)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "warning: no rows in data file")
			return nil
		}

		client := newRequestClient(baseDir)
		sendFn := func(ctx context.Context, r request.Request) (*response.Response, error) {
			return client.Execute(ctx, &r, vars)
		}

		opts := bulk.Options{
			Parallel:        bulkParallel,
			Concurrency:     bulkConcurrency,
			ContinueOnError: bulkContinueOnError,
		}
		if opts.Parallel && opts.Concurrency <= 0 {
			opts.Concurrency = 5
		}

		onStep := func(s bulk.Step) {
			if s.Err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "step %d: error: %v\n", s.Index, masker.Mask(s.Err.Error()))
				return
			}
			if s.Response != nil {
				url := s.Request.URL
				if len(s.Request.Query) > 0 {
					q := ""
					for i, p := range s.Request.Query {
						if i > 0 {
							q += "&"
						}
						q += p.Key + "=" + p.Value
					}
					if q != "" {
						sep := "?"
						if strings.Contains(url, "?") {
							sep = "&"
						}
						url = url + sep + q
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "step %d: %d %s (%s) %s\n", s.Index, s.Response.StatusCode, s.Response.StatusText, s.Response.Duration.Round(time.Millisecond), masker.Mask(url))
			}
		}

		if err := bulk.Run(context.Background(), f.Request, rows, opts, sendFn, onStep); err != nil {
			return err
		}
		return nil
	},
}

func parseBulkData(path string) ([]map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read bulk data %q: %w", path, err)
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		var arr []map[string]any
		if err := json.Unmarshal(data, &arr); err != nil {
			return nil, fmt.Errorf("parse bulk JSON %q: %w", path, err)
		}
		rows := make([]map[string]string, 0, len(arr))
		for _, m := range arr {
			row := make(map[string]string, len(m))
			for k, v := range m {
				switch x := v.(type) {
				case string:
					row[k] = x
				case float64:
					if x == float64(int64(x)) {
						row[k] = fmt.Sprintf("%d", int64(x))
					} else {
						row[k] = fmt.Sprintf("%v", x)
					}
				case bool:
					row[k] = fmt.Sprintf("%t", x)
				case nil:
					row[k] = ""
				default:
					b, _ := json.Marshal(x)
					row[k] = string(b)
				}
			}
			rows = append(rows, row)
		}
		return rows, nil
	}
	// CSV default
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse bulk CSV %q: %w", path, err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	header := records[0]
	for i, h := range header {
		header[i] = strings.TrimSpace(h)
	}
	var rows []map[string]string
	for _, rec := range records[1:] {
		row := make(map[string]string, len(header))
		for i, h := range header {
			val := ""
			if i < len(rec) {
				val = rec[i]
			}
			row[h] = val
		}
		// skip empty rows
		empty := true
		for _, v := range row {
			if strings.TrimSpace(v) != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func init() {
	bulkRunCmd.Flags().StringVar(&bulkDataPath, "data", "", "CSV or JSON file with rows (required)")
	_ = bulkRunCmd.MarkFlagRequired("data")
	bulkRunCmd.Flags().BoolVar(&bulkParallel, "parallel", false, "run rows in parallel")
	bulkRunCmd.Flags().IntVar(&bulkConcurrency, "concurrency", 5, "concurrency when --parallel (default 5)")
	bulkRunCmd.Flags().BoolVar(&bulkContinueOnError, "continue-on-error", false, "continue on non-2xx/error instead of stopping")
	bulkCmd.AddCommand(bulkRunCmd)
}
