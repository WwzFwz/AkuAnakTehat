package canonicalize

import (
	"bytes"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"example.com/akuanaktehat/aggregator/internal/domain/tsunami"
	"fmt"
	"sort"
	"time"
)

const SeismicEndpoint = "bmkg.seismic-events"
const WarningEndpoint = "bmkg.tsunami-warnings"
const VolcanicEndpoint = "pvmbg.volcanic-reports"

type SeismicInput struct {
	ID                                    string
	Magnitude, Depth, Latitude, Longitude float64
	Area                                  string
	OccurredAt                            time.Time
	Potential                             bool
	Extra                                 map[string]json.RawMessage
}
type VolcanicInput struct {
	ID, VolcanoID, Alert string
	Eruptions            int
	Ash                  float64
	ReportedAt           time.Time
	Extra                map[string]json.RawMessage
}
type Volcano struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type Item struct {
	Raw      json.RawMessage
	Seismic  *SeismicInput
	Warning  *tsunami.Warning
	Volcanic *VolcanicInput
	Reason   string
	Unknown  []string
}
type Batch struct {
	Endpoint, CorrelationID string
	Watermark               time.Time
	Items                   []Item
}

func Source(endpoint string) string {
	switch endpoint {
	case SeismicEndpoint, WarningEndpoint:
		return "BMKG"
	case VolcanicEndpoint:
		return "PVMBG"
	}
	return ""
}

// Decode isolates invalid records while preserving unknown JSON values.
func Decode(endpoint string, body []byte) ([]Item, error) {
	if len(bytes.TrimSpace(body)) == 0 || bytes.TrimSpace(body)[0] != '[' {
		return nil, errors.New("expected JSON array")
	}
	var raws []json.RawMessage
	if json.Unmarshal(body, &raws) != nil {
		return nil, errors.New("malformed source JSON")
	}
	items := make([]Item, 0, len(raws))
	for _, raw := range raws {
		item := Item{Raw: raw}
		fields := map[string]json.RawMessage{}
		if json.Unmarshal(raw, &fields) != nil || fields == nil {
			item.Reason = "record must be an object"
			items = append(items, item)
			continue
		}
		// Validate JSONB-representable strings/numbers before typed decoding.
		// Unknown fields remain raw; this only bounds values PostgreSQL cannot store.
		if _, err := hazard.ContentHash(hazard.Event{Attributes: fields}); err != nil {
			item.Reason = "source value outside supported JSONB range"
			items = append(items, item)
			continue
		}
		take := func(key string, dst any) {
			v, ok := fields[key]
			if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) || json.Unmarshal(v, dst) != nil {
				if item.Reason == "" {
					item.Reason = "invalid or missing " + key
				}
			}
			delete(fields, key)
		}
		switch endpoint {
		case SeismicEndpoint:
			s := &SeismicInput{}
			take("event_id", &s.ID)
			take("magnitude", &s.Magnitude)
			take("depth_km", &s.Depth)
			take("epicenter_lat", &s.Latitude)
			take("epicenter_lon", &s.Longitude)
			take("region_name", &s.Area)
			take("occurred_at", &s.OccurredAt)
			take("potential_tsunami", &s.Potential)
			s.OccurredAt = s.OccurredAt.UTC().Truncate(time.Microsecond)
			s.Extra = fields
			item.Seismic = s
			if s.ID == "" || s.Area == "" || s.OccurredAt.IsZero() || s.Magnitude < 0 || s.Magnitude > 10 || s.Depth < 0 || s.Latitude < -90 || s.Latitude > 90 || s.Longitude < -180 || s.Longitude > 180 {
				item.Reason = "seismic value out of range"
			}
		case WarningEndpoint:
			w := &tsunami.Warning{}
			take("warning_id", &w.ID)
			take("related_event_id", &w.RelatedID)
			take("threat_level", &w.Threat)
			take("affected_zones", &w.Zones)
			take("estimated_arrival", &w.Arrival)
			w.Arrival = w.Arrival.UTC()
			w.Extra = fields
			item.Warning = w
			if w.ID == "" || w.RelatedID == "" || w.Arrival.IsZero() || len(w.Zones) == 0 || tsunami.Rank(tsunami.Severity(w.Threat)) < 2 {
				item.Reason = "invalid tsunami warning"
			}
		case VolcanicEndpoint:
			v := &VolcanicInput{}
			take("report_id", &v.ID)
			take("volcano_id", &v.VolcanoID)
			take("alert_level", &v.Alert)
			take("eruption_count_24h", &v.Eruptions)
			take("ash_column_height_m", &v.Ash)
			take("reported_at", &v.ReportedAt)
			v.ReportedAt = v.ReportedAt.UTC().Truncate(time.Microsecond)
			v.Extra = fields
			item.Volcanic = v
			if v.ID == "" || v.VolcanoID == "" || v.ReportedAt.IsZero() || v.Eruptions < 0 || v.Ash < 0 || tsunami.Severity(v.Alert) == "" {
				item.Reason = "invalid volcanic report"
			}
		default:
			return nil, fmt.Errorf("unknown endpoint")
		}
		for k := range fields {
			item.Unknown = append(item.Unknown, k)
		}
		sort.Strings(item.Unknown)
		items = append(items, item)
	}
	return items, nil
}
