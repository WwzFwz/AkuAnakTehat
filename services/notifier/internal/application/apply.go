package application

import (
	"context"
	"example.com/akuanaktehat/notifier/internal/contract"
)

type Dedup interface {
	Seen(context.Context, string, int64) (bool, error)
	Record(context.Context, contract.Event, bool) error
}
type Sender interface {
	Send(context.Context, contract.Event) error
}
type Service struct {
	Store  Dedup
	Sender Sender
}

// Send precedes the durable marker. A crash in between may duplicate delivery.
// One sequential consumer owns this SQLite file; it is not a distributed lock.
func (s *Service) Handle(ctx context.Context, e contract.Event) error {
	seen, err := s.Store.Seen(ctx, e.HazardID, e.Version)
	if err != nil {
		return err
	}
	if seen {
		return nil
	}
	alert := e.Hazard.Severity == "SIAGA" || e.Hazard.Severity == "AWAS"
	if alert {
		if err = s.Sender.Send(ctx, e); err != nil {
			return err
		}
	}
	return s.Store.Record(ctx, e, alert)
}
