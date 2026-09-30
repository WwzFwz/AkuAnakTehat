package postgres

import (
	"context"
	"database/sql"
	"errors"
	"example.com/akuanaktehat/aggregator/migrations"
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"time"
)

func Migrate(ctx context.Context, dsn string, timeout time.Duration) error {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return errors.New("invalid migration configuration")
	}
	cfg.ConnectTimeout = timeout
	cfg.RuntimeParams["statement_timeout"] = "30000"
	cfg.RuntimeParams["lock_timeout"] = "10000"
	var db *sql.DB = stdlib.OpenDB(*cfg)
	defer db.Close()
	db.SetMaxOpenConns(1)
	check, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if db.PingContext(check) != nil {
		return errors.New("migration database unavailable")
	}
	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return errors.New("migration driver unavailable")
	}
	src, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", src, "pgx5", driver)
	if err != nil {
		return err
	}
	defer m.Close()
	m.LockTimeout = 10 * time.Second
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}
