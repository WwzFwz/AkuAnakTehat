package observability

import (
	"bytes"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"strings"
	"testing"
)

func TestPostgresTraceKeepsCorrelationWithoutStatementOrArguments(t *testing.T) {
	var output bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	defer slog.SetDefault(old)
	tracer := PostgresTracer{}
	ctx := tracer.TraceQueryStart(WithID(context.Background(), "trace-test"), nil, pgx.TraceQueryStartData{SQL: "secret SQL", Args: []any{"private value"}})
	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{Err: errors.New("private driver error")})
	log := output.String()
	for _, secret := range []string{"secret SQL", "private value", "private driver error"} {
		if strings.Contains(log, secret) {
			t.Fatal("sensitive trace")
		}
	}
	for _, field := range []string{"trace-test", "latency_ms", `"ok":false`} {
		if !strings.Contains(log, field) {
			t.Fatal("incomplete trace", log)
		}
	}
}
