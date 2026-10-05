# middleware

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Proteksi trafik inti dan batas resource jalur baca.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** rate limit token bucket dan semaphore konkurensi sudah diimplementasikan. Penolakan rate dan konkurensi menghasilkan alasan `429` yang berbeda.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `rate_limit.go` | Token bucket per client_id terverifikasi. |
| `concurrency.go` | Semaphore global dengan penolakan terkendali. |
| `timeout.go` | Batas waktu lokal request. |
| `input_limit.go` | Batas header/body dan validasi input umum. |

## Kontrak dan alur

- Middleware membungkus http.Handler; limit dibaca dari config.

## Dependensi

- net/http, authn claims, logger/correlation ID lokal.

## Aturan penting

- Kapasitas penuh segera429 dengan Retry-After, bukan antrean tanpa batas.
- Release semaphore harus terjadi pada semua jalur keluar.
- Batasi pertumbuhan state limiter; jangan menerima client_id arbitrer tanpa verifikasi.
- Catat penolakan terkontrol terpisah dari kegagalan5xx.

## Langkah implementasi dan verifikasi

- Tentukan urutan middleware dengan authn dan limiter pra-autentikasi bila diperlukan.
- Ukur baseline50 koneksi sustained sebelum menambah cache.
