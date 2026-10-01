# dlq

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: record.go.

Membangun record DLQ dengan key/value asli dan header consumer_group, failure_reason, attempts, source_topic, source_partition, source_offset, correlation_id. Consumer menunggu ACK lalu commit offset asal; payload tidak dibungkus ulang. Uji metadata ada di consumer/consumer_test.go.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
