package httpapi

import (
	"example.com/akuanaktehat/aggregator/internal/application/query"
	"net/url"
	"strconv"
	"time"
)

func parseFilter(values url.Values) (query.HazardFilter, error) {
	for key, entries := range values {
		if len(entries) != 1 {
			return query.HazardFilter{}, query.ErrInvalidQuery
		}
		switch key {
		case "type", "severity", "since", "limit", "cursor":
		default:
			return query.HazardFilter{}, query.ErrInvalidQuery
		}
	}

	filter := query.HazardFilter{
		Type:     values.Get("type"),
		Severity: values.Get("severity"),
	}
	if values.Has("type") && filter.Type == "" {
		return query.HazardFilter{}, query.ErrInvalidType
	}
	if values.Has("severity") && filter.Severity == "" {
		return query.HazardFilter{}, query.ErrInvalidSeverity
	}
	if values.Has("limit") {
		limit, err := strconv.Atoi(values.Get("limit"))
		if err != nil {
			return query.HazardFilter{}, query.ErrInvalidLimit
		}
		filter.Limit = limit
	}
	if values.Has("since") {
		stamp, err := time.Parse(time.RFC3339, values.Get("since"))
		if err != nil {
			return query.HazardFilter{}, query.ErrInvalidSince
		}
		filter.Since = &stamp
	}
	if values.Has("cursor") {
		cursor, err := query.DecodeCursor(values.Get("cursor"))
		if err != nil {
			return query.HazardFilter{}, err
		}
		filter.Cursor = &cursor
	}
	return filter, nil
}
