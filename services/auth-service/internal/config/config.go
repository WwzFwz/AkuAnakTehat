package config

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/auth-service/internal/domain"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr, PrivateKey, Issuer, Audience, RedisAddr, RedisPassword string
	AccessTTL, RefreshTTL, RedisTimeout                          time.Duration
	Rate, Burst                                                  int
	Clients                                                      map[string]domain.Client
}

func Env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func duration(k, d string) (time.Duration, error) {
	v, e := time.ParseDuration(Env(k, d))
	if e != nil || v <= 0 {
		return 0, errors.New("invalid " + k)
	}
	return v, nil
}
func integer(k, d string) (int, error) {
	v, e := strconv.Atoi(Env(k, d))
	if e != nil || v < 1 {
		return 0, errors.New("invalid " + k)
	}
	return v, nil
}
func Load() (Config, error) {
	c := Config{Addr: Env("HTTP_ADDR", ":8090"), PrivateKey: os.Getenv("JWT_PRIVATE_KEY_FILE"), Issuer: Env("JWT_ISSUER", "bnpb-auth"), Audience: Env("JWT_AUDIENCE", "bnpb-api"), RedisAddr: Env("REDIS_ADDR", "auth-store:6379"), RedisPassword: os.Getenv("REDIS_PASSWORD")}
	var err error
	if c.PrivateKey == "" || os.Getenv("CLIENTS_FILE") == "" || c.RedisPassword == "" {
		return c, errors.New("JWT_PRIVATE_KEY_FILE, CLIENTS_FILE, REDIS_PASSWORD required")
	}
	if c.AccessTTL, err = duration("ACCESS_TOKEN_TTL", "60s"); err != nil {
		return c, err
	}
	if c.RefreshTTL, err = duration("REFRESH_TOKEN_TTL", "8h"); err != nil {
		return c, err
	}
	if c.RedisTimeout, err = duration("REDIS_TIMEOUT", "200ms"); err != nil {
		return c, err
	}
	if c.AccessTTL < time.Second || c.RefreshTTL < c.AccessTTL {
		return c, errors.New("invalid token TTL bounds")
	}
	if c.Rate, err = integer("TOKEN_RATE_LIMIT", "20"); err != nil {
		return c, err
	}
	if c.Burst, err = integer("TOKEN_RATE_BURST", "40"); err != nil {
		return c, err
	}
	b, err := os.ReadFile(os.Getenv("CLIENTS_FILE"))
	if err != nil {
		return c, err
	}
	var clients []domain.Client
	if err = json.Unmarshal(b, &clients); err != nil {
		return c, errors.New("invalid CLIENTS_FILE JSON")
	}
	c.Clients = make(map[string]domain.Client)
	for _, v := range clients {
		hash, e := hex.DecodeString(v.SecretHash)
		if e != nil || len(hash) != 32 || hex.EncodeToString(hash) != v.SecretHash || v.ID == "" || len(v.Scopes) == 0 {
			return c, errors.New("invalid client record")
		}
		if _, ok := c.Clients[v.ID]; ok {
			return c, errors.New("duplicate client")
		}
		summary := false
		seen := map[string]bool{}
		for _, scope := range v.Scopes {
			if seen[scope] || scope != "hazard:read:summary" && scope != "hazard:read:raw" {
				return c, errors.New("invalid scope")
			}
			seen[scope] = true
			if scope == "hazard:read:summary" {
				summary = true
			}
		}
		if !summary {
			return c, errors.New("summary scope required")
		}
		c.Clients[v.ID] = v
	}
	if len(c.Clients) == 0 {
		return c, errors.New("no clients configured")
	}
	return c, nil
}
