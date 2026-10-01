# dedup

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: sqlite.go.

SQLite processed memiliki primary key (hazard_id, version), event_id, alert, processed_at. Marker ditulis sesudah sender sukses. WAL, synchronous=FULL, busy_timeout 1s, satu koneksi; file /data/processed.db pada named volume. Indeks event_id mendukung halaman audit. Pengujian persistence berada di application/apply_test.go.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
