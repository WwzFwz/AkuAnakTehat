package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"example.com/akuanaktehat/aggregator/internal/application/ingest"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"example.com/akuanaktehat/aggregator/internal/domain/tsunami"
	"github.com/jackc/pgx/v5"
	"time"
)

func (t *transaction) FindHazard(ctx context.Context, source, ref string) (hazard.Record, bool, error) {
	ctx, cancel := t.bounded(ctx)
	defer cancel()
	var r hazard.Record
	var attributes []byte
	e := &r.Event
	err := t.tx.QueryRow(ctx, `SELECT hazard_id::text,source,source_ref_id,hazard_type,severity,area_name,latitude,longitude,occurred_at,ingested_at,attributes,version,content_hash,updated_at,last_seen_at FROM hazard_events WHERE source=$1 AND source_ref_id=$2 FOR UPDATE`, source, ref).Scan(&e.ID, &e.Source, &e.SourceRefID, &e.Type, &e.Severity, &e.Area, &e.Latitude, &e.Longitude, &e.OccurredAt, &e.IngestedAt, &attributes, &r.Version, &r.Hash, &r.UpdatedAt, &r.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, false, nil
	}
	if err != nil {
		return r, false, err
	}
	e.OccurredAt = e.OccurredAt.UTC()
	e.IngestedAt = e.IngestedAt.UTC()
	err = json.Unmarshal(attributes, &e.Attributes)
	return r, true, err
}
func (t *transaction) PutHazard(ctx context.Context, r hazard.Record) error {
	ctx, cancel := t.bounded(ctx)
	defer cancel()
	e := r.Event
	a, err := json.Marshal(e.Attributes)
	if err != nil {
		return err
	}
	_, err = t.tx.Exec(ctx, `INSERT INTO hazard_events(hazard_id,source,source_ref_id,hazard_type,severity,area_name,latitude,longitude,occurred_at,ingested_at,attributes,version,content_hash,updated_at,last_seen_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT(source,source_ref_id) DO UPDATE SET hazard_type=excluded.hazard_type,severity=excluded.severity,area_name=excluded.area_name,latitude=excluded.latitude,longitude=excluded.longitude,occurred_at=excluded.occurred_at,attributes=excluded.attributes,version=excluded.version,content_hash=excluded.content_hash,updated_at=excluded.updated_at,last_seen_at=excluded.last_seen_at`, e.ID, e.Source, e.SourceRefID, e.Type, e.Severity, e.Area, e.Latitude, e.Longitude, e.OccurredAt, e.IngestedAt, a, r.Version, r.Hash, r.UpdatedAt, r.LastSeenAt)
	return err
}
func (t *transaction) FindWarning(ctx context.Context, id string) (tsunami.Warning, bool, error) {
	ctx, cancel := t.bounded(ctx)
	defer cancel()
	var b []byte
	var w tsunami.Warning
	err := t.tx.QueryRow(ctx, `SELECT payload FROM tsunami_warnings WHERE warning_id=$1`, id).Scan(&b)
	if errors.Is(err, pgx.ErrNoRows) {
		return w, false, nil
	}
	if err != nil {
		return w, false, err
	}
	err = json.Unmarshal(b, &w)
	return w, true, err
}
func (t *transaction) PutWarning(ctx context.Context, w tsunami.Warning) error {
	ctx, cancel := t.bounded(ctx)
	defer cancel()
	b, err := json.Marshal(w)
	if err != nil {
		return err
	}
	_, err = t.tx.Exec(ctx, `INSERT INTO tsunami_warnings(warning_id,related_event_id,payload,updated_at) VALUES($1,$2,$3,now()) ON CONFLICT(warning_id) DO UPDATE SET payload=excluded.payload,updated_at=excluded.updated_at`, w.ID, w.RelatedID, b)
	return err
}
func (t *transaction) WarningsFor(ctx context.Context, id string) ([]tsunami.Warning, error) {
	ctx, cancel := t.bounded(ctx)
	defer cancel()
	rows, err := t.tx.Query(ctx, `SELECT payload FROM tsunami_warnings WHERE related_event_id=$1 ORDER BY warning_id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []tsunami.Warning{}
	for rows.Next() {
		var b []byte
		var w tsunami.Warning
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &w); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
func (t *transaction) AppendOutbox(ctx context.Context, e ingest.OutboxEvent) error {
	ctx, cancel := t.bounded(ctx)
	defer cancel()
	_, err := t.tx.Exec(ctx, `INSERT INTO outbox(event_id,hazard_id,version,payload,created_at) VALUES($1,$2,$3,$4,$5)`, e.EventID, e.HazardID, e.Version, e.Payload, e.CreatedAt)
	return err
}
func (t *transaction) SaveCheckpoint(ctx context.Context, endpoint string, watermark time.Time) error {
	ctx, cancel := t.bounded(ctx)
	defer cancel()
	_, err := t.tx.Exec(ctx, `INSERT INTO checkpoints(endpoint,watermark) VALUES($1,$2) ON CONFLICT(endpoint) DO UPDATE SET watermark=greatest(checkpoints.watermark,excluded.watermark)`, endpoint, watermark)
	return err
}
func (t *transaction) Quarantine(ctx context.Context, r ingest.RejectedRecord) error {
	ctx, cancel := t.bounded(ctx)
	defer cancel()
	// Preserve the exact input as text: rejected numbers/NUL escapes must not
	// fail JSONB conversion and roll back otherwise valid records.
	payload, err := json.Marshal(map[string]string{"raw_json": string(r.Payload)})
	if err != nil {
		return err
	}
	_, err = t.tx.Exec(ctx, `INSERT INTO quarantine(source,endpoint,reason,correlation_id,payload,observed_at) VALUES($1,$2,$3,$4,$5,$6)`, r.Source, r.Endpoint, r.Reason, r.CorrelationID, payload, r.ObservedAt)
	return err
}
func (s *Store) ReadCheckpoint(ctx context.Context, endpoint string) (time.Time, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	var stamp time.Time
	err := s.Pool.QueryRow(ctx, `SELECT watermark FROM checkpoints WHERE endpoint=$1`, endpoint).Scan(&stamp)
	if errors.Is(err, pgx.ErrNoRows) {
		return stamp, false, nil
	}
	return stamp, err == nil, err
}

func (s *Store) RecordPoll(ctx context.Context, endpoint string, ok, degraded bool, code string, at time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	source := canonicalize.Source(endpoint)
	_, err = tx.Exec(ctx, `UPDATE source_endpoint_status SET healthy=$2,last_attempt_at=$3,last_success_at=CASE WHEN $4 THEN $3 ELSE last_success_at END,consecutive_failures=CASE WHEN $4 THEN 0 ELSE consecutive_failures+1 END,last_error=NULLIF($5,'') WHERE endpoint=$1`, endpoint, ok && !degraded, at, ok, code)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE source_status SET status=s.state,last_attempt_at=$2,last_success_at=CASE WHEN s.state='HEALTHY' THEN s.success ELSE source_status.last_success_at END,consecutive_failures=s.failures,last_error=s.error FROM (SELECT CASE WHEN bool_and(healthy) THEN 'HEALTHY' WHEN bool_or(healthy) OR bool_or(consecutive_failures=0 AND last_success_at IS NOT NULL) THEN 'DEGRADED' ELSE 'DOWN' END state,min(last_success_at) success,sum(consecutive_failures)::int failures,max(last_error) error FROM source_endpoint_status WHERE source=$1) s WHERE source_status.source=$1`, source, at)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
