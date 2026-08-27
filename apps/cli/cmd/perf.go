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
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/perf"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
	"github.com/Its-Satyajit/reqly/internal/response"
)

var (
	perfRPS         int
	perfDuration    time.Duration
	perfConcurrency int
	perfJSON        bool
)

var perfCmd = &cobra.Command{
	Use:   "perf",
	Short: "Performance testing (lightweight)",
	Long:  `Lightweight load generation with P50/P95/P99 and status histogram.`,
}

var perfRunCmd = &cobra.Command{
	Use:   "run <request-file> [--rps 10] [--duration 30s] [--concurrency 5] [--json]",
	Short: "Run a perf load against a request file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		f, err := requestfile.LoadFile(path)
		if err != nil {
			return err
		}
		baseDir := filepath.Dir(path)
		svc := core.NewRunService(findWorkspaceRoot(baseDir))
		defer svc.Close()

		fileVars := f.VariablesSet()
		noRecord := false
		send := func(ctx context.Context) (time.Duration, int, error) {
			req := f.Request
			res, err := svc.Run(ctx, req, core.RunRequestOptions{
				EnvFlag:       envFlag,
				FileEnv:       f.Environment,
				FileVars:      fileVars,
				RecordHistory: &noRecord,
			})
			if err != nil {
				return 0, 0, err
			}
			// Use response.Duration for latency; fallback to 0.
			var resp *response.Response = res.Response
			return resp.Duration, resp.StatusCode, nil
		}

		opts := perf.Options{
			RPS:         perfRPS,
			Duration:    perfDuration,
			Concurrency: perfConcurrency,
		}
		if opts.RPS == 0 {
			opts.RPS = 10
		}
		if opts.Duration == 0 {
			opts.Duration = 30 * time.Second
		}

		// Normalize request for perf (strip pagination etc. not needed).
		_ = request.Request{}

		result, err := perf.Run(cmd.Context(), opts, send)
		if err != nil {
			return err
		}
		if perfJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "rps=%d duration=%s total=%d p50=%dms p95=%dms p99=%dms errorRate=%.2f%%\n", result.RPS, perfDuration, result.Total, result.P50Ms, result.P95Ms, result.P99Ms, result.ErrorRate*100)
		for code, count := range result.StatusCounts {
			fmt.Fprintf(cmd.OutOrStdout(), "  %d: %d\n", code, count)
		}
		return nil
	},
}

func init() {
	perfRunCmd.Flags().IntVar(&perfRPS, "rps", 10, "requests per second")
	perfRunCmd.Flags().DurationVar(&perfDuration, "duration", 30*time.Second, "total duration")
	perfRunCmd.Flags().IntVar(&perfConcurrency, "concurrency", 0, "max concurrent (default RPS)")
	perfRunCmd.Flags().BoolVar(&perfJSON, "json", false, "machine-readable JSON output")
	perfCmd.AddCommand(perfRunCmd)
}
