package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/pemda-portal/internal/config"
	"example.com/akuanaktehat/pemda-portal/internal/contract"
	"log/slog"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"sync/atomic"
)

type Processor interface {
	Handle(context.Context, contract.Event) error
}

// Delivery makes the acknowledgement boundary explicit and testable.
type Delivery interface {
	DeadLetter(context.Context, *kgo.Record, string, int) error
	Commit(context.Context, *kgo.Record) error
}
type Consumer struct {
	Client       *kgo.Client
	Config       config.Config
	Processor    Processor
	Logger       *slog.Logger
	pingInFlight atomic.Bool
}

func New(c config.Config, p Processor, l *slog.Logger) (*Consumer, error) {
	client, err := kgo.NewClient(kgo.WithHooks(kafkaHook{l}), kgo.SeedBrokers(c.Brokers...), kgo.ConsumerGroup(c.Group), kgo.ConsumeTopics(c.Topic), kgo.DisableAutoCommit(), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), kgo.BlockRebalanceOnPoll(), kgo.RebalanceTimeout(60*time.Second), kgo.FetchMaxBytes(6<<20), kgo.ProducerBatchMaxBytes(5<<20), kgo.RequiredAcks(kgo.AllISRAcks()), kgo.RecordDeliveryTimeout(5*time.Second), kgo.DialTimeout(2*time.Second))
	if err != nil {
		return nil, errors.New("invalid Kafka consumer configuration")
	}
	return &Consumer{Client: client, Config: c, Processor: p, Logger: l}, nil
}
func (c *Consumer) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		fetched := c.Client.PollRecords(ctx, 1)
		if ctx.Err() != nil {
			c.Client.AllowRebalance()
			return ctx.Err()
		}
		if len(fetched.Errors()) > 0 {
			c.Client.AllowRebalance()
			return errors.New("kafka_fetch_failed")
		}
		records := fetched.Records()
		for _, r := range records {
			started := time.Now()
			err := Process(ctx, r, c.Processor, c, c.Config.Attempts, c.Config.Timeout)
			if err != nil {
				c.Client.AllowRebalance()
				return err
			}

			var trace struct {
				EventID       string `json:"event_id"`
				CorrelationID string `json:"correlation_id"`
			}
			_ = json.Unmarshal(r.Value, &trace)
			c.Logger.Info("record_completed", "event_id", trace.EventID, "correlation_id", trace.CorrelationID, "consumer_group", c.Config.Group, "partition", r.Partition, "offset", r.Offset, "latency_ms", time.Since(started).Milliseconds())
		}
		c.Client.AllowRebalance()
	}
	return ctx.Err()
}

// One record at a time: a failure cannot be hidden by committing a later offset.
func Process(ctx context.Context, r *kgo.Record, p Processor, d Delivery, attempts int, timeout time.Duration) error {
	e, err := contract.Decode(r.Key, r.Value)
	reason := "invalid_event"
	used := 1
	if err == nil {
		reason = "processing_failed"
		for used = 1; used <= attempts; used++ {
			attempt, cancel := context.WithTimeout(ctx, timeout)
			err = p.Handle(attempt, e)
			cancel()
			if err == nil {
				break
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if used < attempts {
				timer := time.NewTimer(200 * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}
			}
		}
		if err != nil {
			// A valid record failed due to an operational dependency. Preserve
			// its offset; the supervisor restarts this worker for later retry.
			return errors.New("processing_dependency_unavailable")
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		if err = d.DeadLetter(ctx, r, reason, used); err != nil {
			return err
		}
	}
	return d.Commit(ctx, r)
}
func (c *Consumer) DeadLetter(ctx context.Context, r *kgo.Record, reason string, attempts int) (resultErr error) {
	start := time.Now()
	defer func() { c.traceRecord(r, "dead_letter", start, resultErr) }()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	record := Record(r, c.Config.DLQ, c.Config.Group, reason, attempts)
	result := make(chan error, 1)
	c.Client.Produce(ctx, record, func(_ *kgo.Record, err error) { result <- err })
	select {
	case <-ctx.Done():
		return errors.New("dlq_publish_timeout")
	case err := <-result:
		if err != nil {
			return errors.New("dlq_publish_failed")
		}
	}
	c.Logger.Warn("event_dead_lettered", "consumer_group", c.Config.Group, "source_partition", r.Partition, "source_offset", r.Offset, "reason", reason, "attempts", attempts)
	return nil
}
func (c *Consumer) Commit(ctx context.Context, r *kgo.Record) (resultErr error) {
	start := time.Now()
	defer func() { c.traceRecord(r, "commit_offset", start, resultErr) }()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := c.Client.CommitRecords(ctx, r); err != nil {
		return errors.New("offset_commit_failed")
	}
	return nil
}
func (c *Consumer) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// franz-go v1.18.1 Ping can wait for a connection handshake past ctx's
	// deadline. Bound the HTTP caller's wait and allow at most one unfinished
	// probe, so repeated readiness requests cannot accumulate goroutines.
	if !c.pingInFlight.CompareAndSwap(false, true) {
		return errors.New("Kafka readiness probe in progress")
	}
	result := make(chan error, 1)
	go func() {
		defer c.pingInFlight.Store(false)
		result <- c.Client.Ping(ctx)
	}()
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (c *Consumer) Close() { c.Client.Close() }

// Record preserves original bytes and carries failure metadata in headers.
func Record(r *kgo.Record, topic, group, reason string, attempts int) *kgo.Record {
	corr := ""
	for _, h := range r.Headers {
		if h.Key == "correlation_id" {
			corr = string(h.Value)
		}
	}
	if e, err := contract.Decode(r.Key, r.Value); err == nil {
		corr = e.CorrelationID
	}
	headers := []kgo.RecordHeader{
		{Key: "consumer_group", Value: []byte(group)}, {Key: "failure_reason", Value: []byte(reason)},
		{Key: "attempts", Value: []byte(strconv.Itoa(attempts))}, {Key: "source_topic", Value: []byte(r.Topic)},
		{Key: "source_partition", Value: []byte(strconv.FormatInt(int64(r.Partition), 10))}, {Key: "source_offset", Value: []byte(strconv.FormatInt(r.Offset, 10))},
		{Key: "correlation_id", Value: []byte(corr)},
	}
	return &kgo.Record{Topic: topic, Key: r.Key, Value: r.Value, Headers: headers}
}
