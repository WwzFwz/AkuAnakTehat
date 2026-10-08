# store

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

SQLite menyimpan hazard_view: hazard_id (PK), version, event_id, payload envelope lengkap, updated_at. UPSERT versi lebih tinggi mencegah replay/versi lama menimpa data. WAL, synchronous=FULL, busy_timeout 1s, dan satu koneksi. File berada di named volume /data/view.db. List memakai cursor hazard_id eksklusif. Uji membuka ulang file memastikan data dan dedup bertahan setelah restart.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [sqlite.go](sqlite.go) | SQLite hazard_view, UPSERT versi lebih tinggi, pagination dengan batas byte, dan probe database. |
| [sqlite_test.go](sqlite_test.go) | Pengujian `TestPersistentLatestVersion`. |

## Perilaku dan batas saat ini

Halaman view memakai budget payload 7 MiB. Store mengembalikan paling banyak satu record tambahan untuk menentukan kelanjutan; handler hanya memajukan cursor sampai record terakhir yang benar-benar dikirim.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
