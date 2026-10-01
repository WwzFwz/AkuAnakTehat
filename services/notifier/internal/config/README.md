# config

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: config.go.

Membaca HTTP_ADDR, SQLITE_PATH, KAFKA_BROKERS, KAFKA_TOPIC, KAFKA_DLQ_TOPIC, KAFKA_GROUP_ID, PROCESS_TIMEOUT (default 2s, maksimum 5s), dan MAX_ATTEMPTS (default 3, maksimum 5). Group default sama dengan nama service. Topic sumber dan DLQ harus berbeda.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
