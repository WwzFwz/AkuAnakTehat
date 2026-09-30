package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr, PublicKey, Issuer, Audience, AggregatorURL, InternalKey string
	Timeout                                                       time.Duration
	Rate, Burst, Concurrent, PageDefault, PageMax                 int
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func integer(k, d string) (int, error) {
	v, e := strconv.Atoi(env(k, d))
	if e != nil || v < 1 {
		return 0, errors.New("invalid " + k)
	}
	return v, nil
}
func Load() (Config, error) {
	c := Config{Addr: env("HTTP_ADDR", ":8080"), PublicKey: os.Getenv("JWT_PUBLIC_KEY_FILE"), Issuer: env("JWT_ISSUER", "bnpb-auth"), Audience: env("JWT_AUDIENCE", "bnpb-api"), AggregatorURL: env("AGGREGATOR_URL", "http://aggregator:9000"), InternalKey: os.Getenv("INTERNAL_KEY")}
	if c.PublicKey == "" || c.InternalKey == "" {
		return c, errors.New("JWT_PUBLIC_KEY_FILE and INTERNAL_KEY required")
	}
	u, e := url.Parse(c.AggregatorURL)
	if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return c, errors.New("invalid AGGREGATOR_URL")
	}
	if c.Timeout, e = time.ParseDuration(env("AGGREGATOR_TIMEOUT", "1500ms")); e != nil || c.Timeout <= 0 || c.Timeout > 1500*time.Millisecond {
		return c, errors.New("AGGREGATOR_TIMEOUT must be positive and <=1500ms")
	}
	if c.Rate, e = integer("RATE_LIMIT_RPS", "100"); e != nil {
		return c, e
	}
	if c.Burst, e = integer("RATE_LIMIT_BURST", "200"); e != nil {
		return c, e
	}
	if c.Concurrent, e = integer("MAX_CONCURRENT", "100"); e != nil {
		return c, e
	}
	if c.PageDefault, e = integer("PAGE_DEFAULT", "100"); e != nil {
		return c, e
	}
	if c.PageMax, e = integer("PAGE_MAX", "500"); e != nil {
		return c, e
	}
	if c.PageDefault > c.PageMax || c.PageMax > 500 {
		return c, errors.New("invalid pagination bounds")
	}
	return c, nil
}
