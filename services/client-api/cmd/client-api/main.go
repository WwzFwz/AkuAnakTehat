package main

import (
	"context"
	"errors"
	httpapi "example.com/akuanaktehat/client-api/internal/adapter/inbound/http"
	"example.com/akuanaktehat/client-api/internal/adapter/outbound/aggregator"
	"example.com/akuanaktehat/client-api/internal/application"
	"example.com/akuanaktehat/client-api/internal/authn"
	"example.com/akuanaktehat/client-api/internal/config"
	"example.com/akuanaktehat/client-api/internal/middleware"
	"example.com/akuanaktehat/client-api/internal/observability"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := observability.Logger()
	c, err := config.Load()
	if err != nil {
		log.Error("configuration failed", "error", err)
		os.Exit(1)
	}
	key, err := authn.LoadKey(c.PublicKey)
	if err != nil {
		log.Error("invalid verification key")
		os.Exit(1)
	}
	upstream := aggregator.New(c.AggregatorURL, c.InternalKey, c.Timeout)
	app := application.Service{Upstream: upstream}
	verifier := authn.Verifier{Key: key, Issuer: c.Issuer, Audience: c.Audience}
	handler := httpapi.New(app, verifier, middleware.New(c.Rate, c.Burst, c.Concurrent), c.PageDefault, c.PageMax)
	server := &http.Server{Addr: c.Addr, Handler: observability.Middleware(log, handler), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16384}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	log.Info("listening", "address", c.Addr)
	select {
	case err := <-done:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if server.Shutdown(shutdown) != nil {
			_ = server.Close()
		}
	}
}
