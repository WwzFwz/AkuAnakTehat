package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"example.com/akuanaktehat/dashboard-updater/internal/contract"
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
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS hazard_view(hazard_id TEXT PRIMARY KEY,version INTEGER NOT NULL,event_id TEXT NOT NULL,payload TEXT NOT NULL,updated_at TEXT NOT NULL)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db}, nil
}
func (s *Store) Apply(ctx context.Context, e contract.Event) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO hazard_view(hazard_id,version,event_id,payload,updated_at) VALUES(?,?,?,?,?) ON CONFLICT(hazard_id) DO UPDATE SET version=excluded.version,event_id=excluded.event_id,payload=excluded.payload,updated_at=excluded.updated_at WHERE excluded.version>hazard_view.version`, e.HazardID, e.Version, e.EventID, string(e.Raw), time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
func (s *Store) List(ctx context.Context, after string, limit int) ([]json.RawMessage, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT payload FROM hazard_view WHERE hazard_id>? ORDER BY hazard_id LIMIT ?`, after, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []json.RawMessage{}
	bytes := 0
	for rows.Next() {
		var raw string
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		result = append(result, json.RawMessage(raw))
		bytes += len(raw)
		if bytes > 7<<20 {
			break
		}
	}
	return result, rows.Err()
}
func (s *Store) Ping(ctx context.Context) error { return s.DB.PingContext(ctx) }
func (s *Store) Close() error                   { return s.DB.Close() }
