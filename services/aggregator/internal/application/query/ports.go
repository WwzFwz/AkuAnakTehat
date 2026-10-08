package query

import (
	"context"
)

type HazardQuery interface {
	List(context.Context, HazardFilter) (HazardPage, error)
	Get(context.Context, string) (HazardDetail, error)
}
