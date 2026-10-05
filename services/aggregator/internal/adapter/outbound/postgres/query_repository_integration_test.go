package postgres

import (
	"context"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/application/query"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestQueryPostgres(t *testing.T) {
	if os.Getenv("AGGREGATOR_DB_TEST") != "1" {
		t.Skip("set AGGREGATOR_DB_TEST=1 with DATABASE_URL; uses isolated temporary schema")
	}
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is missing")
	}
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal("test database unavailable")
	}
	defer admin.Close(ctx)

	schema := fmt.Sprintf("query_test_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Error("test schema cleanup failed", err)
		}
	}()

	databaseURL, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("invalid DATABASE_URL")
	}
	params := databaseURL.Query()
	params.Set("search_path", schema)
	databaseURL.RawQuery = params.Encode()
	testDSN := databaseURL.String()
	if err = Migrate(ctx, testDSN, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	db, err := Open(ctx, testDSN, 3, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Pool.Close()

	base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	insert := func(id, source, kind, severity string, occurred time.Time) {
		t.Helper()
		_, err = db.Pool.Exec(ctx, `INSERT INTO hazard_events(hazard_id,source,source_ref_id,hazard_type,severity,area_name,latitude,longitude,occurred_at,ingested_at,attributes,version,content_hash,updated_at,last_seen_at) VALUES($1::uuid,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,1,$12,$10,$10)`, id, source, "ref-"+id[len(id)-1:], kind, severity, "Demo", -7.0, 110.0, occurred, occurred.Add(time.Second), `{"magnitude":6.7}`, []byte{1, 2, 3})
		if err != nil {
			t.Fatal(err)
		}
	}
	insert("00000000-0000-0000-0000-000000000001", "BMKG", "SEISMIC", "SIAGA", base.Add(3*time.Minute))
	insert("00000000-0000-0000-0000-000000000002", "PVMBG", "VOLCANIC", "AWAS", base.Add(2*time.Minute))
	insert("00000000-0000-0000-0000-000000000003", "BMKG", "SEISMIC", "NORMAL", base.Add(time.Minute))
	insert("00000000-0000-0000-0000-000000000004", "BMKG", "SEISMIC", "WASPADA", base.Add(time.Minute))
	stale := base.Add(10 * time.Minute)
	if _, err = db.Pool.Exec(ctx, `UPDATE source_status SET status='HEALTHY',stale_since=NULL WHERE source='BMKG'`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Pool.Exec(ctx, `UPDATE source_status SET status='DOWN',stale_since=$1 WHERE source='PVMBG'`, stale); err != nil {
		t.Fatal(err)
	}

	first, err := db.List(ctx, query.HazardFilter{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Data) != 2 || first.Data[0].ID != "00000000-0000-0000-0000-000000000001" || first.Data[1].ID != "00000000-0000-0000-0000-000000000002" {
		t.Fatalf("unexpected first page: %+v", first.Data)
	}
	if first.NextCursor == "" {
		t.Fatal("first page did not return a cursor")
	}
	if len(first.Sources) != 2 || first.Sources[1].StaleSince == nil || !first.Sources[1].StaleSince.Equal(stale) {
		t.Fatalf("unexpected source metadata: %+v", first.Sources)
	}
	cursor, err := query.DecodeCursor(first.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.List(ctx, query.HazardFilter{Limit: 2, Cursor: &cursor})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Data) != 2 || second.Data[0].ID != "00000000-0000-0000-0000-000000000004" || second.Data[1].ID != "00000000-0000-0000-0000-000000000003" || second.NextCursor != "" {
		t.Fatalf("unexpected second page: %+v, cursor=%q", second.Data, second.NextCursor)
	}

	since := base.Add(time.Minute)
	filtered, err := db.List(ctx, query.HazardFilter{Type: "SEISMIC", Since: &since, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.Data) != 3 || filtered.Data[0].ID != "00000000-0000-0000-0000-000000000001" || filtered.Sources[0].Source != "BMKG" || len(filtered.Sources) != 1 {
		t.Fatalf("unexpected filtered result: %+v, sources=%+v", filtered.Data, filtered.Sources)
	}

	event, err := db.Get(ctx, "00000000-0000-0000-0000-000000000001")
	if err != nil || event.Attributes["magnitude"] == nil {
		t.Fatalf("get failed: event=%+v err=%v", event, err)
	}
	_, err = db.Get(ctx, "00000000-0000-0000-0000-000000000099")
	if !errors.Is(err, query.ErrNotFound) {
		t.Fatalf("missing hazard error=%v; want not found", err)
	}
	if strings.Contains(first.NextCursor, "/") || strings.Contains(first.NextCursor, "+") || strings.Contains(first.NextCursor, "=") {
		t.Fatal("cursor is not unpadded base64url")
	}
}
