package canonicalize

import (
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"example.com/akuanaktehat/aggregator/internal/domain/tsunami"
	"sort"
)

func raw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func extras(src map[string]json.RawMessage) map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	reserved := map[string]json.RawMessage{}
	for k, v := range src {
		if k == "tsunami_warnings" || k == "source_extensions" {
			reserved[k] = v
		} else {
			out[k] = v
		}
	}
	if len(reserved) > 0 {
		out["source_extensions"] = raw(reserved)
	}
	return out
}
func MapSeismic(s SeismicInput, warnings []tsunami.Warning) hazard.Event {
	a := extras(s.Extra)
	a["magnitude"] = raw(s.Magnitude)
	a["depth_km"] = raw(s.Depth)
	a["potential_tsunami"] = raw(s.Potential)
	e := hazard.Event{Source: "BMKG", SourceRefID: s.ID, Type: "SEISMIC", Area: s.Area, Latitude: s.Latitude, Longitude: s.Longitude, OccurredAt: s.OccurredAt, Attributes: a}
	Correlate(&e, warnings)
	return e
}
func Correlate(e *hazard.Event, warnings []tsunami.Warning) {
	var magnitude float64
	var potential bool
	json.Unmarshal(e.Attributes["magnitude"], &magnitude)
	json.Unmarshal(e.Attributes["potential_tsunami"], &potential)
	e.Severity = "NORMAL"
	if magnitude >= 6.5 {
		e.Severity = "SIAGA"
	} else if magnitude >= 5 {
		e.Severity = "WASPADA"
	}
	delete(e.Attributes, "tsunami_warnings")
	if !potential {
		return
	}
	relevant := []tsunami.Warning{}
	for _, w := range warnings {
		if w.RelatedID == e.SourceRefID {
			relevant = append(relevant, w)
		}
	}
	if len(relevant) == 0 {
		return
	}
	sort.Slice(relevant, func(i, j int) bool { return relevant[i].ID < relevant[j].ID })
	level := ""
	payloads := []map[string]json.RawMessage{}
	for _, w := range relevant {
		if tsunami.Rank(tsunami.Severity(w.Threat)) > tsunami.Rank(level) {
			level = tsunami.Severity(w.Threat)
		}
		a := extras(w.Extra)
		a["warning_id"] = raw(w.ID)
		a["related_event_id"] = raw(w.RelatedID)
		a["threat_level"] = raw(w.Threat)
		zones := append([]string{}, w.Zones...)
		sort.Strings(zones)
		a["affected_zones"] = raw(zones)
		a["estimated_arrival"] = raw(w.Arrival)
		payloads = append(payloads, a)
	}
	e.Severity = level
	e.Attributes["tsunami_warnings"] = raw(payloads)
}
func MapVolcanic(v VolcanicInput, reference map[string]Volcano) (hazard.Event, error) {
	r, ok := reference[v.VolcanoID]
	if !ok {
		return hazard.Event{}, errors.New("unknown volcano_id")
	}
	a := extras(v.Extra)
	a["volcano_id"] = raw(v.VolcanoID)
	a["alert_level"] = raw(v.Alert)
	a["eruption_count_24h"] = raw(v.Eruptions)
	a["ash_column_height_m"] = raw(v.Ash)
	return hazard.Event{Source: "PVMBG", SourceRefID: v.ID, Type: "VOLCANIC", Severity: tsunami.Severity(v.Alert), Area: r.Name, Latitude: r.Latitude, Longitude: r.Longitude, OccurredAt: v.ReportedAt, Attributes: a}, nil
}
