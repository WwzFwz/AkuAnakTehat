package outbox

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

type memoryStore struct {
	rows       []Message
	marked     []int64
	rejected   []int64
	failMark   bool
	failReject bool
}

func (s *memoryStore) Pending(context.Context, int) ([]Message, error) { return s.rows, nil }
func (s *memoryStore) MarkPublished(_ context.Context, id int64, _ time.Time) error {
	if s.failMark {
		return errors.New("database down")
	}
	s.marked = append(s.marked, id)
	s.rows = s.rows[1:]
	return nil
}
func (s *memoryStore) DeletePublishedBefore(context.Context, time.Time) (int64, error) { return 0, nil }

type publisher struct {
	sent      []Message
	fail      bool
	permanent bool
}

func (p *publisher) Publish(_ context.Context, m Message) error {
	p.sent = append(p.sent, m)
	if p.permanent && m.ID == 1 {
		return ErrPermanent
	}
	if p.fail {
		return errors.New("no ACK")
	}
	return nil
}
func TestRelayAcknowledgementBoundary(t *testing.T) {
	for _, failure := range []string{"publish", "mark"} {
		t.Run(failure, func(t *testing.T) {
			s := &memoryStore{rows: []Message{{ID: 1, EventID: "stable-1"}, {ID: 2, EventID: "stable-2"}}, failMark: failure == "mark"}
			p := &publisher{fail: failure == "publish"}
			r := &Relay{Store: s, Publisher: p, BatchSize: 100, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
			if r.Drain(context.Background()) == nil {
				t.Fatal("failure hidden")
			}
			if len(s.marked) != 0 || len(p.sent) != 1 {
				t.Fatal("advanced past unresolved first record")
			}
			s.failMark = false
			p.fail = false
			if err := r.Drain(context.Background()); err != nil {
				t.Fatal(err)
			}
			if len(s.marked) != 2 || len(p.sent) != 3 || p.sent[0].EventID != p.sent[1].EventID || p.sent[2].ID != 2 {
				t.Fatal("replay lost stable ID or order")
			}
		})
	}
}

func (s *memoryStore) Reject(_ context.Context, id int64, _ string) error {
	if s.failReject {
		return errors.New("database down")
	}
	s.rejected = append(s.rejected, id)
	s.rows = s.rows[1:]
	return nil
}

func TestRejectionMustBeDurableBeforeRelayContinues(t *testing.T) {
	s := &memoryStore{rows: []Message{{ID: 1}, {ID: 2}}, failReject: true}
	p := &publisher{permanent: true}
	r := &Relay{Store: s, Publisher: p, BatchSize: 100, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := r.Drain(context.Background()); err == nil {
		t.Fatal("lost rejection hidden")
	}
	if len(s.rows) != 2 || len(s.rejected) != 0 || len(s.marked) != 0 || len(p.sent) != 1 {
		t.Fatal("advanced without durable rejection")
	}
}

func TestPermanentFailurePreservesEvidenceAndAllowsNextEvent(t *testing.T) {
	s := &memoryStore{rows: []Message{{ID: 1}, {ID: 2}}}
	p := &publisher{permanent: true}
	r := &Relay{Store: s, Publisher: p, BatchSize: 100, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := r.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(s.rejected) != 1 || s.rejected[0] != 1 || len(s.marked) != 1 || s.marked[0] != 2 {
		t.Fatal("rejection lost or incorrectly marked published")
	}
}
