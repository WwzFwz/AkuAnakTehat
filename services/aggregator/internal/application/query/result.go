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
