# authz

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Menentukan apakah scope terverifikasi boleh memenuhi permintaan.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Allowed(scope, fields, raw) memeriksa izin; Has memeriksa keanggotaan scope.

## Dependensi

- Scope berasal dari Claims terverifikasi; policy dipanggil application dan handler.

## Aturan penting

- Media yang meminta latitude/longitude/source_ref_id/attributes secara eksplisit ditolak403.
- Scope berasal dari token terverifikasi, bukan role atau flag dari client.
- Default-deny untuk operasi/field baru.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [policy.go](policy.go) | Daftar field ringkasan dan mentah, pemeriksaan scope, serta izin field yang diminta. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
