package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/worker/outbox"
	"time"
)

func (s *Store) Pending(ctx context.Context, limit int) ([]outbox.Message, error) {
	if limit < 1 || limit > 1000 {
		return nil, errors.New("invalid outbox limit")
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	rows, err := s.Pool.Query(ctx, `SELECT id,event_id::text,hazard_id::text,version,payload,COALESCE(payload->>'correlation_id','') FROM outbox WHERE published_at IS NULL AND rejected_at IS NULL ORDER BY id LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []outbox.Message{}
	for rows.Next() {
		var m outbox.Message
		if err = rows.Scan(&m.ID, &m.EventID, &m.HazardID, &m.Version, &m.Payload, &m.CorrelationID); err != nil {
			return nil, err
		}
		var compact bytes.Buffer
		if err = json.Compact(&compact, m.Payload); err != nil {
			return nil, err
		}
		m.Payload = append([]byte(nil), compact.Bytes()...)
		result = append(result, m)
	}
	return result, rows.Err()
}
func (s *Store) MarkPublished(ctx context.Context, id int64, at time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	tag, err := s.Pool.Exec(ctx, `UPDATE outbox SET published_at=COALESCE(published_at,$2) WHERE id=$1`, id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("outbox row not found")
	}
	return nil
}
func (s *Store) DeletePublishedBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	tag, err := s.Pool.Exec(ctx, `DELETE FROM outbox WHERE published_at IS NOT NULL AND published_at<$1`, cutoff)
	return tag.RowsAffected(), err
}

// Preserve rejected payloads for inspection/redrive; never mark them published.
func (s *Store) Reject(ctx context.Context, id int64, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	tag, err := s.Pool.Exec(ctx, `UPDATE outbox SET rejected_at=now(),rejection_reason=$2 WHERE id=$1 AND published_at IS NULL`, id, reason)
	if err == nil && tag.RowsAffected() != 1 {
		return errors.New("outbox row not pending")
	}
	return err
}
