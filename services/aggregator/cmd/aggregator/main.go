package main

import (
	"context"
	"errors"
	httpapi "example.com/akuanaktehat/aggregator/internal/adapter/inbound/http"
	"example.com/akuanaktehat/aggregator/internal/adapter/outbound/bmkg"
	kafkaproducer "example.com/akuanaktehat/aggregator/internal/adapter/outbound/kafka"
	"example.com/akuanaktehat/aggregator/internal/adapter/outbound/postgres"
	"example.com/akuanaktehat/aggregator/internal/adapter/outbound/pvmbg"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"example.com/akuanaktehat/aggregator/internal/application/ingest"
	queryapp "example.com/akuanaktehat/aggregator/internal/application/query"
	"example.com/akuanaktehat/aggregator/internal/config"
	"example.com/akuanaktehat/aggregator/internal/observability"
	"example.com/akuanaktehat/aggregator/internal/worker/outbox"
	"example.com/akuanaktehat/aggregator/internal/worker/poller"
	"example.com/akuanaktehat/aggregator/reference"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "aggregator")
	slog.SetDefault(logger)
	if err := run(logger); err != nil {
		logger.Error("aggregator_stopped", "reason", err.Error())
		os.Exit(1)
	}
}
func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err = postgres.Migrate(ctx, cfg.DatabaseURL, cfg.DBTimeout); err != nil {
		return errors.New("migration failed")
	}
	store, err := postgres.Open(ctx, cfg.DatabaseURL, cfg.PoolSize, cfg.DBTimeout)
	if err != nil {
		return err
	}
	defer store.Pool.Close()
	relayStore, err := postgres.Open(ctx, cfg.DatabaseURL, 2, cfg.DBTimeout)
	if err != nil {
		return err
	}
	defer relayStore.Pool.Close()
	queryStore, err := postgres.Open(ctx, cfg.DatabaseURL, cfg.QueryPoolSize, cfg.DBTimeout)
	if err != nil {
		return err
	}
	defer queryStore.Pool.Close()
	producer, err := kafkaproducer.New(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.PublishTimeout)
	if err != nil {
		return err
	}
	defer producer.Close()
	refs, err := reference.Load()
	if err != nil {
		return errors.New("invalid volcano reference")
	}
	service := &ingest.Service{UOW: store, Volcanoes: refs}
	bm := bmkg.New(cfg.BMKGURL, cfg.BMKGKey, cfg.BMKGTimeout)
	defer bm.Close()
	pv := pvmbg.New(cfg.PVMBGURL, cfg.PVMBGToken, cfg.PVMBGTimeout)
	defer pv.Close()
	endpoint := func(name string, fetch poller.Fetch) *poller.Endpoint {
		return &poller.Endpoint{Name: name, Fetch: fetch, Breaker: poller.Breaker{Threshold: cfg.BreakerFailures, Cooldown: cfg.BreakerCooldown}}
	}
	workers := []*poller.Worker{{Endpoints: []*poller.Endpoint{endpoint(canonicalize.SeismicEndpoint, bm.FetchSeismic), endpoint(canonicalize.WarningEndpoint, bm.FetchWarnings)}, Interval: cfg.BMKGInterval}, {Endpoints: []*poller.Endpoint{endpoint(canonicalize.VolcanicEndpoint, pv.FetchReports)}, Interval: cfg.PVMBGInterval}}
	var wg sync.WaitGroup
	relay := &outbox.Relay{Store: relayStore, Publisher: producer, Interval: cfg.OutboxInterval, Retention: cfg.OutboxRetention, BatchSize: 100, Logger: logger}
	wg.Add(1)
	go func() { defer wg.Done(); relay.Run(ctx) }()
	for _, w := range workers {
		w.Checkpoints = store
		w.Status = store
		w.Ingest = service
		w.Overlap = cfg.Overlap
		w.Logger = logger
		wg.Add(1)
		go func(w *poller.Worker) { defer wg.Done(); w.Run(ctx) }(w)
	}
	queryService := queryapp.Service{Repository: queryStore}
	internalAPI := httpapi.New(queryService, cfg.InternalKey, cfg.DBTimeout)
	srv := &http.Server{Addr: cfg.Addr, Handler: observability.Handler(store.Ping, queryStore.Ping, internalAPI, logger), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	exited := make(chan error, 1)
	go func() { exited <- srv.ListenAndServe() }()
	logger.Info("ingest_started", "address", cfg.Addr)
	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-exited:
		stop()
	}
	stop()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if srv.Shutdown(shutdown) != nil {
		srv.Close()
	}
	wg.Wait()
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return errors.New("HTTP server failed")
	}
	return nil
}
