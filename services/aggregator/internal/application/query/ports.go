package query

import (
	"context"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
)

type HazardQuery interface {
	List(context.Context, HazardFilter) (HazardPage, error)
	Get(context.Context, string) (hazard.Event, error)
}
