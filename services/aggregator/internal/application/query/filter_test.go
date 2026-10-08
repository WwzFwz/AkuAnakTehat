package query

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	filter HazardFilter
}

func (f *fakeRepository) List(_ context.Context, filter HazardFilter) (HazardPage, error) {
	f.filter = filter
	return HazardPage{}, nil
}

func (f *fakeRepository) Get(context.Context, string) (HazardDetail, error) {
	return HazardDetail{}, nil
}

func TestNormalizeFilter(t *testing.T) {
	since := time.Date(2026, 9, 1, 0, 0, 0, 123456789, time.FixedZone("test", 7*60*60))
	repository := &fakeRepository{}
	service := Service{Repository: repository}
	_, err := service.List(context.Background(), HazardFilter{Since: &since})
	if err != nil {
		t.Fatal(err)
	}
	if repository.filter.Limit != DefaultLimit {
		t.Fatalf("limit=%d; want %d", repository.filter.Limit, DefaultLimit)
	}
	if repository.filter.Since == nil || repository.filter.Since.Location() != time.UTC {
		t.Fatal("since was not normalized to UTC")
	}
}

func TestNormalizeFilterRejectsInvalidValues(t *testing.T) {
	cases := []struct {
		name   string
		filter HazardFilter
		want   error
	}{
		{name: "type", filter: HazardFilter{Type: "OTHER"}, want: ErrInvalidType},
		{name: "severity", filter: HazardFilter{Severity: "CRITICAL"}, want: ErrInvalidSeverity},
		{name: "limit", filter: HazardFilter{Limit: MaxLimit + 1}, want: ErrInvalidLimit},
		{name: "zero since", filter: HazardFilter{Since: func() *time.Time { stamp := time.Time{}; return &stamp }()}, want: ErrInvalidSince},
		{name: "cursor", filter: HazardFilter{Cursor: &Cursor{OccurredAt: time.Now()}}, want: ErrInvalidCursor},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.filter.Normalize(); !errors.Is(err, tc.want) {
				t.Fatalf("error=%v; want %v", err, tc.want)
			}
		})
	}
}

func TestCursorRoundTrip(t *testing.T) {
	want := Cursor{OccurredAt: time.Date(2026, 9, 1, 0, 0, 0, 123456000, time.FixedZone("test", 7*60*60)), HazardID: "00000000-0000-0000-0000-000000000001"}
	encoded, err := EncodeCursor(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !got.OccurredAt.Equal(want.OccurredAt.UTC()) || got.HazardID != want.HazardID {
		t.Fatalf("cursor=%+v; want %+v", got, want)
	}
	if _, err = DecodeCursor("not-a-cursor"); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("error=%v; want invalid cursor", err)
	}
	trailing, err := EncodeCursor(want)
	if err != nil {
		t.Fatal(err)
	}
	trailing += "eyJ0cmFpbGluZyI6dHJ1ZX0"
	if _, err = DecodeCursor(trailing); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("trailing cursor error=%v; want invalid cursor", err)
	}
}

func TestServiceNormalizesEmptyResults(t *testing.T) {
	service := Service{Repository: &fakeRepository{}}
	page, err := service.List(context.Background(), HazardFilter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if page.Data == nil || page.Sources == nil {
		t.Fatal("empty query results must be non-nil slices")
	}
}
