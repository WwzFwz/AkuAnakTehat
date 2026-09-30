package postgres

import (
	"context"
	"example.com/akuanaktehat/aggregator/internal/application/ingest"
	"github.com/jackc/pgx/v5"
	"time"
)

type transaction struct {
	tx  pgx.Tx
	ctx context.Context
}

func (t *transaction) bounded(ctx context.Context) (context.Context, context.CancelFunc) {
	deadline, _ := t.ctx.Deadline()
	return context.WithDeadline(ctx, deadline)
}

func (s *Store) WithTx(ctx context.Context, fn func(ingest.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	if err = fn(&transaction{tx: tx, ctx: ctx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
