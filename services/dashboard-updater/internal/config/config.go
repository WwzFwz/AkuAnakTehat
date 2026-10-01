package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address, SQLitePath, Topic, DLQ, Group string
	Brokers                                []string
	Timeout                                time.Duration
	Attempts                               int
}

func Load() (Config, error) {
	c := Config{Address: value("HTTP_ADDR", ":8091"), SQLitePath: value("SQLITE_PATH", "/data/view.db"), Topic: value("KAFKA_TOPIC", "bnpb.hazard-events.v1"), DLQ: value("KAFKA_DLQ_TOPIC", "bnpb.hazard-events.v1.dlq"), Group: value("KAFKA_GROUP_ID", "dashboard-updater"), Brokers: strings.Split(value("KAFKA_BROKERS", "kafka:9092"), ",")}
	var err error
	c.Timeout, err = time.ParseDuration(value("PROCESS_TIMEOUT", "2s"))
	if err != nil || c.Timeout < time.Millisecond || c.Timeout > 5*time.Second {
		return c, errors.New("invalid PROCESS_TIMEOUT (1ms..5s)")
	}
	c.Attempts, err = strconv.Atoi(value("MAX_ATTEMPTS", "3"))
	if err != nil || c.Attempts < 1 || c.Attempts > 5 {
		return c, errors.New("invalid MAX_ATTEMPTS (1..5)")
	}
	if c.Topic == c.DLQ {
		return c, errors.New("source and DLQ topics must differ")
	}
	for i, b := range c.Brokers {
		c.Brokers[i] = strings.TrimSpace(b)
		if c.Brokers[i] == "" {
			return c, errors.New("empty Kafka broker")
		}
	}
	return c, nil
}
func value(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
