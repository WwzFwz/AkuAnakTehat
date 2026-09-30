# hazard

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Model HazardEvent dan aturan murni yang dipakai ingest serta query.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `event.go` | Entitas kanonik, Source, HazardType, dan Severity. |
| `record.go` | Metadata internal: version, content_hash, updated_at, dan last_seen_at. |
| `hash.go` | Hash deterministik dari isi bisnis. |
| `validation.go` | Validasi field kanonik tanpa I/O. |

## Kontrak dan alur

- HazardEvent memuat field Ringkasan dan Mentah sesuai spesifikasi.
- Operasi yang direncanakan: Validate(event), ContentHash(event), dan perbandingan severity.

## Dependensi

- Standard library saja; dipakai application/canonicalize, ingest, dan query.

## Aturan penting

- Tidak meng-import HTTP, driver database, Kafka, atau adapter.
- Hash tidak memuat ingested_at, updated_at, last_seen_at, atau correlation ID.
- hazard_id stabil melalui identitas (source, source_ref_id); metadata internal tidak masuk respons publik.

## Langkah implementasi dan verifikasi

- Sepakati representasi tipe dan validasi.
- Verifikasi hash tidak berubah karena urutan key JSON, tetapi berubah saat isi bisnis berubah.
