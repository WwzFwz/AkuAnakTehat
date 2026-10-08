package main

import (
	"context"
	"errors"
	httpapi "example.com/akuanaktehat/auth-service/internal/adapter/inbound/http"
	redisstore "example.com/akuanaktehat/auth-service/internal/adapter/outbound/redis"
	"example.com/akuanaktehat/auth-service/internal/application"
	"example.com/akuanaktehat/auth-service/internal/config"
	"example.com/akuanaktehat/auth-service/internal/observability"
	"example.com/akuanaktehat/auth-service/internal/token"
	"github.com/redis/go-redis/v9"
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
	key, err := token.LoadKey(c.PrivateKey)
	if err != nil {
		log.Error("invalid signing key")
		os.Exit(1)
	}
	client := redis.NewClient(&redis.Options{Addr: c.RedisAddr, Password: c.RedisPassword, DialTimeout: c.RedisTimeout, ReadTimeout: c.RedisTimeout, WriteTimeout: c.RedisTimeout, ContextTimeoutEnabled: true, MaxRetries: -1, PoolSize: 20, PoolTimeout: c.RedisTimeout})
	client.AddHook(observability.RedisHook{})
	defer client.Close()
	store := &redisstore.Store{Client: client, Timeout: c.RedisTimeout}
	app := &application.Service{Clients: c.Clients, Store: store, Signer: token.Signer{Key: key, Issuer: c.Issuer, Audience: c.Audience, Now: time.Now}, AccessTTL: c.AccessTTL, RefreshTTL: c.RefreshTTL, Now: time.Now}
	server := &http.Server{Addr: c.Addr, Handler: observability.Middleware(log, httpapi.New(app, c.Rate, c.Burst)), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16384}
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
