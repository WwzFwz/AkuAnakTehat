package outbox

import (
	"context"
	"log/slog"
	"time"
)

type Relay struct {
	Store               Store
	Publisher           Publisher
	Interval, Retention time.Duration
	BatchSize           int
	Logger              *slog.Logger
}

// Drain stops at the first failed publish/mark. Never skip a pending version.
// No database transaction is held while waiting for a broker ACK.
func (r *Relay) Drain(ctx context.Context) error {
	messages, err := r.Store.Pending(ctx, r.BatchSize)
	if err != nil {
		return err
	}
	for _, m := range messages {
		started := time.Now()
		if err = r.Publisher.Publish(ctx, m); err != nil {
			return err
		}
		publishLatency := time.Since(started).Milliseconds()
		if err = r.Store.MarkPublished(ctx, m.ID, time.Now().UTC()); err != nil {
			return err
		}
		r.Logger.Info("outbox_published", "event_id", m.EventID, "hazard_id", m.HazardID, "version", m.Version, "correlation_id", m.CorrelationID, "publish_latency_ms", publishLatency)
	}
	return nil
}
func (r *Relay) Run(ctx context.Context) {
	timer := time.NewTicker(r.Interval)
	defer timer.Stop()
	cleanup := time.Now()
	for ctx.Err() == nil {
		if err := r.Drain(ctx); err != nil && ctx.Err() == nil {
			r.Logger.Warn("outbox_retry", "reason", "publish_or_store_unavailable")
		}
		if time.Since(cleanup) >= time.Hour {
			if _, err := r.Store.DeletePublishedBefore(ctx, time.Now().Add(-r.Retention)); err != nil {
				r.Logger.Warn("outbox_cleanup_failed")
			}
			cleanup = time.Now()
		}
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}
