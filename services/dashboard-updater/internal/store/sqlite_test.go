package store

import (
	"context"
	"encoding/json"
	"example.com/akuanaktehat/dashboard-updater/internal/contract"
	"path/filepath"
	"testing"
)

func TestPersistentLatestVersion(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "view.db")
	s, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	e := contract.Event{HazardID: "hazard", Version: 2, EventID: "event-2", Raw: json.RawMessage(`{"hazard_id":"hazard","version":2,"future":9007199254740993}`)}
	if err = s.Apply(ctx, e); err != nil {
		t.Fatal(err)
	}
	for _, v := range []int64{2, 1} {
		e.Version = v
		e.Raw = json.RawMessage(`{"version":0}`)
		if err = s.Apply(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	data, err := s.List(ctx, "", 10)
	if err != nil || len(data) != 1 || string(data[0]) != `{"hazard_id":"hazard","version":2,"future":9007199254740993}` {
		t.Fatal("replay/stale version changed durable view", data, err)
	}
	e.Version = 3
	e.Raw = json.RawMessage(`{"version":3}`)
	if err = s.Apply(ctx, e); err != nil {
		t.Fatal(err)
	}
	data, err = s.List(ctx, "", 10)
	if err != nil || string(data[0]) != `{"version":3}` {
		t.Fatal("new version not applied")
	}
	data, err = s.List(ctx, "hazard", 10)
	if err != nil || len(data) != 0 {
		t.Fatal("cursor not exclusive")
	}
}
