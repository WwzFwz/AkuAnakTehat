# config

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

Membaca HTTP_ADDR, SQLITE_PATH, KAFKA_BROKERS, KAFKA_TOPIC, KAFKA_DLQ_TOPIC, KAFKA_GROUP_ID, PROCESS_TIMEOUT (default 2s, maksimum 5s), dan MAX_ATTEMPTS (default 3, maksimum 5). Group default sama dengan nama service. Topic sumber dan DLQ harus berbeda.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [config.go](config.go) | Membaca environment dan memvalidasi konfigurasi sebelum service dijalankan. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
