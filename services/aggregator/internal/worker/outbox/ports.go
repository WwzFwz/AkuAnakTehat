package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var ErrPermanent = errors.New("permanent_publication_failure")

type Message struct {
	ID                int64
	EventID, HazardID string
	Version           int64
	Payload           json.RawMessage
	CorrelationID     string
}
type Store interface {
	Pending(context.Context, int) ([]Message, error)
	MarkPublished(context.Context, int64, time.Time) error
	Reject(context.Context, int64, string) error
	DeletePublishedBefore(context.Context, time.Time) (int64, error)
}
type Publisher interface {
	Publish(context.Context, Message) error
}
