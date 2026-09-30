package config

import (
	"example.com/akuanaktehat/pvmbg-mock/internal/auth"
	"fmt"
	"net"
	"os"
	"time"
)

type Config struct {
	HTTPAddr, TokenHash, AdminHash         string
	GenerationInterval, DelayMin, DelayMax time.Duration
}

func Load() (Config, error) {
	c := Config{HTTPAddr: os.Getenv("HTTP_ADDR"), TokenHash: os.Getenv("PVMBG_TOKEN_HASH"), AdminHash: os.Getenv("ADMIN_KEY_HASH")}
	if c.HTTPAddr == "" {
		c.HTTPAddr = ":8082"
	}
	if _, _, err := net.SplitHostPort(c.HTTPAddr); err != nil {
		return c, fmt.Errorf("HTTP_ADDR: %w", err)
	}
	if !auth.ValidHash(c.TokenHash) || !auth.ValidHash(c.AdminHash) {
		return c, fmt.Errorf("PVMBG_TOKEN_HASH and ADMIN_KEY_HASH must be lowercase SHA-256 hex")
	}
	if c.TokenHash == c.AdminHash {
		return c, fmt.Errorf("data and admin credentials must differ")
	}
	var err error
	c.GenerationInterval, err = duration("GENERATION_INTERVAL", 10*time.Second)
	if err != nil {
		return c, err
	}
	if c.GenerationInterval <= 0 || c.GenerationInterval > 10*time.Second {
		return c, fmt.Errorf("GENERATION_INTERVAL must be >0 and <=10s")
	}
	c.DelayMin, err = duration("PVMBG_DELAY_MIN", 500*time.Millisecond)
	if err != nil {
		return c, err
	}
	c.DelayMax, err = duration("PVMBG_DELAY_MAX", 3*time.Second)
	if err != nil {
		return c, err
	}
	if c.DelayMin < 0 || c.DelayMax < c.DelayMin || c.DelayMax > 20*time.Second {
		return c, fmt.Errorf("PVMBG delays must satisfy 0 <= MIN <= MAX <=20s")
	}
	return c, nil
}
func duration(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s duration", name)
	}
	return d, nil
}
