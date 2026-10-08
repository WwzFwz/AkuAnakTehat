package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"example.com/akuanaktehat/aggregator/internal/application/ingest"
	"example.com/akuanaktehat/aggregator/reference"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestIngestPostgres(t *testing.T) {
	if os.Getenv("AGGREGATOR_DB_TEST") != "1" {
		t.Skip("set AGGREGATOR_DB_TEST=1 with DATABASE_URL; uses isolated temporary schema")
	}
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is missing; pass --env-from-file ./env/aggregator.env to compose run")
	}
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		safe := strings.ReplaceAll(err.Error(), dsn, "[DSN]")
		if u, e := url.Parse(dsn); e == nil && u.User != nil {
			if password, ok := u.User.Password(); ok && password != "" {
				safe = strings.ReplaceAll(safe, password, "[redacted]")
			}
		}
		t.Fatal("test database unavailable:", safe)
	}
	defer admin.Close(ctx)
	schema := fmt.Sprintf("ingest_test_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Error("test schema cleanup failed")
		}
	}()
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("invalid DSN")
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	dsn = u.String()
	if err = Migrate(ctx, dsn, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(ctx, dsn, 2*time.Second); err != nil {
		t.Fatal("restart migration failed", err)
	}
	db, err := Open(ctx, dsn, 5, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Pool.Close()
	refs, _ := reference.Load()
	svc := &ingest.Service{UOW: db, Volcanoes: refs}
	stamp := time.Now().UTC().Truncate(time.Microsecond)
	apply := func(endpoint, body string) ingest.Stats {
		t.Helper()
		items, err := canonicalize.Decode(endpoint, []byte(body))
		if err != nil {
			t.Fatal(err)
		}
		stats, err := svc.ApplyBatch(ctx, canonicalize.Batch{Endpoint: endpoint, CorrelationID: "integration", Watermark: stamp, Items: items})
		if err != nil {
			t.Fatal(err)
		}
		return stats
	}
	count := func(table string) int {
		t.Helper()
		var n int
		if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	warning := `[{"warning_id":"w","related_event_id":"e","threat_level":"Waspada","affected_zones":["Coast"],"estimated_arrival":"2026-09-01T01:00:00Z"}]`
	event := `[{"event_id":"e","magnitude":7,"depth_km":10,"epicenter_lat":-7,"epicenter_lon":110,"region_name":"Demo","occurred_at":"2026-09-01T00:00:00.123456789Z","potential_tsunami":true}]`
	apply(canonicalize.WarningEndpoint, warning)
	if count("tsunami_warnings") != 1 || count("hazard_events") != 0 {
		t.Fatal("early warning not retained independently")
	}
	if stats := apply(canonicalize.SeismicEndpoint, event); stats.Changed != 1 {
		t.Fatal(stats)
	}
	var id, severity string
	var version int64
	if err = db.Pool.QueryRow(ctx, "SELECT hazard_id::text,severity,version FROM hazard_events WHERE source_ref_id='e'").Scan(&id, &severity, &version); err != nil || severity != "WASPADA" || version != 1 {
		t.Fatal("warning precedence", err, severity, version)
	}
	if stats := apply(canonicalize.SeismicEndpoint, event); stats.Unchanged != 1 || count("outbox") != 1 {
		t.Fatal("replay created duplicate outbox")
	}
	late := `[{"warning_id":"w","related_event_id":"e","threat_level":"Awas","affected_zones":["Coast"],"estimated_arrival":"2026-09-01T01:00:00Z"}]`
	apply(canonicalize.WarningEndpoint, late)
	apply(canonicalize.WarningEndpoint, late)
	apply(canonicalize.SeismicEndpoint, event)
	if count("outbox") != 2 {
		t.Fatal("warning replay or timestamp precision produced duplicate outbox")
	}
	var payload []byte
	if err = db.Pool.QueryRow(ctx, "SELECT payload FROM outbox ORDER BY id DESC LIMIT 1").Scan(&payload); err != nil {
		t.Fatal(err)
	}
	var envelope ingest.Envelope
	if err = json.Unmarshal(payload, &envelope); err != nil || envelope.HazardID != id || envelope.Version != 2 || envelope.Hazard.Severity != "AWAS" {
		t.Fatal("incorrect snapshot envelope")
	}
	volcanic := `[{"report_id":"v1","volcano_id":"VOLCANO-DEMO-01","alert_level":"Siaga","eruption_count_24h":2,"ash_column_height_m":100,"reported_at":"2026-09-01T00:00:00Z","confidence_level":0.9},{"report_id":"invalid"}]`
	stats := apply(canonicalize.VolcanicEndpoint, volcanic)
	if stats.Changed != 1 || stats.Rejected != 1 || count("quarantine") != 1 {
		t.Fatal("partial batch quarantine failed", stats)
	}
	if _, exists, err := db.ReadCheckpoint(ctx, canonicalize.VolcanicEndpoint); err != nil || !exists {
		t.Fatal("handled batch did not checkpoint")
	}
	badValues := `[{"report_id":"nul","future":"\u0000"},{"report_id":"huge","future":0e999999999}]`
	if stats := apply(canonicalize.VolcanicEndpoint, badValues); stats.Rejected != 2 || count("quarantine") != 3 {
		t.Fatal("unrepresentable JSON values poisoned quarantine")
	}
	// Inject a real PostgreSQL error after hazard/outbox writes, at checkpoint.
	beforeHazard, beforeOutbox := count("hazard_events"), count("outbox")
	if _, err = db.Pool.Exec(ctx, `ALTER TABLE checkpoints ADD CONSTRAINT fail_test CHECK(endpoint <> 'pvmbg.volcanic-reports') NOT VALID`); err != nil {
		t.Fatal(err)
	}
	items, _ := canonicalize.Decode(canonicalize.VolcanicEndpoint, []byte(`[ {"report_id":"rollback","volcano_id":"VOLCANO-DEMO-02","alert_level":"Normal","eruption_count_24h":0,"ash_column_height_m":0,"reported_at":"2026-09-01T00:00:00Z"}]`))
	_, err = svc.ApplyBatch(ctx, canonicalize.Batch{Endpoint: canonicalize.VolcanicEndpoint, Watermark: stamp.Add(time.Minute), CorrelationID: "rollback", Items: items})
	if err == nil || count("hazard_events") != beforeHazard+1 || count("outbox") != beforeOutbox+1 {
		t.Fatal("failed checkpoint lost the committed record prefix")
	}
	current, _, err := db.ReadCheckpoint(ctx, canonicalize.VolcanicEndpoint)
	if err != nil || !current.Equal(stamp) {
		t.Fatal("failed transaction advanced checkpoint")
	}
	if _, err = db.Pool.Exec(ctx, `ALTER TABLE checkpoints DROP CONSTRAINT fail_test`); err != nil {
		t.Fatal(err)
	}
	// Database deadline cancels a blocked transaction without leaking a checked-out connection.
	short := &Store{Pool: db.Pool, Timeout: 50 * time.Millisecond}
	err = short.WithTx(ctx, func(tx ingest.Tx) error { _, err := tx.(*transaction).tx.Exec(ctx, "SELECT pg_sleep(1)"); return err })
	if err == nil {
		t.Fatal("database timeout not enforced")
	}
	if err = db.Ping(ctx); err != nil {
		t.Fatal("pool unusable after rollback")
	}
	var operationErr error
	_ = short.WithTx(ctx, func(tx ingest.Tx) error {
		time.Sleep(80 * time.Millisecond)
		operationErr = tx.SaveCheckpoint(ctx, "deadline-test", stamp)
		return operationErr
	})
	if !errors.Is(operationErr, context.DeadlineExceeded) {
		t.Fatal("repository ignored transaction deadline", operationErr)
	}
	sentinel := errors.New("abort")
	if err = db.WithTx(ctx, func(ingest.Tx) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal("callback error lost")
	}
	if err = db.RecordPoll(ctx, canonicalize.SeismicEndpoint, true, false, "", stamp); err != nil {
		t.Fatal(err)
	}
	if err = db.RecordPoll(ctx, canonicalize.WarningEndpoint, false, false, "source_fetch_failed", stamp); err != nil {
		t.Fatal(err)
	}
	var health string
	db.Pool.QueryRow(ctx, `SELECT status FROM source_status WHERE source='BMKG'`).Scan(&health)
	if health != "DEGRADED" {
		t.Fatal("one healthy endpoint hid failure", health)
	}
	if err = db.RecordPoll(ctx, canonicalize.WarningEndpoint, true, false, "", stamp.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	db.Pool.QueryRow(ctx, `SELECT status FROM source_status WHERE source='BMKG'`).Scan(&health)
	if health != "HEALTHY" {
		t.Fatal("source did not recover", health)
	}
	t.Run("outbox ACK and cleanup", func(t *testing.T) {
		pending, err := db.Pending(ctx, 100)
		if err != nil || len(pending) < 2 {
			t.Fatal("missing pending snapshots", err)
		}
		for i, m := range pending {
			if len(m.Payload) == 0 || m.EventID == "" || m.CorrelationID == "" {
				t.Fatal("incomplete relay payload")
			}
			if i > 0 && m.ID <= pending[i-1].ID {
				t.Fatal("outbox order unstable")
			}
		}
		old := time.Now().UTC().Add(-48 * time.Hour)
		if err = db.MarkPublished(ctx, pending[0].ID, old); err != nil {
			t.Fatal(err)
		}
		if err = db.MarkPublished(ctx, pending[0].ID, time.Now()); err != nil {
			t.Fatal(err)
		}
		if err = db.MarkPublished(ctx, pending[1].ID, time.Now()); err != nil {
			t.Fatal(err)
		}
		deleted, err := db.DeletePublishedBefore(ctx, time.Now().Add(-24*time.Hour))
		if err != nil || deleted != 1 {
			t.Fatal("cleanup did not preserve first ACK timestamp", deleted, err)
		}
		remaining, err := db.Pending(ctx, 100)
		if err != nil || len(remaining) != len(pending)-2 {
			t.Fatal("cleanup touched unpublished rows", err)
		}
		if err = db.MarkPublished(ctx, -1, time.Now()); err == nil {
			t.Fatal("missing row marked as published")
		}
	})
	exerciseRobustness(t, db, svc)

}
