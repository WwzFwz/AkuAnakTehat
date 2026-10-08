# pvmbg-mock

Composition root: validasi konfigurasi, seed, generator, HTTP server, signal dan graceful shutdown.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [main.go](main.go) | Memuat konfigurasi, merangkai dependensi, menjalankan worker dan HTTP, serta menangani shutdown. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
