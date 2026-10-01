# http

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: handler.go.

GET /health untuk liveness; GET /ready memeriksa SQLite dan Kafka dalam 2s. GET /processed mengembalikan ledger internal {data: [{event_id,hazard_id,version,alert,processed_at}], next_cursor}. limit 1..200 (default 100), after adalah event_id eksklusif. alert=true berarti simulator kirim telah dipanggil, bukan bukti penerimaan oleh pihak eksternal. Port hanya loopback.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
