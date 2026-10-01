package dlq

import (
	"example.com/akuanaktehat/notifier/internal/contract"
	"github.com/twmb/franz-go/pkg/kgo"
	"strconv"
)

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
