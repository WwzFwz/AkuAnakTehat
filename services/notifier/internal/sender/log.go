package sender

import (
	"context"
	"example.com/akuanaktehat/notifier/internal/contract"
	"log/slog"
)

// Log simulates delivery. No external notification provider is contacted.
type Log struct{ Logger *slog.Logger }

func (s *Log) Send(ctx context.Context, e contract.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.Logger.Info("notification_simulated", "event_id", e.EventID, "hazard_id", e.HazardID, "version", e.Version, "severity", e.Hazard.Severity, "correlation_id", e.CorrelationID)
	return nil
}
