# token

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Penandatangan JWT Ed25519 dan pembangkit token acak.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Signer.Sign menerbitkan JWT; Random membangkitkan refresh token dan identifier; Hash menghitung hash token.

## Dependensi

- crypto/rand, crypto/ed25519, dan library JWT; key dari konfigurasi lokal.

## Aturan penting

- Private key hanya berada pada auth-service.
- JWT memuat iss/aud/sub/scope/exp/jti; algoritma tetap EdDSA.
- Gunakan sumber acak kriptografis, bukan math/rand.
- Tidak menulis secret/token ke log.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [token.go](token.go) | Memuat private key Ed25519, menandatangani JWT, dan membangkitkan token acak serta hash. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
