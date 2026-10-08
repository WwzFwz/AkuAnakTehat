# middleware

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Proteksi trafik inti dan batas resource jalur baca.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Handler memanggil Limits.Enter setelah autentikasi; hasilnya release callback atau alasan penolakan.

## Dependensi

- sync dan time; identitas terverifikasi diberikan oleh handler HTTP.

## Aturan penting

- Kapasitas penuh segera429 dengan Retry-After, bukan antrean tanpa batas.
- Release semaphore harus terjadi pada semua jalur keluar.
- Batasi pertumbuhan state limiter; jangan menerima client_id arbitrer tanpa verifikasi.
- Catat penolakan terkontrol terpisah dari kegagalan5xx.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [limits.go](limits.go) | Token bucket per identitas terverifikasi, semaphore global, dan pembatasan state limiter. |
| [limits_test.go](limits_test.go) | Pengujian `TestLimitsDistinguishConcurrencyAndRateRejection`. |

## Perilaku dan batas saat ini

Timeout HTTP server dan ukuran header diatur di cmd/client-api/main.go. Timeout outbound 1,5 detik berada di adapter Aggregator. Package ini tidak menyediakan middleware timeout atau pembatas body generik.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
