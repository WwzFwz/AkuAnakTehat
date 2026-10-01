package kafka

import (
	"context"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/worker/outbox"
	"github.com/twmb/franz-go/pkg/kgo"
	"time"
)

type Producer struct {
	Client    *kgo.Client
	Topic     string
	Timeout   time.Duration
	pendingID int64
	pending   chan error
}

func New(brokers []string, topic string, timeout time.Duration) (*Producer, error) {
	c, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.RequiredAcks(kgo.AllISRAcks()), kgo.RecordDeliveryTimeout(timeout), kgo.ProducerBatchMaxBytes(1<<20), kgo.MaxBufferedRecords(100), kgo.DialTimeout(2*time.Second))
	if err != nil {
		return nil, errors.New("invalid Kafka producer configuration")
	}
	return &Producer{Client: c, Topic: topic, Timeout: timeout}, nil
}
func (p *Producer) Publish(ctx context.Context, m outbox.Message) error {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()
	// The relay is sequential. Keep an uncertain in-flight result across local
	// timeouts: issuing later versions while this one is unresolved is unsafe.
	if p.pending == nil {
		p.pendingID = m.ID
		p.pending = make(chan error, 1)
		result := p.pending
		record := &kgo.Record{Topic: p.Topic, Key: []byte(m.HazardID), Value: m.Payload, Headers: []kgo.RecordHeader{{Key: "correlation_id", Value: []byte(m.CorrelationID)}, {Key: "event_id", Value: []byte(m.EventID)}}}
		p.Client.Produce(ctx, record, func(_ *kgo.Record, err error) { result <- err })
	}
	if p.pendingID != m.ID {
		return errors.New("unresolved earlier publication")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-p.pending:
		p.pending = nil
		return err
	}
}
func (p *Producer) Close() { p.Client.Close() }
