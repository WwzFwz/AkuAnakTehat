# dlq

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

Membangun record DLQ dengan key/value asli dan header consumer_group, failure_reason, attempts, source_topic, source_partition, source_offset, correlation_id. Consumer menunggu ACK lalu commit offset asal; payload tidak dibungkus ulang. Uji metadata ada di consumer/consumer_test.go.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [record.go](record.go) | Membentuk record DLQ dengan payload asal serta metadata consumer dan kegagalannya. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
