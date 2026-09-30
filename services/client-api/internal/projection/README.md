# projection

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Menyusun respons menggunakan allowlist field sesuai hak akses.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Uji integrasi runtime belum dilakukan. Berkas yang sudah ada: `projector.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `allowlist.go` | Daftar field ringkasan dan raw. |
| `projector.go` | Membentuk objek respons baru dari data internal. |

## Kontrak dan alur

- Project(hazard, scopes) menghasilkan objek yang aman untuk diserialisasi.

## Dependensi

- DTO lokal client-api dan scope terverifikasi; dipakai application.

## Aturan penting

- Ringkasan: hazard_id, source, hazard_type, severity, area_name, occurred_at, ingested_at.
- Mentah: source_ref_id, latitude, longitude, attributes.
- Proyeksi berlaku pada list/detail dan setiap request, termasuk cache hit.
- Jangan memodifikasi map/slice bersama yang mungkin berasal dari cache.

## Langkah implementasi dan verifikasi

- Implementasikan allowlist sebelum endpoint bisnis dipublikasikan.
- Verifikasi field baru tidak otomatis terlihat oleh Media.
