package perf

import (
	"context"
	"testing"
	"time"
)

func TestPerfP50P95(t *testing.T) {
	// Sorted latencies 10,20,30,40,50 → P50=30, P95=50, P99=50
	send := func(ctx context.Context) (time.Duration, int, error) { return 10 * time.Millisecond, 200, nil }
	_ = send
	lat := []int64{10, 20, 30, 40, 50}
	if got := percentile(lat, 50); got != 30 {
		t.Fatalf("p50 %d want 30", got)
	}
	if got := percentile(lat, 95); got != 50 {
		t.Fatalf("p95 %d want 50", got)
	}
	if got := percentile([]int64{}, 50); got != 0 {
		t.Fatalf("empty p50 want 0 got %d", got)
	}
}

func TestPerfRun(t *testing.T) {
	opts := Options{RPS: 5, Duration: time.Second, Concurrency: 2}
	i := 0
	send := func(ctx context.Context) (time.Duration, int, error) {
		i++
		if i%5 == 0 {
			return 5 * time.Millisecond, 500, nil
		}
		return 10 * time.Millisecond, 200, nil
	}
	res, err := Run(context.Background(), opts, send)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Total != 5 {
		t.Fatalf("total %d want 5", res.Total)
	}
	if res.StatusCounts[200] != 4 {
		t.Fatalf("200 count %d", res.StatusCounts[200])
	}
	if res.StatusCounts[500] != 1 {
		t.Fatalf("500 count %d", res.StatusCounts[500])
	}
	if res.ErrorRate == 0 {
		t.Fatalf("errorRate want >0")
	}
}
