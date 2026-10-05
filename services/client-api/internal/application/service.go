package application

import (
	"context"
	"errors"
	"example.com/akuanaktehat/client-api/internal/authz"
	"example.com/akuanaktehat/client-api/internal/projection"
	"net/url"
)

var ErrUnavailable = errors.New("dependency_unavailable")
var ErrNotFound = errors.New("not_found")
var ErrQuery = errors.New("invalid_query")
var ErrInvalidType = errors.New("invalid_type")
var ErrInvalidSeverity = errors.New("invalid_severity")
var ErrInvalidSince = errors.New("invalid_since")
var ErrInvalidLimit = errors.New("invalid_limit")
var ErrInvalidCursor = errors.New("invalid_cursor")
var ErrOverloaded = errors.New("upstream_overloaded")

type Source struct {
	Source     string `json:"source"`
	Status     string `json:"status"`
	StaleSince string `json:"stale_since,omitempty"`
}
type Page struct {
	Data       []map[string]any `json:"data"`
	NextCursor string           `json:"next_cursor,omitempty"`
	Sources    []Source         `json:"sources"`
}
type Aggregator interface {
	List(context.Context, url.Values) (Page, error)
	Get(context.Context, string) (map[string]any, error)
	Ready(context.Context) error
}
type Service struct{ Upstream Aggregator }

func (s Service) List(ctx context.Context, scope string, query url.Values, fields []string, raw bool) (Page, error) {
	if err := authz.Allowed(scope, fields, raw); err != nil {
		return Page{}, err
	}
	p, err := s.Upstream.List(ctx, query)
	if err != nil {
		return Page{}, err
	}
	result := Page{NextCursor: p.NextCursor, Sources: p.Sources, Data: make([]map[string]any, 0, len(p.Data))}
	for _, h := range p.Data {
		result.Data = append(result.Data, projection.Project(h, scope, fields))
	}
	return result, nil
}
func (s Service) Get(ctx context.Context, scope, id string, fields []string, raw bool) (map[string]any, error) {
	if err := authz.Allowed(scope, fields, raw); err != nil {
		return nil, err
	}
	h, err := s.Upstream.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return projection.Project(h, scope, fields), nil
}
