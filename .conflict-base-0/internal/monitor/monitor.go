// Package monitor implements scheduled health checks (status + latency threshold).
package monitor

import (
	"context"
	"time"
)

// Threshold defines health success criteria.
type Threshold struct {
	Status    int   `json:"status"`
	LatencyMs int64 `json:"latencyMs"`
}

// Result is one tick's health check.
type Result struct {
	OK        bool      `json:"ok"`
	Status    int       `json:"status"`
	LatencyMs int64     `json:"latencyMs"`
	Threshold Threshold `json:"threshold"`
	At        time.Time `json:"at"`
}

// SendFunc sends one request; returns latency and status.
type SendFunc func(ctx context.Context) (time.Duration, int, error)

// Run ticks every interval until context cancel, calling send and checking threshold.
// Interval 0 → single check. Returns on context cancel.
func Run(ctx context.Context, interval time.Duration, threshold Threshold, send SendFunc, onResult func(Result)) error {
	check := func() error {
		lat, status, err := send(ctx)
		if err != nil {
			onResult(Result{OK: false, Status: status, LatencyMs: lat.Milliseconds(), Threshold: threshold, At: time.Now()})
			return nil
		}
		ok := true
		if threshold.Status != 0 && status != threshold.Status {
			ok = false
		}
		if threshold.LatencyMs != 0 && lat.Milliseconds() > threshold.LatencyMs {
			ok = false
		}
		onResult(Result{OK: ok, Status: status, LatencyMs: lat.Milliseconds(), Threshold: threshold, At: time.Now()})
		return nil
	}
	if interval <= 0 {
		return check()
	}
	if err := check(); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := check(); err != nil {
				return err
			}
		}
	}
}
