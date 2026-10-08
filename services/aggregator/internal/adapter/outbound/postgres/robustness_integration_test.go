package postgres

import (
	"context"
	"encoding/json"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"example.com/akuanaktehat/aggregator/internal/application/ingest"
	"example.com/akuanaktehat/aggregator/internal/application/query"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func exerciseRobustness(t *testing.T, db *Store, svc *ingest.Service) {
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer slog.SetDefault(oldLogger)
	ctx := context.Background()
	stamp := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.Now = func() time.Time { return stamp }
	defer func() { svc.Now = nil }()
	count := func(table string) int {
		var n int
		if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	t.Run("compact envelope boundary and oversized record isolation", func(t *testing.T) {
		initial := count("hazard_events")
		for _, delta := range []int{-1, 0, 1} {
			v := canonicalize.VolcanicInput{ID: fmt.Sprintf("size-%d", delta), VolcanoID: "VOLCANO-DEMO-01", Alert: "Siaga", ReportedAt: stamp, Extra: map[string]json.RawMessage{"blob": json.RawMessage(`""`)}}
			event, _ := canonicalize.MapVolcanic(v, svc.Volcanoes)
			event.ID = "00000000-0000-0000-0000-000000000001"
			event.IngestedAt = stamp
			raw, _ := json.Marshal(ingest.Envelope{SchemaVersion: 1, EventID: event.ID, EventType: "hazard.upserted", HazardID: event.ID, Version: 1, CorrelationID: "boundary", PublishedAt: stamp, Hazard: event})
			v.Extra["blob"], _ = json.Marshal(strings.Repeat("x", hazard.MaxEventBytes-len(raw)+delta))
			source, _ := json.Marshal(map[string]any{"report_id": v.ID, "volcano_id": v.VolcanoID, "alert_level": v.Alert, "eruption_count_24h": 0, "ash_column_height_m": 0, "reported_at": stamp, "blob": v.Extra["blob"]})
			normal := canonicalize.VolcanicInput{ID: fmt.Sprintf("after-size-%d", delta), VolcanoID: v.VolcanoID, Alert: "Normal", ReportedAt: stamp}
			stats, err := svc.ApplyBatch(ctx, canonicalize.Batch{Endpoint: canonicalize.VolcanicEndpoint, Watermark: stamp, CorrelationID: "boundary", Items: []canonicalize.Item{{Raw: source, Volcanic: &v}, {Raw: []byte(`{}`), Volcanic: &normal}}})
			if err != nil {
				t.Fatal(err)
			}
			if delta <= 0 && (stats.Changed != 2 || stats.Rejected != 0) {
				t.Fatal(stats)
			}
			if delta > 0 && (stats.Changed != 1 || stats.Rejected != 1) {
				t.Fatal(stats)
			}
			if delta <= 0 {
				var payload []byte
				err = db.Pool.QueryRow(ctx, `SELECT o.payload FROM outbox o JOIN hazard_events h USING(hazard_id) WHERE h.source_ref_id=$1`, v.ID).Scan(&payload)
				if err != nil {
					t.Fatal(err)
				}
				var e ingest.Envelope
				if err = json.Unmarshal(payload, &e); err != nil {
					t.Fatal(err)
				}
				compact, _ := json.Marshal(e)
				if len(compact) != hazard.MaxEventBytes+delta {
					t.Fatalf("size=%d", len(compact))
				}
			}
		}
		if count("hazard_events") != initial+5 {
			t.Fatal("oversized record entered canonical store or blocked following record")
		}
		// Byte-limited pages must retain every valid record without exceeding the HTTP budget.
		found := map[string]bool{}
		var cursor *query.Cursor
		for pages := 0; pages < 8; pages++ {
			page, err := db.List(ctx, query.HazardFilter{Type: "VOLCANIC", Since: &stamp, Limit: 500, Cursor: cursor})
			if err != nil {
				t.Fatal(err)
			}
			wire, _ := json.Marshal(page)
			if len(wire) > hazard.MaxPageBytes {
				t.Fatal("page exceeds HTTP budget")
			}
			for _, e := range page.Data {
				if found[e.ID] {
					t.Fatal("duplicate page item")
				}
				found[e.ID] = true
			}
			if page.NextCursor == "" {
				break
			}
			c, err := query.DecodeCursor(page.NextCursor)
			if err != nil {
				t.Fatal(err)
			}
			cursor = &c
		}
		if len(found) != 5 {
			t.Fatalf("byte pagination lost data: %d", len(found))
		}
		t.Log("4 MiB - 1 and 4 MiB accepted; 4 MiB + 1 quarantined; following records and byte pagination pass")
	})
	t.Run("backlog spans transaction deadlines and safely resumes", func(t *testing.T) {
		// Slow real PostgreSQL writes, then fail the middle of a response.
		_, err := db.Pool.Exec(ctx, `CREATE FUNCTION backlog_delay() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.source_ref_id LIKE 'backlog-%' THEN PERFORM pg_sleep(0.004); IF NEW.source_ref_id='backlog-300' THEN RAISE EXCEPTION 'temporary test failure'; END IF; END IF; RETURN NEW; END $$; CREATE TRIGGER backlog_delay BEFORE INSERT ON hazard_events FOR EACH ROW EXECUTE FUNCTION backlog_delay()`)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Pool.Exec(ctx, `DROP TRIGGER IF EXISTS backlog_delay ON hazard_events; DROP FUNCTION IF EXISTS backlog_delay()`)
		items := make([]canonicalize.Item, 600)
		for i := range items {
			items[i] = canonicalize.Item{Raw: []byte(`{}`), Volcanic: &canonicalize.VolcanicInput{ID: fmt.Sprintf("backlog-%03d", i), VolcanoID: "VOLCANO-DEMO-01", Alert: "Normal", ReportedAt: stamp}}
		}
		b := canonicalize.Batch{Endpoint: canonicalize.VolcanicEndpoint, Watermark: stamp.Add(time.Hour), CorrelationID: "backlog", Items: items}
		old, _, _ := db.ReadCheckpoint(ctx, b.Endpoint)
		before := count("outbox")
		start := time.Now()
		stats, err := svc.ApplyBatch(ctx, b)
		if err == nil || stats.Changed != 300 {
			t.Fatal("partial failure not preserved", stats, err)
		}
		checkpoint, _, _ := db.ReadCheckpoint(ctx, b.Endpoint)
		if !checkpoint.Equal(old) {
			t.Fatal("partial response advanced checkpoint")
		}
		_, err = db.Pool.Exec(ctx, `CREATE OR REPLACE FUNCTION backlog_delay() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.source_ref_id LIKE 'backlog-%' THEN PERFORM pg_sleep(0.004); END IF; RETURN NEW; END $$`)
		if err != nil {
			t.Fatal(err)
		}
		stats, err = svc.ApplyBatch(ctx, b)
		if err != nil || stats.Changed != 300 || stats.Unchanged != 300 {
			t.Fatal("resume failed", stats, err)
		}
		if count("outbox") != before+600 {
			t.Fatal("replay duplicated events")
		}
		checkpoint, _, _ = db.ReadCheckpoint(ctx, b.Endpoint)
		if !checkpoint.Equal(b.Watermark) {
			t.Fatal("complete response missing checkpoint")
		}
		if time.Since(start) <= db.Timeout {
			t.Fatal("test did not span the transaction deadline")
		}
		t.Logf("600-record response, mid-batch failure, replay and recovery passed in %s with transaction timeout %s", time.Since(start), db.Timeout)
	})
}
