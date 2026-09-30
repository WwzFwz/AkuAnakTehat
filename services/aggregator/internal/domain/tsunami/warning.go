package tsunami

import (
	"encoding/json"
	"time"
)

type Warning struct {
	ID        string                     `json:"warning_id"`
	RelatedID string                     `json:"related_event_id"`
	Threat    string                     `json:"threat_level"`
	Zones     []string                   `json:"affected_zones"`
	Arrival   time.Time                  `json:"estimated_arrival"`
	Extra     map[string]json.RawMessage `json:"extra,omitempty"`
}

func Severity(level string) string {
	switch level {
	case "Normal":
		return "NORMAL"
	case "Waspada":
		return "WASPADA"
	case "Siaga":
		return "SIAGA"
	case "Awas":
		return "AWAS"
	}
	return ""
}
func Rank(level string) int {
	switch level {
	case "NORMAL":
		return 1
	case "WASPADA":
		return 2
	case "SIAGA":
		return 3
	case "AWAS":
		return 4
	}
	return 0
}
