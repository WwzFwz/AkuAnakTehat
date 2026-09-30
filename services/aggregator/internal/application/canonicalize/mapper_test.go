package canonicalize

import (
	"bytes"
	"encoding/json"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"example.com/akuanaktehat/aggregator/internal/domain/tsunami"
	"testing"
	"time"
)

func TestSeverityAndWarningRules(t *testing.T) {
	for _, tc := range []struct {
		m    float64
		want string
	}{{4.99, "NORMAL"}, {5, "WASPADA"}, {6.49, "WASPADA"}, {6.5, "SIAGA"}} {
		e := MapSeismic(SeismicInput{ID: "event", Magnitude: tc.m}, nil)
		if e.Severity != tc.want {
			t.Fatalf("magnitude %v: %s", tc.m, e.Severity)
		}
	}
	s := SeismicInput{ID: "event", Magnitude: 7, Potential: true}
	w := tsunami.Warning{ID: "warning", RelatedID: "event", Threat: "Waspada"}
	if e := MapSeismic(s, []tsunami.Warning{w}); e.Severity != "WASPADA" {
		t.Fatal("warning overrides magnitude; do not take maximum")
	}
	w2 := w
	w2.ID = "warning-2"
	w2.Threat = "Awas"
	a := MapSeismic(s, []tsunami.Warning{w, w2})
	b := MapSeismic(s, []tsunami.Warning{w2, w})
	ha, _ := hazard.ContentHash(a)
	hb, _ := hazard.ContentHash(b)
	if a.Severity != "AWAS" || !bytes.Equal(ha, hb) {
		t.Fatal("highest warning or deterministic ordering broken")
	}
	s.Potential = false
	if e := MapSeismic(s, []tsunami.Warning{w2}); e.Severity != "SIAGA" {
		t.Fatal("warning applied to non-tsunami event")
	}
}
func TestTolerantReaderAndSemanticHash(t *testing.T) {
	body := []byte(`[{"report_id":"r","volcano_id":"v","alert_level":"Siaga","eruption_count_24h":2,"ash_column_height_m":100,"reported_at":"2026-09-01T00:00:00.123456789Z","confidence_level":0.95,"future":{"list":[true,null,9007199254740993]}},{"report_id":"broken"},null]`)
	items, err := Decode(VolcanicEndpoint, body)
	if err != nil || len(items) != 3 || items[0].Reason != "" || items[1].Reason == "" || items[2].Reason == "" {
		t.Fatal("record isolation failed")
	}
	refs := map[string]Volcano{"v": {Name: "demo", Latitude: -7, Longitude: 110}}
	e, err := MapVolcanic(*items[0].Volcanic, refs)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(e.Attributes["future"], []byte("9007199254740993")) {
		t.Fatal("unknown number was rounded")
	}
	if e.OccurredAt.Nanosecond() != 123456000 {
		t.Fatal("canonical precision must match PostgreSQL microseconds")
	}
	h, err := hazard.ContentHash(e)
	if err != nil {
		t.Fatal(err)
	}
	e.ID = "different"
	e.IngestedAt = time.Now()
	e.Attributes["ash_column_height_m"] = json.RawMessage("1e2")
	same, err := hazard.ContentHash(e)
	if err != nil || !bytes.Equal(h, same) {
		t.Fatal("metadata or number formatting changed hash")
	}
	e.Attributes["confidence_level"] = json.RawMessage("0.9")
	changed, _ := hazard.ContentHash(e)
	if bytes.Equal(h, changed) {
		t.Fatal("business changes must change hash")
	}
	if _, err = Decode(VolcanicEndpoint, []byte(`{"data":[]}`)); err == nil {
		t.Fatal("invalid envelope accepted")
	}
	if _, err = MapVolcanic(*items[0].Volcanic, nil); err == nil {
		t.Fatal("unknown volcano accepted")
	}
}
