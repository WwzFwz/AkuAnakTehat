package postgres

import (
	"context"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/observability"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Store struct {
	Pool    *pgxpool.Pool
	Timeout time.Duration
}

func Open(ctx context.Context, dsn string, max int32, timeout time.Duration) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("invalid database configuration")
	}
	cfg.ConnConfig.Tracer = observability.PostgresTracer{}
	cfg.MaxConns = max
	cfg.ConnConfig.ConnectTimeout = timeout
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprint(timeout.Milliseconds())
	cfg.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = fmt.Sprint(timeout.Milliseconds() * 2)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("database pool unavailable")
	}
	s := &Store{pool, timeout}
	if err = s.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("database unavailable")
	}
	return s, nil
}
func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.Pool.Ping(ctx)
}
