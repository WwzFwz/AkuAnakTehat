package application

import (
	"context"
	"errors"
	"example.com/akuanaktehat/notifier/internal/contract"
	"example.com/akuanaktehat/notifier/internal/dedup"
	"path/filepath"
	"testing"
)

type sendStub struct {
	sent int
	fail bool
}

func (s *sendStub) Send(context.Context, contract.Event) error {
	if s.fail {
		return errors.New("delivery failed")
	}
	s.sent++
	return nil
}
func TestPersistentDedupAndDeliveryOrder(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "processed.db")
	db, err := dedup.Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	sender := &sendStub{fail: true}
	svc := &Service{Store: db, Sender: sender}
	e := contract.Event{HazardID: "hazard", Version: 1, EventID: "event-1", Hazard: contract.Hazard{Severity: "SIAGA"}}
	if svc.Handle(ctx, e) == nil {
		t.Fatal("send failure hidden")
	}
	if seen, err := db.Seen(ctx, e.HazardID, e.Version); err != nil || seen {
		t.Fatal("dedup written before delivery")
	}
	sender.fail = false
	if err = svc.Handle(ctx, e); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = dedup.Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc.Store = db
	if err = svc.Handle(ctx, e); err != nil {
		t.Fatal(err)
	}
	if sender.sent != 1 {
		t.Fatal("replay after restart sent twice")
	}
	e.Version = 2
	e.EventID = "event-2"
	e.Hazard.Severity = "NORMAL"
	if err = svc.Handle(ctx, e); err != nil {
		t.Fatal(err)
	}
	if sender.sent != 1 {
		t.Fatal("NORMAL sent notification")
	}
	e.Version = 3
	e.EventID = "event-3"
	e.Hazard.Severity = "AWAS"
	if err = svc.Handle(ctx, e); err != nil {
		t.Fatal(err)
	}
	if sender.sent != 2 {
		t.Fatal("new alert version not sent")
	}
	rows, err := db.List(ctx, "", 10)
	if err != nil || len(rows) != 3 {
		t.Fatal("durable audit incomplete")
	}
}
