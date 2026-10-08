package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/pemda-portal/internal/contract"
	"github.com/twmb/franz-go/pkg/kgo"
	"strings"
	"testing"
	"time"
)

const sample = `{"schema_version":1,"event_id":"0199a100-0000-7000-8000-000000000001","event_type":"hazard.upserted","hazard_id":"0199a100-0000-7000-8000-000000000002","version":1,"correlation_id":"test","published_at":"2026-09-30T01:00:01Z","hazard":{"hazard_id":"0199a100-0000-7000-8000-000000000002","source":"BMKG","source_ref_id":"TEST-1","hazard_type":"SEISMIC","severity":"SIAGA","area_name":"Demo","latitude":0,"longitude":0,"occurred_at":"2026-09-30T01:00:00Z","ingested_at":"2026-09-30T01:00:01Z","attributes":{"precise":9007199254740993}},"future_field":{"active":true}}`

type processStub struct{ calls, failures int }

func (p *processStub) Handle(context.Context, contract.Event) error {
	p.calls++
	if p.calls <= p.failures {
		return errors.New("store unavailable")
	}
	return nil
}

type deliveryStub struct {
	dead, committed, attempts int
	failDLQ, failCommit       bool
}

func (d *deliveryStub) DeadLetter(_ context.Context, _ *kgo.Record, _ string, n int) error {
	d.dead++
	d.attempts = n
	if d.failDLQ {
		return errors.New("no DLQ ACK")
	}
	return nil
}
func (d *deliveryStub) Commit(context.Context, *kgo.Record) error {
	d.committed++
	if d.failCommit {
		return errors.New("no offset ACK")
	}
	return nil
}
func TestProcessingBoundaries(t *testing.T) {
	cases := []struct {
		name                   string
		invalid                bool
		failures               int
		failDLQ, failCommit    bool
		calls, dead, committed int
		wantErr                bool
	}{
		{"success", false, 0, false, false, 1, 0, 1, false}, {"retry succeeds", false, 1, false, false, 2, 0, 1, false}, {"exhausted", false, 9, false, false, 3, 0, 0, true}, {"poison", true, 0, false, false, 0, 1, 1, false}, {"DLQ unavailable", true, 0, true, false, 0, 1, 0, true}, {"offset unavailable", false, 0, false, true, 1, 0, 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &kgo.Record{Key: []byte("0199a100-0000-7000-8000-000000000002"), Value: []byte(sample)}
			if tc.invalid {
				r.Value = []byte("invalid JSON")
			}
			p := &processStub{failures: tc.failures}
			d := &deliveryStub{failDLQ: tc.failDLQ, failCommit: tc.failCommit}
			err := Process(context.Background(), r, p, d, 3, time.Second)
			if (err != nil) != tc.wantErr || p.calls != tc.calls || d.dead != tc.dead || d.committed != tc.committed {
				t.Fatalf("unsafe boundary: p=%+v d=%+v error=%v", p, d, err)
			}
			if tc.failures > 3 && p.calls != 3 {
				t.Fatal("wrong attempt count")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d := &deliveryStub{}
	if Process(ctx, &kgo.Record{}, &processStub{}, d, 3, time.Second) == nil || d.committed != 0 || d.dead != 0 {
		t.Fatal("shutdown committed unfinished record")
	}
}
func TestDLQMetadata(t *testing.T) {
	original := &kgo.Record{Topic: "source", Partition: 2, Offset: 42, Key: []byte("bad"), Value: []byte("{broken"), Headers: []kgo.RecordHeader{{Key: "correlation_id", Value: []byte("trace")}}}
	r := Record(original, "dead", "group", "invalid_event", 1)
	h := map[string]string{}
	for _, header := range r.Headers {
		h[header.Key] = string(header.Value)
	}
	if string(r.Value) != string(original.Value) || r.Topic != "dead" || h["consumer_group"] != "group" || h["source_partition"] != "2" || h["source_offset"] != "42" || h["source_topic"] != "source" || h["failure_reason"] != "invalid_event" || h["attempts"] != "1" || h["correlation_id"] != "trace" {
		t.Fatal("DLQ lost original payload or context")
	}
}

func TestRetryAfterDependencyRecovery(t *testing.T) {
	r := &kgo.Record{Key: []byte("0199a100-0000-7000-8000-000000000002"), Value: []byte(sample)}
	p := &processStub{failures: 3}
	d := &deliveryStub{}
	if Process(context.Background(), r, p, d, 3, time.Second) == nil {
		t.Fatal("dependency failure hidden")
	}
	if d.dead != 0 || d.committed != 0 {
		t.Fatal("valid record abandoned during dependency outage")
	}
	if err := Process(context.Background(), r, p, d, 3, time.Second); err != nil {
		t.Fatal(err)
	}
	if d.committed != 1 || d.dead != 0 {
		t.Fatal("recovery did not complete original record")
	}
}
func TestEnvelopeSizeBoundaries(t *testing.T) {
	var e map[string]any
	if err := json.Unmarshal([]byte(sample), &e); err != nil {
		t.Fatal(err)
	}
	e["padding"] = ""
	base, _ := json.Marshal(e)
	for _, delta := range []int{-1, 0, 1} {
		e["padding"] = strings.Repeat("x", (4<<20)-len(base)+delta)
		raw, _ := json.Marshal(e)
		_, err := contract.Decode([]byte("0199a100-0000-7000-8000-000000000002"), raw)
		if (err != nil) != (delta > 0) {
			t.Fatalf("bytes=%d error=%v", len(raw), err)
		}
	}
}
