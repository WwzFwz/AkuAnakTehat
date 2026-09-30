# http

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Transport penerbitan token dan proteksi endpoint sensitif.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Uji integrasi runtime belum dilakukan. Berkas yang sudah ada: `router.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `router.go` | POST /oauth/token serta health/readiness. |
| `handler.go` | Parse grant, validasi body, dan panggil application. |
| `errors.go` | Respons error token tanpa informasi sensitif. |

## Kontrak dan alur

- Grant bentuk client_credentials dan refresh_token dipakai untuk protokol tugas M1, bukan klaim OAuth penuh.

## Dependensi

- application, config, dan observability lokal.

## Aturan penting

- Batasi ukuran body dan request rate.
- Respons token no-store; jangan log body, client_secret, atau refresh token.
- 401/4xx terklasifikasi untuk kredensial invalid; dependency error dibedakan.

## Langkah implementasi dan verifikasi

- Sepakati request/response dengan script Tim Lapangan dan load test.
- Hubungkan flow normal expiry→refresh→request sukses.
