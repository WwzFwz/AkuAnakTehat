# authn

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Membuktikan identitas client melalui verifikasi JWT dengan public key lokal.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Verifier.Verify(rawToken) menghasilkan Claims terverifikasi atau error autentikasi.

## Dependensi

- Library golang-jwt/jwt/v5; public key dari konfigurasi. Dipakai middleware/adapter HTTP.

## Aturan penting

- Hanya EdDSA yang diizinkan; periksa issuer, audience, expiration, dan klaim wajib.
- Tidak memiliki private key atau memanggil auth-service per request.
- Verifier tidak menambahkan leeway; token yang sudah melewati exp ditolak.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [verifier.go](verifier.go) | Memuat public key dan memverifikasi algoritma serta claim wajib JWT. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
