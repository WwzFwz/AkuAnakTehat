package ingest

import (
	"context"
	"encoding/json"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"example.com/akuanaktehat/aggregator/internal/domain/tsunami"
	"time"
)

type UnitOfWork interface {
	WithTx(context.Context, func(Tx) error) error
}
type Tx interface {
	FindHazard(context.Context, string, string) (hazard.Record, bool, error)
	PutHazard(context.Context, hazard.Record) error
	FindWarning(context.Context, string) (tsunami.Warning, bool, error)
	PutWarning(context.Context, tsunami.Warning) error
	WarningsFor(context.Context, string) ([]tsunami.Warning, error)
	AppendOutbox(context.Context, OutboxEvent) error
	SaveCheckpoint(context.Context, string, time.Time) error
	Quarantine(context.Context, RejectedRecord) error
}
type OutboxEvent struct {
	EventID, HazardID string
	Version           int64
	Payload           json.RawMessage
	CreatedAt         time.Time
}
type RejectedRecord struct {
	Source, Endpoint, Reason, CorrelationID string
	Payload                                 json.RawMessage
	ObservedAt                              time.Time
}
type CheckpointReader interface {
	ReadCheckpoint(context.Context, string) (time.Time, bool, error)
}
type StatusStore interface {
	RecordPoll(context.Context, string, bool, bool, string, time.Time) error
}
