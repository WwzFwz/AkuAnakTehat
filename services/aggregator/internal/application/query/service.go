package query

import (
	"context"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"fmt"
)

type Service struct {
	Repository HazardQuery
}

func (s Service) List(ctx context.Context, filter HazardFilter) (HazardPage, error) {
	normalized, err := filter.Normalize()
	if err != nil {
		return HazardPage{}, err
	}
	if s.Repository == nil {
		return HazardPage{}, ErrUnavailable
	}
	page, err := s.Repository.List(ctx, normalized)
	if err != nil {
		return HazardPage{}, classify(err)
	}
	if page.Data == nil {
		page.Data = []hazard.Event{}
	}
	if page.Sources == nil {
		page.Sources = []SourceStatus{}
	}
	return page, nil
}

func (s Service) Get(ctx context.Context, id string) (hazard.Event, error) {
	if !ValidHazardID(id) {
		return hazard.Event{}, ErrNotFound
	}
	if s.Repository == nil {
		return hazard.Event{}, ErrUnavailable
	}
	event, err := s.Repository.Get(ctx, id)
	if err != nil {
		return hazard.Event{}, classify(err)
	}
	return event, nil
}

func classify(err error) error {
	switch {
	case errors.Is(err, ErrInvalidType), errors.Is(err, ErrInvalidSeverity),
		errors.Is(err, ErrInvalidSince), errors.Is(err, ErrInvalidLimit),
		errors.Is(err, ErrInvalidCursor), errors.Is(err, ErrNotFound),
		errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, ErrUnavailable):
		return err
	default:
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
}
