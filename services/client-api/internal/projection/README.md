# projection

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Menyusun respons menggunakan allowlist field sesuai hak akses.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Project(hazard, scope, fields) menghasilkan map baru berisi field yang diizinkan.

## Dependensi

- DTO lokal client-api dan scope terverifikasi; dipakai application.

## Aturan penting

- Ringkasan: hazard_id, source, hazard_type, severity, area_name, occurred_at, ingested_at.
- Mentah: source_ref_id, latitude, longitude, attributes.
- Proyeksi berlaku pada list dan detail setiap request; cache belum digunakan.
- Jangan memodifikasi map/slice bersama yang mungkin berasal dari cache.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [projector.go](projector.go) | Membentuk objek respons baru berdasarkan allowlist scope dan pilihan field. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
