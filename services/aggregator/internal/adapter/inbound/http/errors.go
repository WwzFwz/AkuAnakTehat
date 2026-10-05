package httpapi

import (
	"context"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/application/query"
	"net/http"
)

type errorResponse struct {
	Error         string `json:"error"`
	Message       string `json:"message"`
	CorrelationID string `json:"correlation_id"`
}

func classifyError(err error) (int, string, string) {
	switch {
	case errors.Is(err, query.ErrInvalidQuery):
		return http.StatusBadRequest, "invalid_query", "query is invalid"
	case errors.Is(err, query.ErrInvalidType):
		return http.StatusBadRequest, "invalid_type", "type is invalid"
	case errors.Is(err, query.ErrInvalidSeverity):
		return http.StatusBadRequest, "invalid_severity", "severity is invalid"
	case errors.Is(err, query.ErrInvalidSince):
		return http.StatusBadRequest, "invalid_since", "since is invalid"
	case errors.Is(err, query.ErrInvalidLimit):
		return http.StatusBadRequest, "invalid_limit", "limit is invalid"
	case errors.Is(err, query.ErrInvalidCursor):
		return http.StatusBadRequest, "invalid_cursor", "cursor is invalid"
	case errors.Is(err, query.ErrNotFound):
		return http.StatusNotFound, "not_found", "hazard was not found"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded), errors.Is(err, query.ErrUnavailable):
		return http.StatusServiceUnavailable, "dependency_unavailable", "query dependency is unavailable"
	default:
		return http.StatusServiceUnavailable, "dependency_unavailable", "query dependency is unavailable"
	}
}
