# application

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Orkestrasi baca downstream: otorisasi, panggil Aggregator, lalu proyeksi.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- List(ctx, scope, query, fields, raw) dan Get(ctx, scope, id, fields, raw).
- Port Aggregator membaca melalui HTTP; tidak mengekspos driver/database.

## Dependensi

- authz, projection, dan DTO lokal. Adapter outbound mengimplementasikan port.

## Aturan penting

- Tidak meng-import module Aggregator, membaca canonical-db, atau menghubungi mock.
- Tidak melewatkan proyeksi pada cabang detail/error/cache.
- Kegagalan Aggregator menghasilkan ketidaktersediaan yang eksplisit.
- Metadata sumber tetap tersedia tanpa membocorkan pesan error internal.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [service.go](service.go) | Port Aggregator, DTO Page dan Source, otorisasi list/get, proyeksi, dan freshness detail. |

## Perilaku dan batas saat ini

Freshness sources tetap disertakan pada detail dan raw detail, termasuk ketika fields membatasi field hazard. Metadata disaring melalui tipe Source; allowlist Media tidak dilonggarkan.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
