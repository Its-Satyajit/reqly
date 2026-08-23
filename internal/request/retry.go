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

package request

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Its-Satyajit/reqly/internal/response"
)

// Retry strategies for the delay between attempts.
const (
	RetryStrategyFixed       = "fixed"
	RetryStrategyExponential = "exponential"
)

const (
	defaultRetryDelayMs    = int64(1000)
	defaultRetryMaxDelayMs = int64(30000)
)

// defaultRetryOn is the status set retried when the user does not override
// it: rate limiting plus gateway-class failures. Bare 500s are excluded —
// they are usually deterministic server bugs, not transient blips.
var defaultRetryOn = []int{
	http.StatusTooManyRequests,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
}

// RetryEvent describes one automatic retry about to happen after a failed
// attempt. Attempt is the 1-based number of the attempt that just failed,
// TotalAttempts is Count+1, Delay is how long the engine waits before
// resending, and exactly one of Err / StatusCode classifies the failure.
type RetryEvent struct {
	Attempt       int
	TotalAttempts int
	Delay         time.Duration
	Err           error
	StatusCode    int
}

// normalized returns a copy of the policy with defaults filled in. The
// receiver is never mutated so request structs stay shareable.
func (p *Retry) normalized() *Retry {
	out := &Retry{
		Count:      p.Count,
		DelayMs:    p.DelayMs,
		MaxDelayMs: p.MaxDelayMs,
		RetryOn:    append([]int(nil), p.RetryOn...),
	}
	switch p.Strategy {
	case RetryStrategyFixed:
		out.Strategy = RetryStrategyFixed
	default:
		out.Strategy = RetryStrategyExponential
	}
	if out.DelayMs <= 0 {
		out.DelayMs = defaultRetryDelayMs
	}
	if out.MaxDelayMs <= 0 {
		out.MaxDelayMs = defaultRetryMaxDelayMs
	}
	if out.RetryOn == nil {
		out.RetryOn = append([]int(nil), defaultRetryOn...)
	}
	return out
}

// retryable reports whether a response status should trigger another attempt.
func (p *Retry) retryable(status int) bool {
	for _, code := range p.RetryOn {
		if code == status {
			return true
		}
	}
	return false
}

// backoff returns the wait following a failed attempt. Attempt 1 waits the
// base delay; exponential doubles each subsequent attempt. Both strategies
// clamp to MaxDelayMs.
func (p *Retry) backoff(attempt int) time.Duration {
	max := time.Duration(p.MaxDelayMs) * time.Millisecond
	delay := time.Duration(p.DelayMs) * time.Millisecond
	if p.Strategy == RetryStrategyExponential {
		for range attempt - 1 {
			delay *= 2
			if delay >= max {
				break
			}
		}
	}
	return clampDelay(delay, max)
}

func clampDelay(delay, max time.Duration) time.Duration {
	if delay > max || delay < 0 {
		return max
	}
	return delay
}

// parseRetryAfter interprets a Retry-After header value as a delay relative
// to now. Seconds and HTTP-date forms are supported; ok is false when the
// value is absent, negative, or unparseable.
func parseRetryAfter(value string, now time.Time) (delay time.Duration, ok bool) {
	if value == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(value); err == nil {
		if secs < 0 {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}
	if when, err := http.ParseTime(value); err == nil {
		if d := when.Sub(now); d > 0 {
			return d, true
		}
		return 0, true
	}
	return 0, false
}

// delayFor computes the wait following a failed attempt number. A
// Retry-After header on 429/503 overrides the computed backoff, clamped to
// MaxDelayMs so a hostile or broken server cannot stall the client.
func (p *Retry) delayFor(resp *response.Response, attempt int, now time.Time) time.Duration {
	max := time.Duration(p.MaxDelayMs) * time.Millisecond
	if resp != nil &&
		(resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable) &&
		resp.Headers != nil {
		if values := resp.Headers["Retry-After"]; len(values) > 0 {
			if delay, ok := parseRetryAfter(values[0], now); ok {
				return clampDelay(delay, max)
			}
		}
	}
	return p.backoff(attempt)
}

// isContextErr reports whether the error stems from caller cancellation —
// those are never retried because the caller wants out, not the server.
func isContextErr(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
