# store

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: sqlite.go, sqlite_test.go.

SQLite menyimpan hazard_view: hazard_id (PK), version, event_id, payload envelope lengkap, updated_at. UPSERT versi lebih tinggi mencegah replay/versi lama menimpa data. WAL, synchronous=FULL, busy_timeout 1s, dan satu koneksi. File berada di named volume /data/view.db. List memakai cursor hazard_id eksklusif. Uji membuka ulang file memastikan data dan dedup bertahan setelah restart.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
