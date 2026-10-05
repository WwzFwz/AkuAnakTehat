package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr, DatabaseURL, BMKGURL, BMKGKey, PVMBGURL, PVMBGToken, InternalKey                      string
	BMKGInterval, PVMBGInterval, Overlap, BMKGTimeout, PVMBGTimeout, DBTimeout, BreakerCooldown time.Duration
	PoolSize, QueryPoolSize                                                                     int32
	BreakerFailures                                                                             int
	KafkaBrokers                                                                                []string
	KafkaTopic                                                                                  string
	PublishTimeout, OutboxInterval, OutboxRetention                                             time.Duration
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func Load() (Config, error) {
	c := Config{Addr: env("HTTP_ADDR", ":9000"), DatabaseURL: os.Getenv("DATABASE_URL"), BMKGURL: env("BMKG_URL", "http://bmkg-mock:8081"), BMKGKey: os.Getenv("BMKG_API_KEY"), PVMBGURL: env("PVMBG_URL", "http://pvmbg-mock:8082"), PVMBGToken: os.Getenv("PVMBG_TOKEN"), InternalKey: os.Getenv("INTERNAL_KEY")}
	c.KafkaBrokers = strings.Split(env("KAFKA_BROKERS", "kafka:9092"), ",")
	c.KafkaTopic = env("KAFKA_TOPIC", "bnpb.hazard-events.v1")
	for i, broker := range c.KafkaBrokers {
		c.KafkaBrokers[i] = strings.TrimSpace(broker)
		if c.KafkaBrokers[i] == "" {
			return c, errors.New("invalid KAFKA_BROKERS")
		}
	}
	for _, v := range []struct {
		key, def string
		dst      *time.Duration
	}{{"KAFKA_PUBLISH_TIMEOUT", "5s", &c.PublishTimeout}, {"OUTBOX_POLL_INTERVAL", "1s", &c.OutboxInterval}, {"OUTBOX_RETENTION", "24h", &c.OutboxRetention}} {
		d, err := time.ParseDuration(env(v.key, v.def))
		if err != nil || d < time.Millisecond || d > 30*24*time.Hour {
			return c, errors.New("invalid " + v.key)
		}
		*v.dst = d
	}
	if c.DatabaseURL == "" || c.BMKGKey == "" || c.PVMBGToken == "" || c.InternalKey == "" {
		return c, errors.New("DATABASE_URL, BMKG_API_KEY, PVMBG_TOKEN and INTERNAL_KEY are required")
	}
	for _, s := range []string{c.BMKGURL, c.PVMBGURL} {
		u, err := url.Parse(s)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return c, errors.New("invalid source URL")
		}
	}
	for _, v := range []struct {
		key, defaultValue string
		dst               *time.Duration
	}{{"BMKG_POLL_INTERVAL", "2s", &c.BMKGInterval}, {"PVMBG_POLL_INTERVAL", "5s", &c.PVMBGInterval}, {"POLL_OVERLAP", "10s", &c.Overlap}, {"BMKG_TIMEOUT", "1s", &c.BMKGTimeout}, {"PVMBG_TIMEOUT", "4s", &c.PVMBGTimeout}, {"DB_TIMEOUT", "2s", &c.DBTimeout}, {"BREAKER_COOLDOWN", "10s", &c.BreakerCooldown}} {
		d, err := time.ParseDuration(env(v.key, v.defaultValue))
		if err != nil || d < time.Millisecond || d > time.Hour {
			return c, errors.New("invalid " + v.key)
		}
		*v.dst = d
	}
	n, err := strconv.Atoi(env("DB_POOL_SIZE", "5"))
	if err != nil || n < 2 || n > 20 {
		return c, errors.New("invalid DB_POOL_SIZE")
	}
	c.PoolSize = int32(n)
	n, err = strconv.Atoi(env("QUERY_DB_POOL_SIZE", "3"))
	if err != nil || n < 1 || n > 20 {
		return c, errors.New("invalid QUERY_DB_POOL_SIZE")
	}
	c.QueryPoolSize = int32(n)
	c.BreakerFailures, err = strconv.Atoi(env("BREAKER_FAILURES", "3"))
	if err != nil || c.BreakerFailures < 1 || c.BreakerFailures > 100 {
		return c, errors.New("invalid BREAKER_FAILURES")
	}
	return c, nil
}
