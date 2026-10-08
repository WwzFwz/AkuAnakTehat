# application

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Use case pertukaran kredensial dan refresh tanpa login ulang pada alur normal.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Issue(ctx, id, secret) dan Refresh(ctx, rawToken) menghasilkan TokenPair.
- Port store menerima data pengganti dan mengubah token lama/new/family secara atomik sesuai kontrak.

## Dependensi

- domain; token dan adapter Redis memenuhi port melalui composition root.

## Aturan penting

- Default access TTL60 s, refresh TTL8 jam; configurable.
- Alur normal refresh tidak meminta login manual.
- Kehilangan respons setelah rotasi dapat memaksa autentikasi ulang; tidak mengklaim atomic HTTP+Redis.
- Client demo tidak me-refresh token yang sama secara paralel.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [service.go](service.go) | Port Store dan Signer, penerbitan pasangan token, refresh, serta pemetaan error grant. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
