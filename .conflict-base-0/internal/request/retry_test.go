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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func flakyServer(t *testing.T, failures int32, status int) *httptest.Server {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) <= failures {
			w.WriteHeader(status)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func alwaysStatusServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestExecute_NoRetryConfig_AttemptsIsOne(t *testing.T) {
	srv := flakyServer(t, 2, http.StatusServiceUnavailable)

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 passthrough, got %d", resp.StatusCode)
	}
	if resp.Attempts != 1 {
		t.Fatalf("expected 1 attempt with no retry config, got %d", resp.Attempts)
	}
}

func TestExecute_RetriesUntilSuccess(t *testing.T) {
	srv := flakyServer(t, 2, http.StatusBadGateway)

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 3, DelayMs: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected eventual 200, got %d", resp.StatusCode)
	}
	if resp.Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", resp.Attempts)
	}
}

func TestExecute_RetryExhaustedReturnsLastResponse(t *testing.T) {
	srv := alwaysStatusServer(t, http.StatusServiceUnavailable)

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 2, DelayMs: 1},
	}, nil)
	if err != nil {
		t.Fatalf("exhausted retries must return last response, not error: %v", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
	if resp.Attempts != 3 {
		t.Fatalf("expected count+1=3 attempts, got %d", resp.Attempts)
	}
}

func TestExecute_ZeroCountMeansOff(t *testing.T) {
	srv := alwaysStatusServer(t, http.StatusServiceUnavailable)

	client := NewClient()
	resp, _ := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{},
	}, nil)
	if resp.Attempts != 1 {
		t.Fatalf("expected zero-count retry to be off, got %d attempts", resp.Attempts)
	}
}

func TestExecute_FixedStrategyWaitsConstantDelay(t *testing.T) {
	srv := flakyServer(t, 2, http.StatusServiceUnavailable)

	client := NewClient()
	start := time.Now()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 2, Strategy: RetryStrategyFixed, DelayMs: 30},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if resp.Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", resp.Attempts)
	}
	// Two fixed waits of 30ms each.
	if elapsed < 60*time.Millisecond {
		t.Fatalf("expected >=60ms total backoff, got %v", elapsed)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("fixed strategy waited too long: %v", elapsed)
	}
}

func TestExecute_ExponentialBackoffCapped(t *testing.T) {
	srv := flakyServer(t, 3, http.StatusServiceUnavailable)

	client := NewClient()
	start := time.Now()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry: &Retry{
			Count:      3,
			Strategy:   RetryStrategyExponential,
			DelayMs:    10,
			MaxDelayMs: 25,
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if resp.Attempts != 4 {
		t.Fatalf("expected 4 attempts, got %d", resp.Attempts)
	}
	// Waits: 10 + 20 + 25(capped from 40) = 55ms.
	if elapsed < 55*time.Millisecond {
		t.Fatalf("expected >=55ms total capped exponential backoff, got %v", elapsed)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("capped exponential waited too long: %v", elapsed)
	}
}

func TestExecute_DefaultStrategyIsExponential(t *testing.T) {
	policy := (&Retry{}).normalized()
	if policy.Strategy != RetryStrategyExponential {
		t.Fatalf("expected default strategy exponential, got %q", policy.Strategy)
	}
	if policy.DelayMs != 1000 {
		t.Fatalf("expected default delay 1000ms, got %d", policy.DelayMs)
	}
	if policy.MaxDelayMs != 30000 {
		t.Fatalf("expected default cap 30000ms, got %d", policy.MaxDelayMs)
	}
	if len(policy.RetryOn) != 4 {
		t.Fatalf("expected default retryOn 429/502/503/504, got %v", policy.RetryOn)
	}
}

func TestExecute_Bare500NotRetriedByDefault(t *testing.T) {
	srv := alwaysStatusServer(t, http.StatusInternalServerError)

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 5, DelayMs: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Attempts != 1 {
		t.Fatalf("expected bare 500 to pass through untried, got %d attempts", resp.Attempts)
	}
}

func TestExecute_RetryOnOverrideIncludes500(t *testing.T) {
	srv := flakyServer(t, 1, http.StatusInternalServerError)

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 3, DelayMs: 1, RetryOn: []int{500}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || resp.Attempts != 2 {
		t.Fatalf("expected 500 retried to success in 2 attempts, got %d/%d", resp.StatusCode, resp.Attempts)
	}
}

func TestExecute_NetworkErrorRetried(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	baseURL := srv.URL
	srv.Close() // Nothing listens there anymore.

	client := NewClient()
	_, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    baseURL,
		Retry:  &Retry{Count: 2, DelayMs: 1},
	}, nil)
	if err == nil {
		t.Fatal("expected network error after exhausting retries")
	}
}

