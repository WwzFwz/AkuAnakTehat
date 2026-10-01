package main

import (
	"context"
	"errors"
	"example.com/akuanaktehat/notifier/internal/application"
	"example.com/akuanaktehat/notifier/internal/config"
	"example.com/akuanaktehat/notifier/internal/consumer"
	"example.com/akuanaktehat/notifier/internal/dedup"
	api "example.com/akuanaktehat/notifier/internal/http"
	"example.com/akuanaktehat/notifier/internal/sender"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if run() != nil {
		os.Exit(1)
	}
}
func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "notifier")
	c, err := config.Load()
	if err != nil {
		logger.Error("invalid_config", "reason", err.Error())
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 5*time.Second)
	db, err := dedup.Open(startup, c.SQLitePath)
	cancel()
	if err != nil {
		logger.Error("sqlite_start_failed")
		return err
	}
	defer db.Close()
	app := &application.Service{Store: db, Sender: &sender.Log{Logger: logger}}
	worker, err := consumer.New(c, app, logger)
	if err != nil {
		logger.Error("kafka_config_failed")
		return err
	}
	defer worker.Close()
	server := &http.Server{Addr: c.Address, Handler: api.Handler(db, worker, db), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	result := make(chan error, 2)
	go func() { result <- server.ListenAndServe() }()
	go func() { result <- worker.Run(ctx) }()
	logger.Info("consumer_started", "consumer_group", c.Group, "topic", c.Topic)
	remaining := 2
	select {
	case <-ctx.Done():
	case err = <-result:
		remaining--
		if !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("consumer_stopped", "reason", "worker_or_http_failed")
		}
	}
	stop()
	shutdown, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()
	_ = server.Shutdown(shutdown)
	for ; remaining > 0; remaining-- {
		<-result
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
