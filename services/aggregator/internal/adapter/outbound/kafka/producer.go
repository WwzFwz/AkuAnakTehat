package kafka

import (
	"context"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"example.com/akuanaktehat/aggregator/internal/observability"
	"example.com/akuanaktehat/aggregator/internal/worker/outbox"
	"github.com/twmb/franz-go/pkg/kerr"
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
	c, err := kgo.NewClient(kgo.WithHooks(observability.KafkaHook{}), kgo.SeedBrokers(brokers...), kgo.RequiredAcks(kgo.AllISRAcks()), kgo.RecordDeliveryTimeout(timeout), kgo.ProducerBatchMaxBytes(5<<20), kgo.MaxBufferedRecords(100), kgo.DialTimeout(2*time.Second))
	if err != nil {
		return nil, errors.New("invalid Kafka producer configuration")
	}
	return &Producer{Client: c, Topic: topic, Timeout: timeout}, nil
}
func (p *Producer) Publish(ctx context.Context, m outbox.Message) error {
	if len(m.Payload) > hazard.MaxEventBytes {
		return outbox.ErrPermanent
	}
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()
	// The relay is sequential. Keep an uncertain in-flight result across local
	// timeouts: issuing later versions while this one is unresolved is unsafe.
	if p.pending == nil {
		p.pendingID = m.ID
		p.pending = make(chan error, 1)
		result := p.pending
		record := &kgo.Record{Topic: p.Topic, Key: []byte(m.HazardID), Value: m.Payload, Headers: []kgo.RecordHeader{{Key: "correlation_id", Value: []byte(m.CorrelationID)}, {Key: "event_id", Value: []byte(m.EventID)}}}
		started := time.Now()
		p.Client.Produce(ctx, record, func(_ *kgo.Record, err error) {
			observability.Outbound(observability.WithID(ctx, m.CorrelationID), "kafka", "produce", started, err)
			result <- err
		})
	}
	if p.pendingID != m.ID {
		return errors.New("unresolved earlier publication")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-p.pending:
		p.pending = nil
		if errors.Is(err, kerr.MessageTooLarge) || errors.Is(err, kerr.RecordListTooLarge) {
			return outbox.ErrPermanent
		}
		return err
	}
}
func (p *Producer) Close() { p.Client.Close() }
