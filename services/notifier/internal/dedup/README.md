# dedup

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

SQLite processed memiliki primary key (hazard_id, version), event_id, alert, processed_at. Marker ditulis sesudah sender sukses. WAL, synchronous=FULL, busy_timeout 1s, satu koneksi; file /data/processed.db pada named volume. Indeks event_id mendukung halaman audit. Pengujian persistence berada di application/apply_test.go.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [sqlite.go](sqlite.go) | SQLite processed untuk marker (hazard_id, version) dan halaman audit. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
