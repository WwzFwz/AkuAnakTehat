package application

import (
	"context"
	"example.com/akuanaktehat/pemda-portal/internal/contract"
)

type View interface {
	Apply(context.Context, contract.Event) error
}
type Service struct{ Store View }

func (s *Service) Handle(ctx context.Context, e contract.Event) error { return s.Store.Apply(ctx, e) }
