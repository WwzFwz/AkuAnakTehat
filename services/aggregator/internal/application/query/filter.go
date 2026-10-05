package query

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

const (
	DefaultLimit = 100
	MaxLimit     = 500
	MaxCursorLen = 2048
)

var (
	ErrInvalidType     = errors.New("invalid_type")
	ErrInvalidSeverity = errors.New("invalid_severity")
	ErrInvalidSince    = errors.New("invalid_since")
	ErrInvalidLimit    = errors.New("invalid_limit")
	ErrInvalidCursor   = errors.New("invalid_cursor")
	ErrNotFound        = errors.New("not_found")
	ErrUnavailable     = errors.New("dependency_unavailable")
)

type Cursor struct {
	OccurredAt time.Time
	HazardID   string
}

type HazardFilter struct {
	Type     string
	Severity string
	Since    *time.Time
	Limit    int
	Cursor   *Cursor
}

func (f HazardFilter) Normalize() (HazardFilter, error) {
	if f.Type != "" && f.Type != "SEISMIC" && f.Type != "VOLCANIC" {
		return HazardFilter{}, ErrInvalidType
	}
	if f.Severity != "" && !validSeverity(f.Severity) {
		return HazardFilter{}, ErrInvalidSeverity
	}
	if f.Limit == 0 {
		f.Limit = DefaultLimit
	}
	if f.Limit < 1 || f.Limit > MaxLimit {
		return HazardFilter{}, ErrInvalidLimit
	}
	if f.Since != nil {
		if f.Since.IsZero() {
			return HazardFilter{}, ErrInvalidSince
		}
		since := f.Since.UTC()
		f.Since = &since
	}
	if f.Cursor != nil {
		cursor := *f.Cursor
		if !validCursor(cursor) {
			return HazardFilter{}, ErrInvalidCursor
		}
		cursor.OccurredAt = cursor.OccurredAt.UTC()
		f.Cursor = &cursor
	}
	return f, nil
}

func EncodeCursor(cursor Cursor) (string, error) {
	if !validCursor(cursor) {
		return "", ErrInvalidCursor
	}
	payload := struct {
		OccurredAt time.Time `json:"occurred_at"`
		HazardID   string    `json:"hazard_id"`
	}{OccurredAt: cursor.OccurredAt.UTC(), HazardID: cursor.HazardID}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", ErrInvalidCursor
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func DecodeCursor(raw string) (Cursor, error) {
	if raw == "" || len(raw) > MaxCursorLen {
		return Cursor{}, ErrInvalidCursor
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	var payload struct {
		OccurredAt time.Time `json:"occurred_at"`
		HazardID   string    `json:"hazard_id"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	if err = decoder.Decode(&payload); err != nil || !validCursor(Cursor{OccurredAt: payload.OccurredAt, HazardID: payload.HazardID}) {
		return Cursor{}, ErrInvalidCursor
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return Cursor{}, ErrInvalidCursor
	}
	return Cursor{OccurredAt: payload.OccurredAt.UTC(), HazardID: payload.HazardID}, nil
}

func ValidHazardID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for i, r := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}
			continue
		}
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func validCursor(cursor Cursor) bool {
	return !cursor.OccurredAt.IsZero() && ValidHazardID(cursor.HazardID)
}

func validSeverity(value string) bool {
	switch value {
	case "NORMAL", "WASPADA", "SIAGA", "AWAS":
		return true
	default:
		return false
	}
}
