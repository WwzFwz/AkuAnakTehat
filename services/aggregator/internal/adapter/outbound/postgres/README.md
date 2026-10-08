# postgres

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Implementasi penyimpanan Aggregator, dengan pemisahan kepemilikan file untuk transaksi ingest, query, dan outbox.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Memenuhi UnitOfWork/Tx milik ingest, HazardQuery milik query, serta OutboxStore milik worker/outbox.
- Driver menggunakan pgx; interface tidak mengekspos driver ke domain.

## Dependensi

- Port application/ingest, application/query, worker/outbox, serta migrations.

## Aturan penting

- Hanya Aggregator memiliki akses langsung Canonical Store.
- Semua repository di dalam callback WithTx memakai transaksi yang sama.
- Query menggunakan parameter; pool total dibatasi; timeout lokal aktif.
- Pool terpisah membatasi perebutan koneksi, bukan menjamin CPU/IO database terisolasi.
- LISTEN/NOTIFY tambahan; relay inti tidak bergantung padanya.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [ingest_integration_test.go](ingest_integration_test.go) | Uji transaksi ingest, korelasi, hash, checkpoint, dan persistensi memakai PostgreSQL. |
| [ingest_repository.go](ingest_repository.go) | Penyimpanan hazard, warning, checkpoint, karantina, outbox, dan status polling. |
| [migrate.go](migrate.go) | Menjalankan migrasi embedded menggunakan golang-migrate. |
| [outbox_repository.go](outbox_repository.go) | Membaca pending outbox, menyimpan ACK atau penolakan, dan membersihkan row published. |
| [pool.go](pool.go) | Membuka pool PostgreSQL dengan batas koneksi, timeout, dan tracer. |
| [query_repository.go](query_repository.go) | Query hazard dengan filter, pagination jumlah dan byte, serta freshness sumber. |
| [query_repository_integration_test.go](query_repository_integration_test.go) | Uji filter, pagination, detail, dan freshness memakai PostgreSQL. |
| [robustness_integration_test.go](robustness_integration_test.go) | Uji PostgreSQL untuk batas 4 MiB, pagination byte, dan replay backlog 600 record. |
| [unit_of_work.go](unit_of_work.go) | Menjalankan callback transaksi ingest dengan timeout, rollback, dan retry terbatas. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
