# hazard

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Model HazardEvent dan aturan murni yang dipakai ingest serta query.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- HazardEvent memuat field Ringkasan dan Mentah sesuai spesifikasi.
- Event dan Record menyimpan model; ContentHash menghitung hash isi bisnis. Validasi input berada di canonicalize.Decode dan aturan severity berada di canonicalize serta domain/tsunami.

## Dependensi

- Standard library saja; dipakai application/canonicalize, ingest, dan query.

## Aturan penting

- Tidak meng-import HTTP, driver database, Kafka, atau adapter.
- Hash tidak memuat ingested_at, updated_at, last_seen_at, atau correlation ID.
- hazard_id stabil melalui identitas (source, source_ref_id); metadata internal tidak masuk respons publik.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [event.go](event.go) | HazardEvent, metadata Record, UUID, dan hash deterministik isi bisnis. |
| [limits.go](limits.go) | Batas envelope 4 MiB, halaman 8 MiB, dan error event terlalu besar. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
