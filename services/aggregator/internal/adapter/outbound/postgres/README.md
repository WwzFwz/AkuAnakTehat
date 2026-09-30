# postgres

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Implementasi penyimpanan Aggregator, dengan pemisahan kepemilikan file untuk transaksi ingest, query, dan outbox.

**Pemilik rencana:** A/B/C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `pool.go` | Membuka pool ingest/baca, timeout, dan lifecycle koneksi. |
| `unit_of_work.go` | A: WithTx, commit, rollback, dan repository terikat transaksi. |
| `ingest_repository.go` | A: hazard, warning, watermark, quarantine, dan status sumber. |
| `query_repository.go` | B: list/get dan metadata sumber. |
| `outbox_repository.go` | C: pending, mark published, dan housekeeping. |
| `migrate.go` | Menjalankan migrasi embedded milik Aggregator. |

## Kontrak dan alur

- Memenuhi UnitOfWork/Tx milik ingest, HazardQuery milik query, serta OutboxStore milik worker/outbox.
- Driver direncanakan pgx; interface tidak mengekspos driver ke domain.

## Dependensi

- Port application/ingest, application/query, worker/outbox, serta migrations.

## Aturan penting

- Hanya Aggregator memiliki akses langsung Canonical Store.
- Semua repository di dalam callback WithTx memakai transaksi yang sama.
- Query menggunakan parameter; pool total dibatasi; timeout lokal aktif.
- Pool terpisah membatasi perebutan koneksi, bukan menjamin CPU/IO database terisolasi.
- LISTEN/NOTIFY tambahan; relay inti tidak bergantung padanya.

## Langkah implementasi dan verifikasi

- A/B/C sepakati port dan skema sebelum coding.
- Verifikasi rollback, upsert unik, urutan versi, keyset pagination, dan outbox ACK state.
