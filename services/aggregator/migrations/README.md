# migrations

[Panduan service](../README.md) · [Peta repository](../../../README.md)

Skema berversi Canonical Store; dimiliki dan dijalankan oleh Aggregator.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Runner menggunakan golang-migrate dan mencatat versi pada schema_migrations.
- Schema_observations hanya ditambahkan bila fitur pencatatan drift persisten dikerjakan.

## Dependensi

- Dipanggil adapter/outbound/postgres saat inisialisasi; service lain tidak menjalankan migrasi ini.

## Aturan penting

- attributes JSONB menerima confidence_level tanpa ALTER TABLE.
- `source_status.stale_since` mencatat awal status non-healthy untuk metadata freshness query.
- UNIQUE(source, source_ref_id); koordinat kanonik NOT NULL.
- Outbox menyimpan payload per versi, bukan mengambil isi hazard terbaru saat publish.
- Skema checkpoint harus benar-benar dibuat; watermark bukan hanya state memori.
- Migrasi yang sudah diterapkan tidak diedit sembarangan; perubahan berikutnya memakai versi baru.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [001_initial.down.sql](001_initial.down.sql) | Membatalkan skema awal Canonical Store. |
| [001_initial.up.sql](001_initial.up.sql) | Menerapkan skema awal Canonical Store. |
| [002_source_status_stale_since.down.sql](002_source_status_stale_since.down.sql) | Membatalkan stale_since pada status sumber. |
| [002_source_status_stale_since.up.sql](002_source_status_stale_since.up.sql) | Menerapkan stale_since pada status sumber. |
| [003_outbox_rejections.down.sql](003_outbox_rejections.down.sql) | Membatalkan status penolakan permanen outbox. |
| [003_outbox_rejections.up.sql](003_outbox_rejections.up.sql) | Menerapkan status penolakan permanen outbox. |
| [embed.go](embed.go) | Embed file SQL untuk runner migrasi Aggregator. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
