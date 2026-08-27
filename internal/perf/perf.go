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

// Package perf implements lightweight load generation: constant-RPS worker
// pool, sorted latencies for P50/P95/P99, status histogram, error-rate.
package perf

import (
	"context"
	"sort"
	"sync"
	"time"
)

// Options configures a perf run.
type Options struct {
	RPS         int           // requests per second; 0 → single burst
	Duration    time.Duration // total run time; 0 → burst of RPS requests
	Concurrency int           // max concurrent sends; 0 → RPS
}

// Result is the summary after a perf run.
type Result struct {
	RPS          int         `json:"rps"`
	DurationMs   int64       `json:"durationMs"`
	Total        int         `json:"total"`
	LatenciesMs  []int64     `json:"latenciesMs,omitempty"`
	P50Ms        int64       `json:"p50Ms"`
	P95Ms        int64       `json:"p95Ms"`
	P99Ms        int64       `json:"p99Ms"`
	StatusCounts map[int]int `json:"statusCounts"`
	ErrorRate    float64     `json:"errorRate"`
}

// SendFunc is one request send: returns latency and status code.
type SendFunc func(ctx context.Context) (latency time.Duration, status int, err error)

// Run executes a perf load via ticker + concurrency semaphore.
func Run(ctx context.Context, opts Options, send SendFunc) (Result, error) {
	if opts.RPS <= 0 && opts.Duration == 0 {
		// Single request burst when no RPS/duration given.
		opts.RPS = 1
		opts.Duration = time.Second
	}
	if opts.RPS <= 0 {
		opts.RPS = 1
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = opts.RPS
	}
	if opts.Concurrency > opts.RPS*2 && opts.RPS > 0 {
		opts.Concurrency = opts.RPS * 2
	}

	total := opts.RPS
	if opts.Duration > 0 {
		secs := int(opts.Duration.Seconds())
		if secs < 1 {
			secs = 1
		}
		total = opts.RPS * secs
	}

	latencies := make([]int64, 0, total)
	statusCounts := make(map[int]int)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, opts.Concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(time.Second / time.Duration(opts.RPS))
	defer ticker.Stop()

	sent := 0
	for sent < total {
		select {
		case <-ctx.Done():
			goto done
		case <-ticker.C:
		}
		if sent >= total {
			break
		}
		sent++
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			lat, status, err := send(ctx)
			ms := lat.Milliseconds()
			mu.Lock()
			latencies = append(latencies, ms)
			if err != nil {
				statusCounts[0]++
			} else {
				statusCounts[status]++
			}
			mu.Unlock()
		}()
	}
done:
	wg.Wait()

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p50 := percentile(latencies, 50)
	p95 := percentile(latencies, 95)
	p99 := percentile(latencies, 99)

	nonSuccess := 0
	for code, c := range statusCounts {
		if code == 0 || code >= 400 {
			nonSuccess += c
		}
	}
	var rate float64
	if len(latencies) > 0 {
		rate = float64(nonSuccess) / float64(len(latencies))
	}

	return Result{
		RPS:          opts.RPS,
		DurationMs:   opts.Duration.Milliseconds(),
		Total:        len(latencies),
		LatenciesMs:  latencies,
		P50Ms:        p50,
		P95Ms:        p95,
		P99Ms:        p99,
		StatusCounts: statusCounts,
		ErrorRate:    rate,
	}, nil
}

func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	// Nearest rank percentile calculation
	k := (p*len(sorted) + 99) / 100
	if k < 1 {
		k = 1
	}
	if k > len(sorted) {
		k = len(sorted)
	}
	return sorted[k-1]
}
