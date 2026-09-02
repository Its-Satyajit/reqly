package monitor

import (
	"context"
	"testing"
	"time"
)

func TestThresholdStatus(t *testing.T) {
	called := false
	onResult := func(r Result) {
		called = true
		if r.OK {
			t.Fatalf("want fail on status mismatch")
		}
	}
	send := func(ctx context.Context) (time.Duration, int, error) { return 10 * time.Millisecond, 500, nil }
	if err := Run(context.Background(), 0, Threshold{Status: 200}, send, onResult); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Fatalf("onResult not called")
	}
}

func TestThresholdLatency(t *testing.T) {
	onResult := func(r Result) {
		if r.OK {
			t.Fatalf("want fail on latency")
		}
	}
	send := func(ctx context.Context) (time.Duration, int, error) { return 100 * time.Millisecond, 200, nil }
	if err := Run(context.Background(), 0, Threshold{Status: 200, LatencyMs: 50}, send, onResult); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Run(ctx, time.Second, Threshold{}, func(ctx context.Context) (time.Duration, int, error) { return 0, 200, nil }, func(Result) {})
	if err != context.Canceled {
		t.Fatalf("want canceled got %v", err)
	}
}
