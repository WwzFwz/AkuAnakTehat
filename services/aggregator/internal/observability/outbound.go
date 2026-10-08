package observability

import (
	"context"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"github.com/jackc/pgx/v5"
	"github.com/twmb/franz-go/pkg/kgo"
	"log/slog"
	"time"
)

type outboundID struct{}
type queryStart struct{}

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, outboundID{}, id)
}
func ID(ctx context.Context) string { id, _ := ctx.Value(outboundID{}).(string); return id }
func RequestID(ctx context.Context) string {
	if id := ID(ctx); id != "" {
		return id
	}
	id, _ := hazard.UUID()
	return id
}
func Outbound(ctx context.Context, dependency, operation string, started time.Time, err error) {
	slog.Info("outbound_call", "dependency", dependency, "operation", operation, "correlation_id", RequestID(ctx), "latency_ms", float64(time.Since(started).Microseconds())/1000, "ok", err == nil)
}

type PostgresTracer struct{}

func (PostgresTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	ctx = WithID(ctx, RequestID(ctx))
	return context.WithValue(ctx, queryStart{}, time.Now())
}
func (PostgresTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	start, _ := ctx.Value(queryStart{}).(time.Time)
	Outbound(ctx, "postgres", "query", start, data.Err)
}

// Protocol requests may aggregate records or be heartbeats. Give these their own
// operation ID; per-record publish logs retain the ingest correlation ID.
type KafkaHook struct{}

func (KafkaHook) OnBrokerE2E(_ kgo.BrokerMetadata, key int16, e kgo.BrokerE2E) {
	slog.Info("outbound_call", "dependency", "kafka", "api_key", key, "correlation_id", RequestID(context.Background()), "latency_ms", float64(e.DurationE2E().Microseconds())/1000, "ok", e.Err() == nil)
}
