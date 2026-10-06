package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/application/query"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"strings"
)

func (s *Store) List(ctx context.Context, filter query.HazardFilter) (query.HazardPage, error) {
	var err error
	filter, err = filter.Normalize()
	if err != nil {
		return query.HazardPage{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	args := make([]any, 0, 8)
	arg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	var sql strings.Builder
	sql.WriteString(`SELECT hazard_id::text,source,source_ref_id,hazard_type,severity,area_name,latitude,longitude,occurred_at,ingested_at,attributes FROM hazard_events WHERE true`)
	if filter.Type != "" {
		sql.WriteString(" AND hazard_type = ")
		sql.WriteString(arg(filter.Type))
	}
	if filter.Severity != "" {
		sql.WriteString(" AND severity = ")
		sql.WriteString(arg(filter.Severity))
	}
	if filter.Since != nil {
		sql.WriteString(" AND occurred_at >= ")
		sql.WriteString(arg(*filter.Since))
	}
	if filter.Cursor != nil {
		stamp := arg(filter.Cursor.OccurredAt)
		id := arg(filter.Cursor.HazardID)
		sql.WriteString(" AND (occurred_at < ")
		sql.WriteString(stamp)
		sql.WriteString(" OR (occurred_at = ")
		sql.WriteString(stamp)
		sql.WriteString(" AND hazard_id < ")
		sql.WriteString(id)
		sql.WriteString("::uuid))")
	}
	sql.WriteString(" ORDER BY occurred_at DESC, hazard_id DESC LIMIT ")
	sql.WriteString(arg(filter.Limit + 1))

	rows, err := s.Pool.Query(ctx, sql.String(), args...)
	if err != nil {
		return query.HazardPage{}, err
	}
	defer rows.Close()

	data := make([]hazard.Event, 0, filter.Limit)
	for rows.Next() {
		event, scanErr := scanEvent(rows)
		if scanErr != nil {
			return query.HazardPage{}, scanErr
		}
		data = append(data, event)
	}
	if err = rows.Err(); err != nil {
		return query.HazardPage{}, err
	}

	page := query.HazardPage{Data: data}
	if len(page.Data) > filter.Limit {
		last := page.Data[filter.Limit-1]
		page.Data = page.Data[:filter.Limit]
		page.NextCursor, err = query.EncodeCursor(query.Cursor{OccurredAt: last.OccurredAt, HazardID: last.ID})
		if err != nil {
			return query.HazardPage{}, err
		}
	}
	page.Sources, err = s.sourceStatuses(ctx, filter.Type)
	if err != nil {
		return query.HazardPage{}, err
	}
	return page, nil
}

func (s *Store) Get(ctx context.Context, id string) (hazard.Event, error) {
	if !query.ValidHazardID(id) {
		return hazard.Event{}, query.ErrNotFound
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	row := s.Pool.QueryRow(ctx, `SELECT hazard_id::text,source,source_ref_id,hazard_type,severity,area_name,latitude,longitude,occurred_at,ingested_at,attributes FROM hazard_events WHERE hazard_id=$1::uuid`, id)
	event, err := scanEvent(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return hazard.Event{}, query.ErrNotFound
	}
	return event, err
}

func scanEvent(row interface{ Scan(...any) error }) (hazard.Event, error) {
	var event hazard.Event
	var attributes []byte
	err := row.Scan(&event.ID, &event.Source, &event.SourceRefID, &event.Type, &event.Severity, &event.Area, &event.Latitude, &event.Longitude, &event.OccurredAt, &event.IngestedAt, &attributes)
	if err != nil {
		return hazard.Event{}, err
	}
	event.OccurredAt = event.OccurredAt.UTC()
	event.IngestedAt = event.IngestedAt.UTC()
	if len(attributes) == 0 {
		event.Attributes = map[string]json.RawMessage{}
		return event, nil
	}
	if err = json.Unmarshal(attributes, &event.Attributes); err != nil {
		return hazard.Event{}, err
	}
	if event.Attributes == nil {
		event.Attributes = map[string]json.RawMessage{}
	}
	return event, nil
}

func (s *Store) sourceStatuses(ctx context.Context, hazardType string) ([]query.SourceStatus, error) {
	rows, err := s.Pool.Query(ctx, `SELECT source,status,stale_since FROM source_status WHERE source IN ('BMKG','PVMBG')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bySource := map[string]query.SourceStatus{}
	for rows.Next() {
		var source, status string
		var stale pgtype.Timestamptz
		if err = rows.Scan(&source, &status, &stale); err != nil {
			return nil, err
		}
		item := query.SourceStatus{Source: source, Status: status}
		if stale.Valid {
			stamp := stale.Time.UTC()
			item.StaleSince = &stamp
		}
		bySource[source] = item
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	sources := []string{"BMKG", "PVMBG"}
	switch hazardType {
	case "SEISMIC":
		sources = []string{"BMKG"}
	case "VOLCANIC":
		sources = []string{"PVMBG"}
	}
	result := make([]query.SourceStatus, 0, len(sources))
	for _, source := range sources {
		if status, ok := bySource[source]; ok {
			result = append(result, status)
		}
	}
	return result, nil
}
