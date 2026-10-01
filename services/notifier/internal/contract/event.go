package contract

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

// Keep Raw unchanged: additive fields and numeric precision survive projection.
type Event struct {
	SchemaVersion int             `json:"schema_version"`
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	HazardID      string          `json:"hazard_id"`
	Version       int64           `json:"version"`
	CorrelationID string          `json:"correlation_id"`
	PublishedAt   time.Time       `json:"published_at"`
	Hazard        Hazard          `json:"hazard"`
	Raw           json.RawMessage `json:"-"`
}
type Hazard struct {
	ID          string                     `json:"hazard_id"`
	Source      string                     `json:"source"`
	SourceRefID string                     `json:"source_ref_id"`
	Type        string                     `json:"hazard_type"`
	Severity    string                     `json:"severity"`
	Area        string                     `json:"area_name"`
	Latitude    *float64                   `json:"latitude"`
	Longitude   *float64                   `json:"longitude"`
	OccurredAt  time.Time                  `json:"occurred_at"`
	IngestedAt  time.Time                  `json:"ingested_at"`
	Attributes  map[string]json.RawMessage `json:"attributes"`
}

var uuid = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var ErrInvalid = errors.New("invalid_event")

func Decode(key, raw []byte) (Event, error) {
	var e Event
	if len(raw) > 1<<20 || json.Unmarshal(raw, &e) != nil {
		return e, ErrInvalid
	}
	h := e.Hazard
	if e.SchemaVersion != 1 || e.EventType != "hazard.upserted" || !uuid.MatchString(e.EventID) || !uuid.MatchString(e.HazardID) || e.HazardID != h.ID || string(key) != e.HazardID || e.Version < 1 || e.CorrelationID == "" || e.PublishedAt.IsZero() {
		return e, ErrInvalid
	}
	if !((h.Source == "BMKG" && h.Type == "SEISMIC") || (h.Source == "PVMBG" && h.Type == "VOLCANIC")) || h.SourceRefID == "" || h.Area == "" || h.OccurredAt.IsZero() || h.IngestedAt.IsZero() || h.Attributes == nil {
		return e, ErrInvalid
	}
	if h.Latitude == nil || h.Longitude == nil || *h.Latitude < -90 || *h.Latitude > 90 || *h.Longitude < -180 || *h.Longitude > 180 {
		return e, ErrInvalid
	}
	switch h.Severity {
	case "NORMAL", "WASPADA", "SIAGA", "AWAS":
	default:
		return e, ErrInvalid
	}
	e.Raw = append(json.RawMessage(nil), raw...)
	return e, nil
}
