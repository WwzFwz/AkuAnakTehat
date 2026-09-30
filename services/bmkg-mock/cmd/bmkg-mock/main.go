package main

import (
	"context"
	"errors"
	"example.com/akuanaktehat/bmkg-mock/internal/config"
	"example.com/akuanaktehat/bmkg-mock/internal/generator"
	httpapi "example.com/akuanaktehat/bmkg-mock/internal/http"
	"example.com/akuanaktehat/bmkg-mock/internal/store"
	"example.com/akuanaktehat/bmkg-mock/seed"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "bmkg-mock")
	if err := run(logger); err != nil {
		logger.Error("service_failed", "error", err)
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
	data := store.New()
	if err := seed.Load(data); err != nil {
		return err
	}
	handler := httpapi.New(cfg, data, logger)
	go generator.Run(ctx, data, cfg.GenerationInterval, logger)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10, BaseContext: func(net.Listener) context.Context { return ctx }}
	errs := make(chan error, 1)
	go func() { errs <- server.ListenAndServe() }()
	logger.Info("service_started", "address", cfg.HTTPAddr)
	select {
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return err
		}
		return nil
	}
}
