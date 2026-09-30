# authn

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Membuktikan identitas client melalui verifikasi JWT dengan public key lokal.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `claims.go` | Claims terverifikasi: sub, scope, iss, aud, exp, jti. |
| `verifier.go` | Verifikasi signature dan claims yang diwajibkan. |

## Kontrak dan alur

- Verify(ctx, rawToken) menghasilkan VerifiedClaims atau error autentikasi.

## Dependensi

- Library JWT yang dipilih tim; public key dari konfigurasi. Dipakai middleware/adapter HTTP.

## Aturan penting

- Hanya EdDSA yang diizinkan; periksa issuer, audience, expiration, dan klaim wajib.
- Tidak memiliki private key atau memanggil auth-service per request.
- Toleransi waktu harus konsisten dengan skenario demo; token kedaluwarsa ditolak setelah toleransi yang dipilih.

## Langkah implementasi dan verifikasi

- Tetapkan format scope dan error terklasifikasi.
- Verifikasi signature salah, alg salah, issuer/audience salah, dan expiration.
