# http

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Transport penerbitan token dan proteksi endpoint sensitif.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Grant bentuk client_credentials dan refresh_token dipakai untuk protokol tugas M1, bukan klaim OAuth penuh.

## Dependensi

- application, config, dan observability lokal.

## Aturan penting

- Batasi ukuran body dan request rate.
- Respons token no-store; jangan log body, client_secret, atau refresh token.
- 401/4xx terklasifikasi untuk kredensial invalid; dependency error dibedakan.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [router.go](router.go) | Mendaftarkan endpoint, memvalidasi request, dan membentuk respons HTTP. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
