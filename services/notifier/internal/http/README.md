# http

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

GET /health untuk liveness; GET /ready memeriksa SQLite dan Kafka dalam 2s. GET /processed mengembalikan ledger internal {data: [{event_id,hazard_id,version,alert,processed_at}], next_cursor}. limit 1..200 (default 100), after adalah event_id eksklusif. alert=true berarti simulator kirim telah dipanggil, bukan bukti penerimaan oleh pihak eksternal. Port hanya loopback.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [handler.go](handler.go) | Melayani endpoint health, readiness, dan inspeksi state lokal consumer. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
