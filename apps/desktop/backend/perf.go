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

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/perf"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// PerfRunResult is the bridge result for the desktop perf dashboard.
type PerfRunResult struct {
	Result perf.Result `json:"result"`
}

// PerfRun executes a perf load for the request file at specPath (workspace-relative).
func (s *AppService) PerfRun(specPath string, rps int, durationMs int64, concurrency int) (*PerfRunResult, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to run perf")
	}
	abs, err := s.resolveTestPath(specPath)
	if err != nil {
		return nil, err
	}
	f, err := requestfile.LoadFile(abs)
	if err != nil {
		return nil, err
	}
	svc := core.NewRunService(s.root)
	defer svc.Close()

	fileVars := f.VariablesSet()
	noRecord := false
	send := func(ctx context.Context) (time.Duration, int, error) {
		res, err := svc.Run(ctx, f.Request, core.RunRequestOptions{
			EnvFlag:       "",
			FileEnv:       f.Environment,
			FileVars:      fileVars,
			RecordHistory: &noRecord,
		})
		if err != nil {
			return 0, 0, err
		}
		return res.Response.Duration, res.Response.StatusCode, nil
	}
	opts := perf.Options{
		RPS:         rps,
		Duration:    time.Duration(durationMs) * time.Millisecond,
		Concurrency: concurrency,
	}
	ctx, cancel := context.WithTimeout(context.Background(), opts.Duration+5*time.Second)
	defer cancel()
	res, err := perf.Run(ctx, opts, send)
	if err != nil {
		return nil, fmt.Errorf("perf run: %w", err)
	}
	return &PerfRunResult{Result: res}, nil
}
