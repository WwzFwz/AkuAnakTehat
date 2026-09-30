package poller

import (
	"context"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"example.com/akuanaktehat/aggregator/internal/application/ingest"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"
)

type checkpoints struct{ stamp time.Time }

func (c checkpoints) ReadCheckpoint(context.Context, string) (time.Time, bool, error) {
	return c.stamp, !c.stamp.IsZero(), nil
}

type capture struct {
	batches []canonicalize.Batch
	health  map[string]bool
}

func (c *capture) ApplyBatch(_ context.Context, b canonicalize.Batch) (ingest.Stats, error) {
	c.batches = append(c.batches, b)
	return ingest.Stats{}, nil
}
func (c *capture) RecordPoll(_ context.Context, e string, ok, degraded bool, code string, at time.Time) error {
	c.health[e] = ok
	return nil
}
func TestPartialBMKGAndBreakerRecovery(t *testing.T) {
	start := time.Now()
	b := Breaker{Threshold: 2, Cooldown: time.Second}
	b.Failure(start)
	if !b.Allow(start) {
		t.Fatal("opened too early")
	}
	b.Failure(start)
	if b.Allow(start) {
		t.Fatal("did not open")
	}
	if !b.Allow(start.Add(time.Second)) {
		t.Fatal("probe not allowed")
	}
	b.Failure(start.Add(time.Second))
	if b.Allow(start.Add(time.Second)) {
		t.Fatal("failed probe did not reopen")
	}
	b.Success()
	if !b.Allow(start) {
		t.Fatal("success did not close")
	}
	cap := &capture{health: map[string]bool{}}
	var mu sync.Mutex
	calls := 0
	stamp := time.Now().Add(-time.Minute).UTC()
	fetch := func(ctx context.Context, since time.Time, corr string) ([]canonicalize.Item, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		if !since.Equal(stamp.Add(-10*time.Second)) || corr == "" {
			t.Error("checkpoint overlap/correlation missing")
		}
		return nil, nil
	}
	w := Worker{Endpoints: []*Endpoint{{Name: canonicalize.SeismicEndpoint, Fetch: fetch, Breaker: Breaker{Threshold: 1, Cooldown: time.Hour}}, {Name: canonicalize.WarningEndpoint, Fetch: func(context.Context, time.Time, string) ([]canonicalize.Item, error) {
		return nil, errors.New("failed")
	}, Breaker: Breaker{Threshold: 1, Cooldown: time.Hour}}}, Checkpoints: checkpoints{stamp}, Status: cap, Ingest: cap, Overlap: 10 * time.Second, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	w.Cycle(context.Background())
	w.Cycle(context.Background())
	if len(cap.batches) != 2 || calls != 2 || !cap.health[canonicalize.SeismicEndpoint] || cap.health[canonicalize.WarningEndpoint] {
		t.Fatal("failed endpoint blocked valid endpoint or advanced its checkpoint")
	}
	for _, batch := range cap.batches {
		if batch.Endpoint != canonicalize.SeismicEndpoint || batch.Watermark.Before(start) {
			t.Fatal("bad successful checkpoint")
		}
	}
}
