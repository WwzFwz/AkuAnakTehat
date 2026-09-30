# migrations

[Panduan service](../README.md) · [Peta repository](../../../README.md)

Skema berversi Canonical Store; dimiliki dan dijalankan oleh Aggregator.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

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
- UNIQUE(source, source_ref_id); koordinat kanonik NOT NULL.
- Outbox menyimpan payload per versi, bukan mengambil isi hazard terbaru saat publish.
- Skema checkpoint harus benar-benar dibuat; watermark bukan hanya state memori.
- Migrasi yang sudah diterapkan tidak diedit sembarangan; perubahan berikutnya memakai versi baru.

## Langkah implementasi dan verifikasi

- Sepakati skema dengan A/B/C.
- Siapkan indeks query per tipe dan query gabungan sesuai pengukuran.
- Verifikasi startup baru dan startup ulang tanpa menjalankan DDL dua kali.
