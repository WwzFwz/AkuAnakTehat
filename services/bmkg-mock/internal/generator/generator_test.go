package generator

import (
	"testing"
	"time"

	"example.com/akuanaktehat/bmkg-mock/internal/store"
)

func TestWarningOrderingAndEscalationWatermark(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	data := store.New()
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	for n := 1; n <= 4; n++ {
		g.Tick(data, start.Add(time.Duration(n)*time.Second))
	}
	warnings := data.ListWarningsSince(time.Time{})
	if len(warnings) != 2 {
		t.Fatalf("got %d warnings; want early and late scenarios", len(warnings))
	}
	events := map[string]bool{}
	for _, e := range data.ListSeismicSince(time.Time{}) {
		events[e.EventID] = true
	}
	var lateID, earlyEvent string
	for _, w := range warnings {
		if events[w.RelatedEventID] {
			lateID = w.WarningID
		} else {
			earlyEvent = w.RelatedEventID
		}
	}
	if lateID == "" || earlyEvent == "" {
		t.Fatal("expected warning both before and after its event")
	}
	watermark := start.Add(5 * time.Second)
	g.Tick(data, watermark)
	updates := data.ListWarningsSince(watermark)
	if len(updates) != 1 || updates[0].WarningID != lateID || updates[0].ThreatLevel != "Awas" {
		t.Fatal("changed warning must retain identity and be returned at inclusive modification watermark")
	}
	if len(data.ListWarningsSince(watermark.Add(time.Nanosecond))) != 0 {
		t.Fatal("unchanged warnings must not pass newer watermark")
	}
	found := false
	for _, e := range data.ListSeismicSince(watermark) {
		if e.EventID == earlyEvent {
			found = true
		}
	}
	if !found {
		t.Fatal("early warning's event must arrive on following tick")
	}
	if !updates[0].EstimatedArrival.After(watermark) {
		t.Fatal("fixture requires arrival later than modification")
	}
}
