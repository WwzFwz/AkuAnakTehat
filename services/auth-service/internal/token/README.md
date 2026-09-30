# token

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Penandatangan JWT Ed25519 dan pembangkit token acak.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Cakupan verifikasi runtime fondasi tercatat pada [hasil pengujian](../../../../docs/evidence/foundation/README.md); ini belum bukti P1?P5 lengkap. Berkas yang sudah ada: `token.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `jwt.go` | Sign JWT dengan claims yang disepakati. |
| `opaque.go` | Random refresh token dan identifier session/token. |
| `keys.go` | Load/validate private key dari berkas konfigurasi. |

## Kontrak dan alur

- SignAccessToken(claims, ttl) dan NewRefreshToken() menjadi kebutuhan application.

## Dependensi

- crypto/rand, crypto/ed25519, dan library JWT; key dari konfigurasi lokal.

## Aturan penting

- Private key hanya berada pada auth-service.
- JWT memuat iss/aud/sub/scope/exp/jti; algoritma tetap EdDSA.
- Gunakan sumber acak kriptografis, bukan math/rand.
- Tidak menulis secret/token ke log.

## Langkah implementasi dan verifikasi

- Sepakati format PEM dan claims dengan verifier client-api.
- Verifikasi access token hasil signer bisa diverifikasi public key tanpa koneksi auth-service.
