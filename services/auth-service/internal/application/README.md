# application

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Use case pertukaran kredensial dan refresh tanpa login ulang pada alur normal.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Cakupan verifikasi runtime fondasi tercatat pada [hasil pengujian](../../../../docs/evidence/foundation/README.md); ini belum bukti P1?P5 lengkap. Berkas yang sudah ada: `service.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `issue_token.go` | Validasi identitas lalu terbitkan pasangan token. |
| `refresh_token.go` | Validasi/rotasi refresh token dan penanganan reuse. |
| `ports.go` | Port signer, pembangkit opaque token, dan store rotasi atomik. |

## Kontrak dan alur

- Issue(ctx, credentials) dan Refresh(ctx, refreshToken).
- Port store menerima data pengganti dan mengubah token lama/new/family secara atomik sesuai kontrak.

## Dependensi

- domain; token dan adapter Redis memenuhi port melalui composition root.

## Aturan penting

- Default access TTL60 s, refresh TTL8 jam; configurable.
- Alur normal refresh tidak meminta login manual.
- Kehilangan respons setelah rotasi dapat memaksa autentikasi ulang; tidak mengklaim atomic HTTP+Redis.
- Client demo tidak me-refresh token yang sama secara paralel.

## Langkah implementasi dan verifikasi

- Tetapkan error invalid credentials, expired token, reuse, dan store unavailable.
- Verifikasi scope tidak naik lewat refresh dan reuse mencabut family.
