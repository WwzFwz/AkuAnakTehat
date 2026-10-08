package query

import (
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"time"
)

type SourceStatus struct {
	Source     string     `json:"source"`
	Status     string     `json:"status"`
	StaleSince *time.Time `json:"stale_since,omitempty"`
}

type HazardPage struct {
	Data       []hazard.Event `json:"data"`
	NextCursor string         `json:"next_cursor,omitempty"`
	Sources    []SourceStatus `json:"sources"`
}

// Detail preserves the existing hazard fields and adds the same source metadata as a list.
type HazardDetail struct {
	hazard.Event
	Sources []SourceStatus `json:"sources"`
}
