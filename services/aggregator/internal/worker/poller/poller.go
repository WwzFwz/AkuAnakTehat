package poller

import (
	"context"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"example.com/akuanaktehat/aggregator/internal/application/ingest"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"example.com/akuanaktehat/aggregator/internal/observability"
	"log/slog"
	"math/rand/v2"
	"sort"
	"time"
)

type Fetch func(context.Context, time.Time, string) ([]canonicalize.Item, error)
type Endpoint struct {
	Name    string
	Fetch   Fetch
	Breaker Breaker
}
type Applier interface {
	ApplyBatch(context.Context, canonicalize.Batch) (ingest.Stats, error)
}
type Worker struct {
	Endpoints         []*Endpoint
	Checkpoints       ingest.CheckpointReader
	Status            ingest.StatusStore
	Ingest            Applier
	Interval, Overlap time.Duration
	Logger            *slog.Logger
}
type result struct {
	batch   canonicalize.Batch
	err     error
	code    string
	skipped bool
	start   time.Time
}

func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		started := time.Now()
		w.Cycle(ctx)
		delay := w.Interval - time.Since(started)
		if delay < 0 {
			delay = 0
		}
		delay += time.Duration(rand.Int64N(max(1, int64(w.Interval/10))))
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

// Fetch BMKG endpoints concurrently, then apply sequentially. Each source has
// its own Worker and no worker starts another cycle until this one completes.
func (w *Worker) Cycle(ctx context.Context) {
	pending := make([]chan result, len(w.Endpoints))
	for i, endpoint := range w.Endpoints {
		pending[i] = make(chan result, 1)
		go func(e *Endpoint, out chan<- result) {
			r := result{start: time.Now().UTC()}
			r.batch.Endpoint = e.Name
			if !e.Breaker.Allow(r.start) {
				r.skipped = true
				r.code = "circuit_open"
				out <- r
				return
			}
			corr, err := hazard.UUID()
			if err != nil {
				r.err = err
				r.code = "correlation_unavailable"
				out <- r
				return
			}
			r.batch.CorrelationID = corr
			stamp, exists, err := w.Checkpoints.ReadCheckpoint(observability.WithID(ctx, corr), e.Name)
			if err != nil {
				r.err = err
				r.code = "checkpoint_unavailable"
				out <- r
				return
			}
			if exists {
				stamp = stamp.Add(-w.Overlap)
			}
			// Request-start watermark catches warning updates without requiring a new wire field.
			r.batch.Watermark = time.Now().UTC()
			r.batch.Items, r.err = e.Fetch(ctx, stamp, corr)
			if r.err != nil {
				r.code = "source_fetch_failed"
				e.Breaker.Failure(time.Now())
			} else {
				e.Breaker.Success()
			}
			out <- r
		}(endpoint, pending[i])
	}
	for _, out := range pending {
		r := <-out
		if ctx.Err() != nil {
			continue
		}
		if r.skipped {
			continue
		} // retain last attempt/failure; a skipped poll is not an HTTP attempt
		stats := ingest.Stats{}
		if r.err == nil {
			unknown := map[string]bool{}
			for _, item := range r.batch.Items {
				for _, key := range item.Unknown {
					unknown[key] = true
				}
			}
			if len(unknown) > 0 {
				keys := []string{}
				for k := range unknown {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				w.Logger.Info("schema_drift", "endpoint", r.batch.Endpoint, "fields", keys, "correlation_id", r.batch.CorrelationID)
			}
			stats, r.err = w.Ingest.ApplyBatch(observability.WithID(ctx, r.batch.CorrelationID), r.batch)
			if r.err != nil {
				r.code = "ingest_failed"
			} else if stats.Rejected > 0 {
				r.code = "records_quarantined"
			}
		}
		if err := w.Status.RecordPoll(observability.WithID(ctx, r.batch.CorrelationID), r.batch.Endpoint, r.err == nil, stats.Rejected > 0, r.code, time.Now().UTC()); err != nil {
			w.Logger.Error("source_status_failed", "endpoint", r.batch.Endpoint)
		}
		w.Logger.Info("poll_complete", "endpoint", r.batch.Endpoint, "ok", r.err == nil, "error_code", r.code, "changed", stats.Changed, "unchanged", stats.Unchanged, "rejected", stats.Rejected, "latency_ms", time.Since(r.start).Milliseconds(), "correlation_id", r.batch.CorrelationID)
	}
}