func TestExecute_RetryAfterHeaderOverridesBackoff(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	start := time.Now()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 3, Strategy: RetryStrategyFixed, DelayMs: 5000},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Retry-After: 0 wins over the 5s fixed delay.
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Retry-After should override computed backoff, waited %v", elapsed)
	}
	if resp.Attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", resp.Attempts)
	}
}

func TestExecute_RetryAfterClampedToMaxDelay(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "999999")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	start := time.Now()
	if _, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 1, DelayMs: 1, MaxDelayMs: 25},
	}, nil); err != nil {
		t.Fatal(err)
	}
	// The huge Retry-After must be clamped to the 25ms cap.
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Retry-After should clamp to maxDelayMs, got %v", elapsed)
	}
}

func TestExecute_RetryAfterSecondsRespected(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient()
	start := time.Now()
	if _, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 3, DelayMs: 1},
	}, nil); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed < 900*time.Millisecond {
		t.Fatalf("expected ~1s Retry-After wait, got %v", elapsed)
	}
}

func TestExecute_ContextCancelDuringBackoffAborts(t *testing.T) {
	srv := alwaysStatusServer(t, http.StatusServiceUnavailable)

	ctx, cancel := context.WithCancel(context.Background())
	client := NewClient()
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := client.Execute(ctx, &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 5, Strategy: RetryStrategyFixed, DelayMs: 5000},
	}, nil)
	if err == nil {
		t.Fatal("expected context error from cancelled backoff")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("cancel during backoff should abort promptly, took %v", elapsed)
	}
}

func TestExecute_OnRetryObserverNotified(t *testing.T) {
	srv := flakyServer(t, 2, http.StatusTooManyRequests)

	var events []string
	client := NewClient(WithOnRetry(func(e RetryEvent) {
		events = append(events, fmt.Sprintf("%d/%d:%d:%d", e.Attempt, e.TotalAttempts, e.Delay.Milliseconds(), e.StatusCode))
	}))
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodGet,
		URL:    srv.URL,
		Retry:  &Retry{Count: 3, DelayMs: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", resp.Attempts)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 retry events, got %v", events)
	}
	if events[0] != "1/4:1:429" || events[1] != "2/4:2:429" {
		t.Fatalf("unexpected retry events %v", events)
	}
}

func TestExecute_PostMethodRetried(t *testing.T) {
	srv := flakyServer(t, 1, http.StatusBadGateway)

	client := NewClient()
	resp, err := client.Execute(context.Background(), &Request{
		Method: MethodPost,
		URL:    srv.URL,
		Body:   `{"k":"v"}`,
		Retry:  &Retry{Count: 2, DelayMs: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Attempts != 2 {
		t.Fatalf("expected POST retried (2 attempts), got %d", resp.Attempts)
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		header  string
		wantMin time.Duration
		wantMax time.Duration
	}{
		{"seconds", "2", 1900 * time.Millisecond, 2100 * time.Millisecond},
		{"zero", "0", 0, 10 * time.Millisecond},
		{"negative ignored", "-5", -1, -1},
		{"garbage ignored", "soon", -1, -1},
		{"past date clamped to zero", now.UTC().Format(http.TimeFormat), 0, 10 * time.Millisecond},
		{"future date", now.UTC().Add(2 * time.Second).Format(http.TimeFormat), 900 * time.Millisecond, 2100 * time.Millisecond},
		{"empty ignored", "", -1, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseRetryAfter(tc.header, now)
			if tc.wantMin < 0 {
				if ok {
					t.Fatalf("expected !ok for header %q, got %v", tc.header, got)
				}
				return
			}
			if !ok {
				t.Fatalf("expected usable delay for header %q", tc.header)
			}
			if got < tc.wantMin || got > tc.wantMax {
				t.Fatalf("expected between %v and %v, got %v", tc.wantMin, tc.wantMax, got)
			}
		})
	}
}

func TestRetryableStatus(t *testing.T) {
	policy := (&Retry{}).normalized()
	for _, code := range []int{429, 502, 503, 504} {
		if !policy.retryable(code) {
			t.Errorf("expected %d retryable", code)
		}
	}
	for _, code := range []int{200, 400, 401, 404, 500, 501} {
		if policy.retryable(code) {
			t.Errorf("expected %d NOT retryable", code)
		}
	}
	if !(&Retry{RetryOn: []int{500}}).normalized().retryable(500) {
		t.Error("override retryOn [500] should make 500 retryable")
	}
}

func TestBackoffDelaySequence(t *testing.T) {
	policy := (&Retry{DelayMs: 100, MaxDelayMs: 400}).normalized()
	want := []time.Duration{100, 200, 400, 400}
	for attempt, exp := range want {
		got := policy.backoff(attempt + 1)
		if got != exp*time.Millisecond {
			t.Errorf("attempt %d: expected %v, got %v", attempt+1, exp*time.Millisecond, got)
		}
	}
	if v := strconv.FormatInt(int64(policy.DelayMs), 10); v != "100" {
		t.Errorf("normalization mutated source: %s", v)
	}
}
