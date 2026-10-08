package consumer

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/twmb/franz-go/pkg/kgo"
	"log/slog"
	"time"
)

type kafkaHook struct{ logger *slog.Logger }

func (h kafkaHook) OnBrokerE2E(_ kgo.BrokerMetadata, key int16, e kgo.BrokerE2E) {
	var id [16]byte
	_, _ = rand.Read(id[:])
	h.logger.Info("outbound_call", "dependency", "kafka", "api_key", key, "correlation_id", hex.EncodeToString(id[:]), "latency_ms", float64(e.DurationE2E().Microseconds())/1000, "ok", e.Err() == nil)
}
func (c *Consumer) traceRecord(r *kgo.Record, operation string, start time.Time, err error) {
	var trace struct {
		CorrelationID string `json:"correlation_id"`
	}
	_ = json.Unmarshal(r.Value, &trace)
	if trace.CorrelationID == "" {
		for _, h := range r.Headers {
			if h.Key == "correlation_id" {
				trace.CorrelationID = string(h.Value)
			}
		}
	}
	c.Logger.Info("outbound_call", "dependency", "kafka", "operation", operation, "correlation_id", trace.CorrelationID, "latency_ms", float64(time.Since(start).Microseconds())/1000, "ok", err == nil)
}
