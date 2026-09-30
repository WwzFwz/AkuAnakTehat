package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"time"
)

type Service struct {
	UOW       UnitOfWork
	Volcanoes map[string]canonicalize.Volcano
	Now       func() time.Time
}
type Stats struct{ Changed, Unchanged, Rejected int }
type Envelope struct {
	SchemaVersion int          `json:"schema_version"`
	EventID       string       `json:"event_id"`
	EventType     string       `json:"event_type"`
	HazardID      string       `json:"hazard_id"`
	Version       int64        `json:"version"`
	CorrelationID string       `json:"correlation_id"`
	PublishedAt   time.Time    `json:"published_at"`
	Hazard        hazard.Event `json:"hazard"`
}

func (s *Service) ApplyBatch(ctx context.Context, b canonicalize.Batch) (Stats, error) {
	stats := Stats{}
	source := canonicalize.Source(b.Endpoint)
	if source == "" || b.Watermark.IsZero() || b.CorrelationID == "" {
		return stats, errors.New("invalid batch metadata")
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	now = now.Truncate(time.Microsecond)
	err := s.UOW.WithTx(ctx, func(tx Tx) error {
		reject := func(item canonicalize.Item, reason string) error {
			stats.Rejected++
			return tx.Quarantine(ctx, RejectedRecord{Source: source, Endpoint: b.Endpoint, Payload: item.Raw, Reason: reason, CorrelationID: b.CorrelationID, ObservedAt: now})
		}
		for _, item := range b.Items {
			if item.Reason != "" {
				if err := reject(item, item.Reason); err != nil {
					return err
				}
				continue
			}
			if item.Warning != nil {
				w := *item.Warning
				old, found, err := tx.FindWarning(ctx, w.ID)
				if err != nil {
					return err
				}
				if found && old.RelatedID != w.RelatedID {
					if err = reject(item, "warning related_event_id is immutable"); err != nil {
						return err
					}
					continue
				}
				if err = tx.PutWarning(ctx, w); err != nil {
					return err
				}
				record, exists, err := tx.FindHazard(ctx, "BMKG", w.RelatedID)
				if err != nil {
					return err
				}
				if !exists {
					continue
				}
				warnings, err := tx.WarningsFor(ctx, w.RelatedID)
				if err != nil {
					return err
				}
				e := record.Event
				canonicalize.Correlate(&e, warnings)
				if err = s.save(ctx, tx, e, record, true, b.CorrelationID, now, &stats); err != nil {
					return err
				}
				continue
			}
			var e hazard.Event
			if item.Seismic != nil {
				warnings, err := tx.WarningsFor(ctx, item.Seismic.ID)
				if err != nil {
					return err
				}
				e = canonicalize.MapSeismic(*item.Seismic, warnings)
			} else if item.Volcanic != nil {
				var err error
				e, err = canonicalize.MapVolcanic(*item.Volcanic, s.Volcanoes)
				if err != nil {
					if err = reject(item, err.Error()); err != nil {
						return err
					}
					continue
				}
			} else {
				return errors.New("batch item missing typed input")
			}
			if _, err := hazard.ContentHash(e); err != nil {
				if err = reject(item, "unsupported attribute number"); err != nil {
					return err
				}
				continue
			}
			old, found, err := tx.FindHazard(ctx, e.Source, e.SourceRefID)
			if err != nil {
				return err
			}
			if err = s.save(ctx, tx, e, old, found, b.CorrelationID, now, &stats); err != nil {
				return err
			}
		}
		return tx.SaveCheckpoint(ctx, b.Endpoint, b.Watermark)
	})
	if err != nil {
		return Stats{}, err
	}
	return stats, nil
}
func (s *Service) save(ctx context.Context, tx Tx, e hazard.Event, old hazard.Record, found bool, corr string, now time.Time, stats *Stats) error {
	hash, err := hazard.ContentHash(e)
	if err != nil {
		return err
	}
	record := hazard.Record{Event: e, Version: 1, Hash: hash, UpdatedAt: now, LastSeenAt: now}
	if found {
		record.Event.ID = old.Event.ID
		record.Event.IngestedAt = old.Event.IngestedAt
		record.Version = old.Version
		if bytes.Equal(hash, old.Hash) {
			record.UpdatedAt = old.UpdatedAt
			stats.Unchanged++
			return tx.PutHazard(ctx, record)
		}
		record.Version++
	} else {
		record.Event.ID, err = hazard.UUID()
		if err != nil {
			return err
		}
		record.Event.IngestedAt = now
	}
	if err = tx.PutHazard(ctx, record); err != nil {
		return err
	}
	eventID, err := hazard.UUID()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(Envelope{1, eventID, "hazard.upserted", record.Event.ID, record.Version, corr, now, record.Event})
	if err != nil {
		return err
	}
	if err = tx.AppendOutbox(ctx, OutboxEvent{eventID, record.Event.ID, record.Version, payload, now}); err != nil {
		return err
	}
	stats.Changed++
	return nil
}
