package dedup

import (
	"context"
	"database/sql"
	"encoding/json"
	"example.com/akuanaktehat/notifier/internal/contract"
	_ "modernc.org/sqlite"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type Store struct{ DB *sql.DB }

func Open(ctx context.Context, path string) (*Store, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(absolute), 0700); err != nil {
		return nil, err
	}
	uriPath := filepath.ToSlash(absolute)
	if filepath.VolumeName(absolute) != "" {
		uriPath = "/" + uriPath
	}
	u := url.URL{Scheme: "file", Path: uriPath}
	q := u.Query()
	q.Add("_pragma", "busy_timeout(1000)")
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(FULL)")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS processed(hazard_id TEXT NOT NULL,version INTEGER NOT NULL,event_id TEXT NOT NULL,alert INTEGER NOT NULL,processed_at TEXT NOT NULL,PRIMARY KEY(hazard_id,version));CREATE INDEX IF NOT EXISTS processed_event ON processed(event_id)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db}, nil
}
func (s *Store) Seen(ctx context.Context, id string, version int64) (bool, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM processed WHERE hazard_id=? AND version=?`, id, version).Scan(&n)
	return n > 0, err
}
func (s *Store) Record(ctx context.Context, e contract.Event, alert bool) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO processed(hazard_id,version,event_id,alert,processed_at) VALUES(?,?,?,?,?) ON CONFLICT(hazard_id,version) DO NOTHING`, e.HazardID, e.Version, e.EventID, alert, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
func (s *Store) List(ctx context.Context, after string, limit int) ([]json.RawMessage, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT event_id,hazard_id,version,alert,processed_at FROM processed WHERE event_id>? ORDER BY event_id LIMIT ?`, after, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []json.RawMessage{}
	for rows.Next() {
		var event, id, at string
		var version int64
		var alert bool
		if err = rows.Scan(&event, &id, &version, &alert, &at); err != nil {
			return nil, err
		}
		raw, _ := json.Marshal(map[string]any{"event_id": event, "hazard_id": id, "version": version, "alert": alert, "processed_at": at})
		result = append(result, raw)
	}
	return result, rows.Err()
}
func (s *Store) Ping(ctx context.Context) error { return s.DB.PingContext(ctx) }
func (s *Store) Close() error                   { return s.DB.Close() }
