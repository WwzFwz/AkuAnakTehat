package config

import (
	"example.com/akuanaktehat/bmkg-mock/internal/auth"
	"fmt"
	"net"
	"os"
	"time"
)

type Config struct {
	HTTPAddr, KeyHash              string
	GenerationInterval, FixedDelay time.Duration
}

func Load() (Config, error) {
	c := Config{HTTPAddr: os.Getenv("HTTP_ADDR"), KeyHash: os.Getenv("BMKG_KEY_HASH")}
	if c.HTTPAddr == "" {
		c.HTTPAddr = ":8081"
	}
	if _, _, err := net.SplitHostPort(c.HTTPAddr); err != nil {
		return c, fmt.Errorf("HTTP_ADDR: %w", err)
	}
	if !auth.ValidHash(c.KeyHash) {
		return c, fmt.Errorf("BMKG_KEY_HASH must be lowercase SHA-256 hex")
	}
	var err error
	c.GenerationInterval, err = duration("GENERATION_INTERVAL", 10*time.Second)
	if err != nil {
		return c, err
	}
	if c.GenerationInterval <= 0 || c.GenerationInterval > 10*time.Second {
		return c, fmt.Errorf("GENERATION_INTERVAL must be >0 and <=10s")
	}
	c.FixedDelay, err = duration("BMKG_FIXED_DELAY", 100*time.Millisecond)
	if err != nil {
		return c, err
	}
	if c.FixedDelay < 50*time.Millisecond || c.FixedDelay > 150*time.Millisecond {
		return c, fmt.Errorf("BMKG_FIXED_DELAY must be 50ms..150ms")
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
