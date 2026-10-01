package contract

import (
	"encoding/json"
	"fmt"
	"testing"
)

const sample = `{"schema_version":1,"event_id":"0199a100-0000-7000-8000-000000000001","event_type":"hazard.upserted","hazard_id":"0199a100-0000-7000-8000-000000000002","version":1,"correlation_id":"test","published_at":"2026-09-30T01:00:01Z","hazard":{"hazard_id":"0199a100-0000-7000-8000-000000000002","source":"BMKG","source_ref_id":"TEST-1","hazard_type":"SEISMIC","severity":"SIAGA","area_name":"Demo","latitude":0,"longitude":0,"occurred_at":"2026-09-30T01:00:00Z","ingested_at":"2026-09-30T01:00:01Z","attributes":{"precise":9007199254740993}},"future_field":{"active":true}}`

func TestContract(t *testing.T) {
	key := []byte("0199a100-0000-7000-8000-000000000002")
	e, err := Decode(key, []byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if string(e.Raw) != sample || string(e.Hazard.Attributes["precise"]) != "9007199254740993" {
		t.Fatal("additive data changed")
	}
	for name, mutate := range map[string]func(map[string]any){
		"schema": func(m map[string]any) { m["schema_version"] = 2 }, "version": func(m map[string]any) { m["version"] = 0 }, "event_id": func(m map[string]any) { m["event_id"] = "broken" }, "identity": func(m map[string]any) { m["hazard_id"] = "0199a100-0000-7000-8000-000000000003" }, "coordinates": func(m map[string]any) { delete(m["hazard"].(map[string]any), "latitude") }, "timestamp": func(m map[string]any) { m["published_at"] = "invalid" }, "severity": func(m map[string]any) { m["hazard"].(map[string]any)["severity"] = "OTHER" }, "attributes": func(m map[string]any) { m["hazard"].(map[string]any)["attributes"] = nil },
	} {
		t.Run(name, func(t *testing.T) {
			var m map[string]any
			_ = json.Unmarshal([]byte(sample), &m)
			mutate(m)
			raw, _ := json.Marshal(m)
			if _, err := Decode(key, raw); err == nil {
				t.Fatal("invalid contract accepted", fmt.Sprint(m))
			}
		})
	}
	if _, err = Decode([]byte("wrong-key"), []byte(sample)); err == nil {
		t.Fatal("key mismatch accepted")
	}
}
