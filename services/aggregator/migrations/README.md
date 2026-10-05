# migrations

[Panduan service](../README.md) · [Peta repository](../../../README.md)

Skema berversi Canonical Store; dimiliki dan dijalankan oleh Aggregator.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** jalur A dan migration pendukung query B sudah diimplementasikan. Migration awal dan `002_source_status_stale_since` tersedia; cakupan pengujian ada di [bukti ingest](../../../docs/evidence/ingest/README.md).

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `001_initial.up.sql` | Skema awal hazard_events, outbox, source_status, checkpoint/watermark, tsunami_warnings, dan quarantine. |
| `001_initial.down.sql` | Rollback pasangan migrasi awal untuk lingkungan pengembangan. |
| `embed.go` | Embed *.sql dari folder ini untuk migration runner. |

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

## Langkah implementasi dan verifikasi

- Sepakati skema dengan A/B/C.
- Siapkan indeks query per tipe dan query gabungan sesuai pengukuran.
- Verifikasi startup baru dan startup ulang tanpa menjalankan DDL dua kali.
